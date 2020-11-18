module github.com/lf-edge/eve/pkg/pillar

go 1.13

require (
	github.com/VictorLowther/godmi v0.0.0-20190311134151-270258a8252d // indirect
	github.com/containerd/cgroups v0.0.0-20200824123100-0b889c03f102
	github.com/containerd/containerd v1.4.1-0.20201014210714-22aea1e9a7a0
	github.com/containerd/fifo v0.0.0-20201026212402-0724c46b320c // indirect
	github.com/containerd/typeurl v1.0.1
	github.com/cshari-zededa/eve-tpm2-tools v0.0.4
	github.com/digitalocean/go-libvirt v0.0.0-20190715144809-7b622097a793 // indirect
	github.com/digitalocean/go-qemu v0.0.0-20181112162955-dd7bb9c771b8
	github.com/docker/docker v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/eriknordmark/ipinfo v0.0.0-20190220084921-7ee0839158f9
	github.com/eriknordmark/netlink v0.0.0-20190912172510-3b6b45309321
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.0
	github.com/google/go-containerregistry v0.1.4
	github.com/google/go-tpm v0.3.0
	github.com/google/gopacket v1.1.17
	github.com/gorilla/websocket v1.4.0
	github.com/jackwakefield/gopac v1.0.2
	github.com/kr/fs v0.1.0 // indirect
	github.com/lf-edge/edge-containers v0.0.0-20201111200732-5491ea93dbe4
	github.com/lf-edge/eve/api/go v0.0.0-00010101000000-000000000000
	github.com/lf-edge/eve/libs/zedUpload v0.0.0-20201208021720-782660b12f0b
	github.com/ochapman/godmi v0.0.0-20140902235245-2527e2081a16 // indirect
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/packetcap/go-pcap v0.0.0-20200802095634-4c3b9511add7
	github.com/rackn/gohai v0.0.0-20190321191141-5053e7f1fa36
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/shirou/gopsutil v0.0.0-20190323131628-2cbc9195c892
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/tatsushid/go-fastping v0.0.0-20160109021039-d7bb493dee3e
	github.com/vishvananda/netlink v1.0.1-0.20190823182904-a1c9a648f744 // indirect
	github.com/vishvananda/netns v0.0.0-20190625233234-7109fa855b0f // indirect
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392
	golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f
	google.golang.org/grpc v1.33.0
	gopkg.in/mcuadros/go-syslog.v2 v2.3.0
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06 // indirect
)

replace github.com/lf-edge/eve/api/go => ../../api/go

replace github.com/vishvananda/netlink/nl => github.com/eriknordmark/netlink/nl v0.0.0-20190903203740-41fa442996b8

replace github.com/vishvananda/netlink => github.com/eriknordmark/netlink v0.0.0-20190903203740-41fa442996b8

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0

// because containerd
replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20190205005809-0d3efadf0154
