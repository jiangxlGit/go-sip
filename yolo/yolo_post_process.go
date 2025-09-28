package yolo

import (
	. "go-sip/logger"

	"github.com/swdee/go-rknnlite/postprocess"
	"go.uber.org/zap"
)

// 创建yolo后处理器

// YOLOv5OneClassParams 返回一个配置了单类别（比如只检测「人」）的 YOLOv5Params 参数。
// 此配置适用于：
// - 单类别训练好的 YOLOv5 模型
// - 对应 RKNN 转换后的输出通道为 18（即 3 * (5 + 1)）
// 注意：需保证训练、导出（onnx）、转换（rknn）、后处理的类别数完全一致，否则解析会崩溃或错误。
func YOLOv8OneClassParams() postprocess.YOLOv8Params {
	return postprocess.YOLOv8Params{
		// YOLOv8 是 anchor-free，不再需要 strides 和 anchors

		// BoxThreshold: 分数阈值 (obj_conf * class_conf)
		BoxThreshold: 0.25,

		// NMSThreshold: IoU 抑制阈值
		NMSThreshold: 0.3,

		// ObjectClassNum: 单类别
		ObjectClassNum: 1,

		// MaxObjectNumber: 限制最大保留框数量
		MaxObjectNumber: 5,
	}
}

func YoloPostProcess() *postprocess.YOLOv8 {
	yoloPostProcesser := postprocess.NewYOLOv8(YOLOv8OneClassParams())
	Logger.Info("创建YOLOv8后处理器", zap.Any("参数: ", yoloPostProcesser.Params))
	return yoloPostProcesser
}
