package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/xiaoyin001/game-kj/internal/core/log"
	"github.com/xiaoyin001/game-kj/internal/core/module"
	_ "github.com/xiaoyin001/game-kj/internal/core/util"
)

var (
	logLv  = flag.String("log", "debug", "日志级别")
	logDir = flag.String("logdir", "./logs", "日志目录")
	env    = flag.String("env", "dev", "环境")
	debug  = flag.Bool("debug", true, "是否输出到控制台")
)

func main() {
	flag.Parse()
	fmt.Println("====================================================")
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("Flag: -%s=%v (默认值: %s)\n", f.Name, f.Value.String(), f.DefValue)
	})
	fmt.Println("====================================================")

	// 日志需要最先初始化
	initLog()
	defer log.Close()

	log.Info("this is game server")
	log.Info("this is game server", log.String("name", "xiaoyin"))
	log.Infof("this is game server %s xiaoyin", "1111111111111")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO 启动网络

	// 模块运行
	moduleMgr := module.CreateModuleMgr()
	moduleMgr.Init()
	moduleMgr.Start()
	defer moduleMgr.Stop()

	// 使用signal.Notify监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// 等待系统信号或ctx取消
	select {
	case <-ctx.Done():
		log.Info("收到ctx取消信号")
	case sig := <-sigChan:
		log.Info("收到系统信号", log.String("signal", sig.String()))
		cancel() // 取消ctx
	}
}

func initLog() {
	err := log.InitLogger(log.Options{
		Level:       *logLv,
		LogDir:      *logDir,
		Console:     *debug,
		Development: *env == "dev",
	})
	if err != nil {
		panic(err)
	}
}
