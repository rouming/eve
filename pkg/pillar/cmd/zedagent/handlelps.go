// Copyright (c) 2021 Zededa, Inc.
// SPDX-License-Identifier: Apache-2.0

package zedagent

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	zconfig "github.com/lf-edge/eve/api/go/config"
	"github.com/lf-edge/eve/api/go/profile"
	"github.com/lf-edge/eve/pkg/pillar/flextimer"
	"github.com/lf-edge/eve/pkg/pillar/types"
	"github.com/lf-edge/eve/pkg/pillar/zedcloud"
	"google.golang.org/protobuf/proto"
)

const (
	defaultLPSPort = "8888"
	profileURLPath = "/api/v1/local_profile"
	savedLPSFile   = "lastlocalprofile"
)

// makeLPSBaseURL constructs local server URL without path.
func makeLPSBaseURL(lpsAddr string) (string, error) {
	lpsURL := fmt.Sprintf("http://%s", lpsAddr)
	u, err := url.Parse(lpsURL)
	if err != nil {
		return "", fmt.Errorf("url.Parse: %s", err)
	}
	if u.Port() == "" {
		lpsURL = fmt.Sprintf("%s:%s", lpsURL, defaultLPSPort)
	}
	return lpsURL, nil
}

// Run a periodic fetch of the currentProfile from lps
func lpsTimerTask(handleChannel chan interface{}, getconfigCtx *getconfigContext) {

	ctx := getconfigCtx.zedagentCtx

	// use ConfigInterval as lpsInterval
	lpsInterval := ctx.globalConfig.GlobalValueInt(types.ConfigInterval)
	interval := time.Duration(lpsInterval) * time.Second
	max := float64(interval)
	min := max * 0.3
	ticker := flextimer.NewRangeTicker(time.Duration(min),
		time.Duration(max))
	// Return handle to caller
	handleChannel <- ticker

	log.Functionf("lpsTimerTask: waiting for lpsTrigger")
	//wait for the first trigger comes from parseProfile to have information about lps
	<-getconfigCtx.lpsTrigger
	log.Functionf("lpsTimerTask: waiting for lpsTrigger done")
	//trigger again to pass into loop
	triggerGetLPS(getconfigCtx)

	wdName := agentName + "currentProfile"

	// Run a periodic timer so we always update StillRunning
	stillRunning := time.NewTicker(25 * time.Second)
	ctx.ps.StillRunning(wdName, warningTime, errorTime)
	ctx.ps.RegisterFileWatchdog(wdName)

	for {
		select {
		case <-getconfigCtx.lpsTrigger:
			start := time.Now()
			profileStateMachine(getconfigCtx, false)
			ctx.ps.CheckMaxTimeTopic(wdName, "getLPSConfigTrigger", start,
				warningTime, errorTime)
		case <-ticker.C:
			start := time.Now()
			profileStateMachine(getconfigCtx, false)
			ctx.ps.CheckMaxTimeTopic(wdName, "getLPSConfigTimer", start,
				warningTime, errorTime)
		case <-stillRunning.C:
		}
		ctx.ps.StillRunning(wdName, warningTime, errorTime)
	}
}

func parseLPS(lpsBytes []byte) (*profile.LocalProfile, error) {
	var lps = &profile.LocalProfile{}
	err := proto.Unmarshal(lpsBytes, lps)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling failed: %v", err)
	}
	return lps, nil
}

// read saved local profile in case of particular reboot reason
func readSavedLPS(getconfigCtx *getconfigContext) (*profile.LocalProfile, error) {
	lpsMessage, ts, err := readSavedConfig(
		getconfigCtx.zedagentCtx.globalConfig.GlobalValueInt(types.StaleConfigTime),
		filepath.Join(checkpointDirname, savedLPSFile), false)
	if err != nil {
		return nil, fmt.Errorf("readSavedLPS: %v", err)
	}
	if lpsMessage != nil {
		log.Noticef("Using saved local profile dated %s",
			ts.Format(time.RFC3339Nano))
		return parseLPS(lpsMessage)
	}
	return nil, nil
}

// getLPSConfig connects to local profile server to fetch the current profile
func getLPSConfig(getconfigCtx *getconfigContext, lpsURL string) (*profile.LocalProfile, error) {

	log.Functionf("getLPSConfig(%s)", lpsURL)

	if !getconfigCtx.lpsMap.upToDate {
		err := updateLPSMap(getconfigCtx, lpsURL)
		if err != nil {
			return nil, fmt.Errorf("getLPSConfig: updateLPSMap: %v", err)
		}
		// Make sure HasLPS is set correctly for the AppInstanceConfig
		updateHasLPS(getconfigCtx)
	}

	srvMap := getconfigCtx.lpsMap.servers
	if len(srvMap) == 0 {
		return nil, fmt.Errorf(
			"getLPSConfig: cannot find any configured apps for lpsURL: %s",
			lpsURL)
	}

	var errList []string
	for bridgeName, servers := range srvMap {
		for _, srv := range servers {
			fullURL := srv.lpsAddr + profileURLPath
			lps := &profile.LocalProfile{}
			resp, err := zedcloud.SendLocalProto(
				zedcloudCtx, fullURL, bridgeName, srv.bridgeIP, nil, lps)
			if err != nil {
				errList = append(errList, fmt.Sprintf("SendLocal: %s", err))
				continue
			}
			if resp.StatusCode != http.StatusOK {
				errList = append(errList, fmt.Sprintf("SendLocalProto: wrong response status code: %d",
					resp.StatusCode))
				continue
			}
			if lps.GetServerToken() != getconfigCtx.profileServerToken {
				errList = append(errList,
					fmt.Sprintf("invalid token submitted by local server (%s)", lps.GetServerToken()))
				continue
			}
			return lps, nil
		}
	}
	return nil, fmt.Errorf("getLPSConfig: all attempts failed: %s", strings.Join(errList, ";"))
}

