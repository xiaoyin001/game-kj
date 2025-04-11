
# 安装指定go版本
GO_VERSION=1.22.12
install-go:
	@wget https://studygolang.com/dl/golang/go${GO_VERSION}.linux-amd64.tar.gz
	@tar xf go${GO_VERSION}.linux-amd64.tar.gz
	@sudo rm -rf /usr/local/go
	@sudo mv go /usr/local
	@rm -rf go${GO_VERSION}.linux-amd64.tar.gz
	@go version
