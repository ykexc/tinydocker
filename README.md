# TinyDocker


### 配置环境

- 远程服务器上安装go环境 & 配置goland
    - 在设置go那里的build tag选择下环境
    - run 那里配置下remote

### 实现namespace隔离

#### namespace实现的隔离

- 主机名
- pid
- mount
- 网络
- 进程通信

##### chroot和 pivot_root

chroot可以改变进程的根目录，但不能改变mount(挂载)命名空间的根目录

##### 使用mount挂载proc
mount --make-rprivate #systemd 为init进程时，挂载默认是共享模式挂载的，共享模式挂载会让所有命名空间都能看到各自的挂载的目录

mount -t proc proc /proc

### 实现容器网络

- 创建网桥
- 创建虚拟网卡
- 连接网桥和虚拟网卡
- 基于iptables实现端口映射
- 基于linux内核的网络转发实现共享上网

```shell
#iptables常见命令

#查看某个表下的所有链的规则
iptables -nvL [-t] [table]
#查看所有硬保存的规则
iptables-save
#附加
-A
#删除
-D
#创建自定义链
-N
#删除自定义链
-X
#删除所有规则
-F
#四表五链指的是：(mangle, raw, filter, nat) (INPUT, FORWARD, OUTPUT, PREROUTING, POSTROUTING)
#DNat用于端口转发，SNat用于共享上网

```


