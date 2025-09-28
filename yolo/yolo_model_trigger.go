package yolo

import (
	"fmt"
	. "go-sip/logger"
	"reflect"

	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	GlobalModelTriggerMonitor *ModelTriggerMonitor
	once                      sync.Once
)

// 模型触发事件类型
type TriggerEvent struct {
	Key       string
	ClassName string
	Event     string // "start" or "stop"
	Time      time.Time
}

type ModelTriggerMonitor struct {
	mu        sync.Map // key: string, value: *triggerInfo
	timeout   time.Duration
	eventCh   map[string]chan TriggerEvent
	eventChMu sync.RWMutex  // 保护 eventCh 的读写
	updateCh  chan struct{} // 流增删时通知 select goroutine
}

type triggerInfo struct {
	mu               sync.Mutex
	FirstTriggerFlag bool               // 是否首次触发
	TriggerCount     int                // 触发计数
	FirstTime        time.Time          // 本轮计数的起始时间
	hb               chan struct{}      // 心跳通道
	CancelFunc       context.CancelFunc // 取消函数
}

func GetAiModelTriggerMonitor(streamId string, timeout time.Duration) *ModelTriggerMonitor {
	once.Do(func() {
		GlobalModelTriggerMonitor = &ModelTriggerMonitor{
			timeout:  timeout,
			eventCh:  make(map[string]chan TriggerEvent),
			updateCh: make(chan struct{}, 1),
		}
	})
	GlobalModelTriggerMonitor.addAiEventChan(streamId)
	return GlobalModelTriggerMonitor
}

// 添加/更新流通道
func (m *ModelTriggerMonitor) addAiEventChan(streamId string) chan TriggerEvent {
	m.eventChMu.Lock()
	defer m.eventChMu.Unlock()
	ch := make(chan TriggerEvent, 100) // 缓冲可调
	m.eventCh[streamId] = ch
	// 通知 select goroutine 更新 cases
	select {
	case m.updateCh <- struct{}{}:
	default:
	}
	return ch
}

// 删除流通道
func (m *ModelTriggerMonitor) removeAiEventChan(streamId string) {
	m.eventChMu.Lock()
	defer m.eventChMu.Unlock()
	if ch, ok := m.eventCh[streamId]; ok {
		close(ch)
		delete(m.eventCh, streamId)
	}
	// 通知 select goroutine 更新 cases
	select {
	case m.updateCh <- struct{}{}:
	default:
	}
}

// 构建 reflect.SelectCase
func (m *ModelTriggerMonitor) BuildSelectCases() ([]reflect.SelectCase, []string) {
	m.eventChMu.RLock()
	defer m.eventChMu.RUnlock()

	cases := make([]reflect.SelectCase, 0, len(m.eventCh)+1)
	keys := make([]string, 0, len(m.eventCh))

	// 第一个 case 用来监听 updateCh，通知新增/删除流
	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(m.updateCh),
	})

	for key, ch := range m.eventCh {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
		keys = append(keys, key)
	}
	return cases, keys
}

// EventChan 获取全局事件通道（其他业务可消费）
func (m *ModelTriggerMonitor) GetAiEventChan() map[string]chan TriggerEvent {
	return m.eventCh
}

// Trigger 被调用时表示某个 modelTriggerKey 有触发
func (m *ModelTriggerMonitor) AiEventTrigger(streamId, className, rkPlatform string) {
	modelTriggerKey := fmt.Sprintf("%s_%s_%s", streamId, className, rkPlatform)
	now := time.Now()
	val, _ := m.mu.LoadOrStore(streamId, &triggerInfo{
		FirstTriggerFlag: false,
		TriggerCount:     0,
		FirstTime:        now,
		hb:               make(chan struct{}, 1),
	})
	trigger := val.(*triggerInfo)
	trigger.mu.Lock()
	defer trigger.mu.Unlock()

	// 没有触发时业务
	if !trigger.FirstTriggerFlag {
		// 更新触发计数
		if now.Sub(trigger.FirstTime) > 5*time.Second {
			// 超过 5 秒重置计数
			trigger.FirstTime = now
			trigger.TriggerCount = 1
		} else {
			trigger.TriggerCount++
		}
		// 需要在 5 秒内触发 2 次以上才算有效
		if trigger.TriggerCount >= 2 {
			Logger.Info("首次触发模型", zap.String("modelTriggerKey", modelTriggerKey))
			trigger.FirstTriggerFlag = true

			// 发送开始指令
			triggerStartEvent := TriggerEvent{Key: modelTriggerKey, ClassName: className, Event: "start", Time: now}

			select {
			case m.eventCh[streamId] <- triggerStartEvent:
			default:
				Logger.Error("事件通道已满，丢弃 start 事件", zap.String("key", modelTriggerKey))
			}

			// 启动超时监控
			triggerCtx, cancel := context.WithCancel(context.Background())
			trigger.CancelFunc = cancel
			go m.aiEventTriggerTimeout(triggerCtx, streamId, modelTriggerKey, className, trigger.hb)
		}
	}
	// 发送心跳，重置定时器（非阻塞）
	select {
	case trigger.hb <- struct{}{}:
	default:
	}
}

func (m *ModelTriggerMonitor) aiEventTriggerTimeout(ctx context.Context, streamId, key, className string, hb <-chan struct{}) {
	timer := time.NewTimer(m.timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			m.mu.Delete(streamId)
			Logger.Debug("超时未再触发模型退出", zap.String("modelTriggerKey", key), zap.Duration("timeout", m.timeout))
			return
		case <-hb:
			// 收到心跳，重置定时器
			if !timer.Stop() {
				// 排空已触发但未消费的信号，避免“粘连”
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(m.timeout)
		case <-timer.C:
			Logger.Debug("超时未再触发模型", zap.String("modelTriggerKey", key), zap.Duration("timeout", m.timeout))
			// 发送终止指令
			triggerStopEvent := TriggerEvent{Key: key, ClassName: className, Event: "stop", Time: time.Now()}
			select {
			case m.eventCh[streamId] <- triggerStopEvent:
			default:
				Logger.Error("事件通道已满，丢弃 stop 事件", zap.String("key", key))
			}

			m.mu.Delete(streamId)
			return
		}
	}
}

// Stop 停止某个 streamId 的监控
func (m *ModelTriggerMonitor) AiEventTriggerStop(streamId string) {
	if val, ok := m.mu.Load(streamId); ok {
		info := val.(*triggerInfo)
		if info.CancelFunc != nil {
			info.CancelFunc()
		}
		Logger.Debug("停止模型触发", zap.String("streamId", streamId))
		m.removeAiEventChan(streamId)
	}
}
