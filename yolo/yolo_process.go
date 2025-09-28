package yolo

import (
	. "go-sip/logger"
	"sync"

	"go.uber.org/zap"

	"context"
)

type YoloProcessorStruct struct {
	DeviceId          string
	RkPlatform        string
	StreamId          string
	VedioStreamUrl    string
	YoloModelInfoList []*YoloModelInfo
	StreamProcessor   *StreamProcessor
}

type YoloModelInfo struct {
	RkModelName    string
	RkModelPath    string
	RkPlatform     string
	ClassNames     []string
	ClassNameScore map[string]float64 // 存储类别名称和对应的分数
	Confidence     float32            // 模型置信度
}

var AiModelTriggerChanMap sync.Map

func (yp *YoloProcessorStruct) StartYoloProcessor() error {
	ctx, cancel := context.WithCancel(context.Background())

	rts := make([]*YoloRuntime, 0)

	// 加载模型文件
	for idx, model := range yp.YoloModelInfoList {
		_yr := &YoloRuntime{
			StreamId:       yp.StreamId,
			RkModelName:    model.RkModelName,
			RkModelPath:    model.RkModelPath,
			RkPlatform:     model.RkPlatform,
			ClassNames:     model.ClassNames,
			ClassNameScore: make(map[string]float64),
			Confidence:     model.Confidence,
			frame_interval: idx + 5,
			RtClosed:       make(chan string, 1),
		}
		rm, err := RknnLoadModel(model.RkModelPath, model.RkPlatform)
		if err != nil {
			Logger.Error("加载模型文件失败", zap.Any("error", err))
			return err
		}
		_yr.Rt = rm
		rts = append(rts, _yr)
	}

	// 视频流处理
	sp := &StreamProcessor{
		DeviceId:       yp.DeviceId,
		RkPlatform:     yp.RkPlatform,
		Rts:            rts,
		VedioStreamUrl: yp.VedioStreamUrl,
		CancelFunc:     cancel,
		StreamId:       yp.StreamId,
		YoloResultCh:   make(chan map[string]YoloDetResultStat, 100),
	}
	yp.StreamProcessor = sp

	go sp.VedioStreamProcess(ctx)
	go sp.YoloResultProcess()

	return nil
}

func (yp *YoloProcessorStruct) StopYoloProcessor() {
	if yp.StreamProcessor != nil {
		yp.StreamProcessor.Stop()
	}
}
