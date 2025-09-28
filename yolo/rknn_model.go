package yolo

import (
	. "go-sip/logger"

	"github.com/swdee/go-rknnlite"
	"go.uber.org/zap"
)

// RknnLoadModel 加载模型文件
func RknnLoadModel(rkModelPath, rkPlatform string) (*rknnlite.Runtime, error) {
	Logger.Info("加载模型文件", zap.Any("rkModelPath", rkModelPath), zap.Any("platform", rkPlatform))

	// 将特定的进程或线程绑定到一个或多个CPU核心上，可以避免因为核心切换而导致的缓存失效，从而提升执行效率
	// 通过CPU亲和度可以将它们分配到不同的CPU核心上，减少竞争和冲突。
	err := rknnlite.SetCPUAffinityByPlatform(rkPlatform, rknnlite.FastCores)
	if err != nil {
		Logger.Error("设置CPU亲和度失败", zap.Any("rkPlatform", rkPlatform), zap.Any("error", err))
		return nil, err
	}

	// 创建 RKNN 运行时实例
	rt, err := rknnlite.NewRuntimeByPlatform(rkPlatform, rkModelPath)
	if err != nil {
		Logger.Error("初始化 RKNN 运行时出错: ", zap.Any("error", err))
		return nil, err
	}

	// 设置运行时保留输出张量为int8格式
	rt.SetWantFloat(false)

	return rt, nil

}
