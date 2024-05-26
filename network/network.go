package network

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"net"
	"os"
	"os/signal"
	"syscall"
	"tinydocker/config"
	"tinydocker/log"
)

var (
	myChain = "TINY_DOCKER"
)

type NetConf struct {
	NetworkName string
	IpRange     *net.IPNet
	Driver      string
	BridgeName  string
	BridgeIp    *net.IPNet
}

type netMgr struct {
	Storage map[string]*NetConf
}

var NetMgr = &netMgr{
	Storage: map[string]*NetConf{},
}

func (n *netMgr) Sync() error {
	if _, err := os.Stat(config.NetStoragePath); err != nil {
		if os.IsNotExist(err) {
			os.Create(config.NetStoragePath)
		} else {
			return err
		}
	}
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	err = os.WriteFile(config.NetStoragePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (n *netMgr) LoadConf() error {
	if _, err := os.Stat(config.NetStoragePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	data, err := os.ReadFile(config.NetStoragePath)
	if err != nil {
		return err
	}
	if len(n.Storage) == 0 {
		n.Storage = make(map[string]*NetConf)
	}
	if len(data) == 0 {
		return nil
	}
	err = json.Unmarshal(data, n)
	if err != nil {
		return err
	}
	return nil
}

const (
	defaultNetName = "testbridge"
)

type networktype string

const (
	BridgeNetworkType networktype = "bridge"
	DefaultSubnet                 = "192.169.0.1/24"
)

func (n networktype) String() string {
	return string(n)
}

func Init() error {
	// 对默认网络进行初始化
	if err := BridgeDriver.CreateNetwork(defaultNetName, DefaultSubnet, BridgeNetworkType); err != nil {
		return fmt.Errorf("err=%s", err)
	}
	if err := IpAmfs.SetIpUsed(DefaultSubnet); err != nil {
		return err
	}
	return nil
}

func ConfigDefaultNetworkInNewNet(pid int) (error, net.IP) {
	// 获取ip
	ip, err := IpAmfs.AllocIp(DefaultSubnet)
	if err != nil {
		return fmt.Errorf("ipam alloc ip fail %s", err), nil
	}

	// 主机上创建 veth 设备,并连接到网桥上
	vethLink, networkConf, err := BridgeDriver.CrateVeth(defaultNetName)
	if err != nil {
		return fmt.Errorf("create veth fail err=%s", err), nil
	}
	// 主机上设置子进程网络命名空间 配置
	if err := BridgeDriver.setContainerIp(vethLink.PeerName, pid, ip, networkConf.BridgeIp); err != nil {
		return fmt.Errorf("setContainerIp fail err=%s peername=%s pid=%d ip=%v conf=%+v", err, vethLink.PeerName, pid, ip, networkConf), nil
	}

	// 通知子进程设置完毕
	log.Debug("parent process set ip success")
	log.Info("ip set success ip:ip %s", ip.String())
	return noticeSunProcessNetConfigFin(pid), ip
}

func noticeSunProcessNetConfigFin(pid int) error {
	return syscall.Kill(pid, syscall.SIGUSR2)
}

func WaitParentSetNewNet() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR2)
	<-sigs
	log.Info("Received SIGUSR2 signal, prepare run container")
}

func MappingPort(ip, dPort, tDPort string) {
	tables, err := iptables.New()
	if err != nil {
		log.Error("Failed to create iptables rules")
	}
	exists, err := tables.ChainExists("nat", myChain)
	if err != nil {
		log.Error("Failed to check iptables rules")
	}
	if !exists {
		err = tables.NewChain("nat", myChain)
	}

	err = tables.InsertUnique("nat", myChain, 1, "-p", "tcp", "--dport", dPort, "-j", "DNAT", "--to-destination", ip+":"+tDPort)

	if err != nil {
		log.Error("Failed to insert iptables rules")
	}
	err = tables.Append("nat", "PREROUTING", "-j", myChain)
	if err != nil {
		log.Error("Failed to append iptables rules")
	}
	err = tables.InsertUnique("nat", "POSTROUTING", 1, "-p", "tcp", "--source", ip, "--destination", ip, "--dport", tDPort, "-j", "MASQUERADE")

	if err != nil {
		log.Error("Failed to map port")
	}
	log.Info("mapping port success: %s -> %s", dPort, tDPort)
}
