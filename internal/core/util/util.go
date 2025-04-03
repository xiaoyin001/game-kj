// 这个包只在 main.go中进行import
// 其他地方不能import

package util

// 注册的模块都需要在这里进行import
import (
	_ "github.com/xiaoyin001/game-kj/internal/modules/demo"
)
