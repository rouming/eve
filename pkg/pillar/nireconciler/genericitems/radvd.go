// Copyright (c) 2023 Zededa, Inc.
// SPDX-License-Identifier: Apache-2.0

package genericitems

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	dg "github.com/lf-edge/eve/libs/depgraph"
	"github.com/lf-edge/eve/libs/reconciler"
	"github.com/lf-edge/eve/pkg/pillar/base"
	uuid "github.com/satori/go.uuid"
)

// Radvd : router advertisement daemon (https://linux.die.net/man/5/radvd.conf).
type Radvd struct {
	// ForNI : UUID of the Network Instance for which this radvd instance is created.
	// Mostly used just to force re-start of radvd when one NI is being deleted
	// and subsequently another is created for the same bridge interface name
	// (ForNI will differ in such case).
	ForNI uuid.UUID
	// ListenIf : interface on which radvd should listen.
	ListenIf NetworkIf
}

// Name returns the interface name on which radvd listens.
// This ensures that there cannot be two different radvd instances
// that would attempt to listen on the same interface at the same time.
func (r Radvd) Name() string {
	return r.ListenIf.IfName
}

// Label for the radvd instance.
func (r Radvd) Label() string {
	return "radvd for " + r.ListenIf.IfName
}

// Type of the item.
func (r Radvd) Type() string {
	return RadvdTypename
}

// Equal compares two Radvd instances
func (r Radvd) Equal(other dg.Item) bool {
	r2 := other.(Radvd)
	return r.ForNI == r2.ForNI &&
		r.ListenIf == r2.ListenIf
}

// External returns false.
func (r Radvd) External() bool {
	return false
}

// String describes the radvd instance.
func (r Radvd) String() string {
	return fmt.Sprintf("Radvd: {NI: %s, listenIf: %s}",
		r.ForNI, r.ListenIf.IfName)
}

// Dependencies returns returns the interface on which radvd listens
// as the only dependency.
func (r Radvd) Dependencies() (deps []dg.Dependency) {
	deps = append(deps, dg.Dependency{
		RequiredItem: r.ListenIf.ItemRef,
		Description:  "interface on which radvd listens must exist",
	})
	return deps
}

const radvdConfigTemplate = `
# Automatically generated by zedrouter
# Low preference to allow underlay to have high preference default
interface %s {
	IgnoreIfMissing on;
	AdvSendAdvert on;
	MaxRtrAdvInterval 1800;
	AdvManagedFlag on;
	AdvLinkMTU 1280;
	AdvDefaultPreference low;
	route fd00::/8
	{
		AdvRoutePreference high;
		AdvRouteLifetime 1800;
	};
};
`

const (
	radvdStartTimeout = 3 * time.Second
	radvdStopTimeout  = 10 * time.Second
)

// RadvdConfigurator implements Configurator interface (libs/reconciler) for radvd.
type RadvdConfigurator struct {
	Log *base.LogObject
}

// Create starts radvd.
func (c *RadvdConfigurator) Create(ctx context.Context, item dg.Item) error {
	radvd, isRadvd := item.(Radvd)
	if !isRadvd {
		return fmt.Errorf("invalid item type %T, expected Radvd", item)
	}
	if err := c.createRadvdConfigFile(radvd); err != nil {
		return err
	}
	done := reconciler.ContinueInBackground(ctx)
	go func() {
		err := c.startRadvd(ctx, radvd.Name())
		done(err)
	}()
	return nil
}

// Modify is not implemented.
func (c *RadvdConfigurator) Modify(ctx context.Context, oldItem, newItem dg.Item) (err error) {
	return errors.New("not implemented")
}

// Delete stops radvd.
func (c *RadvdConfigurator) Delete(ctx context.Context, item dg.Item) error {
	radvd, isRadvd := item.(Radvd)
	if !isRadvd {
		return fmt.Errorf("invalid item type %T, expected Radvd", item)
	}
	done := reconciler.ContinueInBackground(ctx)
	go func() {
		err := c.stopRadvd(ctx, radvd.Name())
		if err == nil {
			// Ignore errors from here.
			_ = c.removeRadvdConfigFile(radvd.Name())
			_ = c.removeRadvdPidFile(radvd.Name())
		}
		done(err)
	}()
	return nil
}

// NeedsRecreate always returns true - Modify is not implemented.
func (c *RadvdConfigurator) NeedsRecreate(oldItem, newItem dg.Item) (recreate bool) {
	return true
}

func (c *RadvdConfigurator) radvdConfigPath(instanceName string) string {
	return filepath.Join(zedrouterRunDir, "radvd."+instanceName+".conf")
}

func (c *RadvdConfigurator) radvdPidFile(instanceName string) string {
	return filepath.Join(zedrouterRunDir, "radvd."+instanceName+".pid")
}

func (c *RadvdConfigurator) createRadvdConfigFile(radvd Radvd) error {
	cfgPath := c.radvdConfigPath(radvd.Name())
	file, err := os.Create(cfgPath)
	if err != nil {
		err = fmt.Errorf("failed to create radvd config file %s: %w", cfgPath, err)
		c.Log.Error(err)
		return err
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf(radvdConfigTemplate, radvd.ListenIf.IfName))
	if err != nil {
		err = fmt.Errorf("failed to write radvd config to file %s: %w", cfgPath, err)
		c.Log.Error(err)
		return err
	}
	return nil
}

// Start radvd as a daemon process.
func (c *RadvdConfigurator) startRadvd(ctx context.Context, instanceName string) error {
	cmd := "nohup"
	pidFile := c.radvdPidFile(instanceName)
	args := []string{
		"radvd",
		"-u", "radvd",
		"-C", c.radvdConfigPath(instanceName),
		"-p", pidFile,
	}
	return startProcess(ctx, c.Log, cmd, args, pidFile, radvdStartTimeout, true)
}

func (c *RadvdConfigurator) stopRadvd(ctx context.Context, instanceName string) error {
	pidFile := c.radvdPidFile(instanceName)
	return stopProcess(ctx, c.Log, pidFile, radvdStopTimeout)
}

func (c *RadvdConfigurator) removeRadvdConfigFile(instanceName string) error {
	cfgPath := c.radvdConfigPath(instanceName)
	if err := os.Remove(cfgPath); err != nil {
		err = fmt.Errorf("failed to remove radvd config %s: %w", cfgPath, err)
		c.Log.Error(err)
		return err
	}
	return nil
}

func (c *RadvdConfigurator) removeRadvdPidFile(instanceName string) error {
	pidPath := c.radvdPidFile(instanceName)
	if err := os.Remove(pidPath); err != nil {
		err = fmt.Errorf("failed to remove radvd PID file %s: %w", pidPath, err)
		c.Log.Error(err)
		return err
	}
	return nil
}