// saveOrTouchReceivedLPS updates modification time of received LPS in case of no changes
// or updates content of received LPS in case of changes or no checkpoint file
func saveOrTouchReceivedLPS(getconfigCtx *getconfigContext, lps *profile.LocalProfile) {
	if getconfigCtx.lps == lps.GetLocalProfile() &&
		getconfigCtx.profileServerToken == lps.GetServerToken() &&
		existsSavedConfig(savedLPSFile) {
		touchSavedConfig(savedLPSFile)
		return
	}
	contents, err := proto.Marshal(lps)
	if err != nil {
		log.Errorf("saveOrTouchReceivedLPS Marshalling failed: %s", err)
		return
	}
	saveConfig(savedLPSFile, contents)
	return
}

// parseProfile process local and global profile configuration
// must be called before processing of app instances from config
func parseProfile(ctx *getconfigContext, config *zconfig.EdgeDevConfig) {
	log.Functionf("parseProfile start: globalProfile: %s lps: %s",
		ctx.globalProfile, ctx.lps)
	if ctx.globalProfile != config.GlobalProfile {
		log.Noticef("parseProfile: GlobalProfile changed from %s to %s",
			ctx.globalProfile, config.GlobalProfile)
		ctx.globalProfile = config.GlobalProfile
	}
	ctx.profileServerToken = config.ProfileServerToken
	if ctx.lps != config.LocalProfileServer {
		log.Noticef("parseProfile: LPS changed from %s to %s",
			ctx.lps, config.LocalProfileServer)
		ctx.lps = config.LocalProfileServer
		triggerGetLPS(ctx)
		triggerRadioPOST(ctx)
		updateLocalAppInfoTicker(ctx, false)
		triggerLocalAppInfoPOST(ctx)
		updateLocalDevInfoTicker(ctx, false)
		triggerLocalDevInfoPOST(ctx)
		ctx.lpsThrottledLocation = false
	}
	profileStateMachine(ctx, true)
	log.Functionf("parseProfile done globalProfile: %s currentProfile: %s",
		ctx.globalProfile, ctx.currentProfile)
}

// determineCurrentProfile return current profile based on lps, globalProfile
func determineCurrentProfile(ctx *getconfigContext) string {
	if ctx.lps == "" {
		return ctx.globalProfile
	}
	return ctx.lps
}

// triggerGetLPS notifies task to reload local profile from profileServer
func triggerGetLPS(ctx *getconfigContext) {
	log.Functionf("triggerGetLPS")
	select {
	case ctx.lpsTrigger <- Notify{}:
	default:
	}
}

// run state machine to handle changes to globalProfile, lps,
// or to do periodic fetch of the local profile
// If skipFetch is set we do not look for an update from a lps
// but keep the current lps
func profileStateMachine(ctx *getconfigContext, skipFetch bool) {
	lps := getLPS(ctx, skipFetch)
	if ctx.lps != lps {
		log.Noticef("local profile changed from %s to %s",
			ctx.lps, lps)
		ctx.lps = lps
	}
	currentProfile := determineCurrentProfile(ctx)
	if ctx.currentProfile != currentProfile {
		log.Noticef("current profile changed from %s to %s",
			ctx.currentProfile, currentProfile)
		ctx.currentProfile = currentProfile
		publishZedAgentStatus(ctx)
	}
}

// getLPS returns the local profile to use, and cleans up ctx and
// checkpoint when the local profile server has been removed. If skipCheck
// is not set it will query the local profile server.
// It returns the last known value until it gets a response from the server
// or lps is cleared.
func getLPS(ctx *getconfigContext, skipFetch bool) string {
	lps := ctx.lps
	if lps == "" {
		if ctx.lps != "" {
			log.Noticef("clearing lps checkpoint since no server")
			cleanSavedConfig(savedLPSFile)
		}
		return ""
	}
	if skipFetch {
		return ctx.lps
	}
	lpsURL, err := makeLPSBaseURL(lps)
	if err != nil {
		log.Errorf("getLPS: makeLPSBaseURL: %s", err)
		return ""
	}
	lpsConfig, err := getLPSConfig(ctx, lpsURL)
	if err != nil {
		log.Errorf("getLPS: getLPSConfig: %s", err)
		// Return last known value
		return ctx.lps
	}
	localProfile := lpsConfig.GetLocalProfile()
	saveOrTouchReceivedLPS(ctx, lpsConfig)
	return localProfile
}

// processSavedProfile reads saved local profile and set it
func processSavedProfile(ctx *getconfigContext) {
	lps, err := readSavedLPS(ctx)
	if err != nil {
		log.Functionf("processSavedProfile: readSavedLPS %s", err)
		return
	}
	if lps != nil {
		log.Noticef("starting with lps %s", lps.LocalProfile)
		ctx.lps = lps.LocalProfile
	}
}
