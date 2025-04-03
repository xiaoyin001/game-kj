package demo

import (
	"github.com/xiaoyin001/game-kj/internal/core/log"
	"github.com/xiaoyin001/game-kj/internal/core/module"
)

func init() {
	module.RegisterModule(newModelDemo())
}

// 模块需要实现 module.Module 接口
var _ module.Module = (*ModelDemo)(nil)

func newModelDemo() *ModelDemo {
	return &ModelDemo{}
}

// 模块模型
type ModelDemo struct {
	// 可定义模块自己的模型
}

func (m *ModelDemo) Name() string {
	return "demo"
}

func (m *ModelDemo) Init() error {
	log.Info(m.Name() + " 模块初始化")
	return nil
}

func (m *ModelDemo) Start() error {
	log.Info(m.Name() + " 模块启动")
	return nil
}

func (m *ModelDemo) Stop() error {
	log.Info(m.Name() + " 模块停止")
	return nil
}

func (m *ModelDemo) LoadCfg(isReload bool) error {
	log.Info(m.Name() + " 模块加载配置")
	return nil
}
