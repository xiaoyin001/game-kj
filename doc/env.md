# 环境搭建

1. 环境初始配置【手动配置在 ~/.bashrc】
``` shell
# 根据自己安装go语言包的位置来配置
export GOROOT=/usr/local/go
export GOBIN=/usr/local/go/bin
# 根据自己的需要配置go的相关包的存放路径
export GOPATH=/home/xiaoyin01/AAA_Codes/go-path

# 方便全局使用
export PATH=$PATH:$GOROOT:$GOPATH:$GOBIN
```

2. go语言包
``` shell
# 当前安装的go版本【linux-amd64】
VER=1.22.12
wget https://go.dev/dl/go${VER}.linux-amd64.tar.gz && tar xf go${VER}.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo mv go /usr/local && rm -rf go${VER}.linux-amd64.tar.gz
```


