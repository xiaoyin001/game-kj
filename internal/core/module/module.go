// module.go - 模块管理系统

// 模块注册，管理创建初始化启动停止都是在主线程中进行

package module

import (
	"github.com/xiaoyin001/game-kj/internal/core/log"
)

var (
	moduleInstance []*moduleInfo = make([]*moduleInfo, 0)
)

// 注册模块
func RegisterModule(module Module) {
	moduleInfo := &moduleInfo{
		Module: module,
		state:  ModuleStateRegistered,
	}

	moduleInstance = append(moduleInstance, moduleInfo)
}

type Module interface {
	// 模块名称
	Name() string
	// 模块初始化
	Init() error
	// 模块启动
	Start() error
	// 模块停止
	Stop() error
	// 加载模块配置
	LoadCfg(isReload bool) error
}

// 模块状态
type ModuleState uint8

const (
	// 已经注册
	ModuleStateRegistered ModuleState = 0
	// 已经初始化
	ModuleStateInitialized ModuleState = 1
	// 配置加载完毕
	ModuleStateCfgLoaded ModuleState = 2
	// 已经启动
	ModuleStateStarted ModuleState = 3
	// 已经停止
	ModuleStateStopped ModuleState = 4
)

// 模块信息
type moduleInfo struct {
	Module // 模块实例

	state      ModuleState // 模块状态
	isDisabled bool        // 是否禁用该模块,true表示禁用不会启动,false表示正常启动
	weight     int         // 启动顺序（权重，数字越小越先启动）
}

func CreateModuleMgr() *moduleMgr {
	moduleMgr := &moduleMgr{
		modules: make(map[string]*moduleInfo),
	}

	for _, module := range moduleInstance {
		log.Info("注册模块", log.String("moduleName", module.Name()))

		moduleMgr.modules[module.Name()] = module
	}

	return moduleMgr
}

type moduleMgr struct {
	modules map[string]*moduleInfo
}

func (m *moduleMgr) Init() {
	for _, module := range moduleInstance {
		module.Init()
	}

	for _, module := range moduleInstance {
		// 加载模块的固定配置，有判断这个模块是否进行启动，启动的顺序

		// 如果是需要启动的，进行加载其余配置【下面的配置可能都需要考虑要按照模块的先后顺序进行加载】
		module.LoadCfg(false)
	}
}

func (m *moduleMgr) Start() {
	for _, module := range m.modules {
		if module.isDisabled {
			continue
		}

		module.Start()
	}
}

func (m *moduleMgr) Stop() {
	for _, module := range m.modules {
		module.Stop()
	}
}
