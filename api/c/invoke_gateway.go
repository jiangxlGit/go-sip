package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go-sip/api/middleware"
	. "go-sip/common"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	"go-sip/utils"

	"go.uber.org/zap"

	"io"
)

// 获取sip服务信息 (提供给客户端调用)
func GetSipServerTcpAddr(gateway_base_url, device_id string) string {

	tcp_addr := ""

	// 获取网关地址
	gateway_url := fmt.Sprintf("http://%s%s", gateway_base_url, OpenSipServerInfoURL)

	// 将结构体编码为 JSON
	body := &model.SipServerInfoBody{}
	body.DeviceId = device_id
	jsonData, err := json.Marshal(body)
	if err != nil || jsonData == nil {
		Logger.Error("json marshal error")
	} else {
		// 调用网关接口
		req, err := http.NewRequest("POST", gateway_url, bytes.NewBuffer(jsonData))
		if err != nil {
			Logger.Error("http.NewRequest error", zap.Any("gateway_url", gateway_url), zap.Error(err))
		} else {
			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				Logger.Error("http.DefaultClient.Do error", zap.Any("gateway_url", gateway_url), zap.Error(err))
			} else {
				defer resp.Body.Close()
				// 读取 body 内容
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					Logger.Error("io.ReadAll error", zap.Any("gateway_url", gateway_url), zap.Error(err))
				} else {
					// 转为字符串
					// bodyString := string(bodyBytes)
					// json 解析
					result := model.ApiResult{}
					err := utils.JSONDecode(bodyBytes, &result)
					if err != nil || result.Code != model.CodeSucc {
						Logger.Error("连接错误", zap.Any("gateway_url", gateway_url), zap.Error(err))
					} else {
						// Logger.Info("http.DefaultClient.Do success", zap.Any("gateway_url", gateway_url), zap.Any("bodyString", bodyString))
						tcp_addr = result.Result.(string)
					}
				}
			}
		}
	}

	return tcp_addr
}

// 调用网关接口，获取设备关联的模型列表
func GetDeviceRelationAiModelList(deviceId string) (map[string][]*model.AiModelInfo, error) {
	if deviceId == "" {
		return nil, fmt.Errorf("deviceId is empty")
	}
	// 获取网关地址
	gateway_url := fmt.Sprintf("http://%s%s?deviceId=%s&deviceType=%s", m.CMConfig.Gateway, OpenAiModelRelationListURL, deviceId, m.CMConfig.DeviceType)

	// 调用网关接口
	httpClient := middleware.GetHttpClient(m.CMConfig.OpenApi.ClientId, m.CMConfig.OpenApi.SecretKey)
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient is nil")
	}
	resp, err := httpClient.Get(gateway_url)
	if err != nil {
		return nil, fmt.Errorf("http.Get error: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error("io.ReadAll error", zap.Any("gateway_url", gateway_url), zap.Error(err))
	} else {
		// json 解析
		result := model.IpcAiModelResult{}
		err := utils.JSONDecode(bodyBytes, &result)
		if err != nil || result.Code != model.CodeSucc {
			Logger.Error("GetAllAiModelList json.Unmarshal error", zap.Any("gateway_url", gateway_url), zap.Error(err))
		} else {
			// Logger.Info("http.DefaultClient.Do success", zap.Any("gateway_url", gateway_url))
			return result.Result, nil
		}
	}
	return nil, fmt.Errorf("get ai model list error")
}

// 调用网关接口，获取已关联并且已启用的模型列表
func GetAiModelRelationList(deviceId string) ([]*model.IotDeviceRelationAiModelInfo, error) {
	if deviceId == "" {
		return nil, fmt.Errorf("deviceId is empty")
	}
	// 获取网关地址
	gateway_url := fmt.Sprintf("http://%s%s?deviceId=%s", m.CMConfig.Gateway, OpenAiModelRelationListURL, deviceId)

	// 调用网关接口
	httpClient := middleware.GetHttpClient(m.CMConfig.OpenApi.ClientId, m.CMConfig.OpenApi.SecretKey)
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient is nil")
	}
	resp, err := httpClient.Get(gateway_url)
	if err != nil {
		return nil, fmt.Errorf("http.Get error: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error("io.ReadAll error", zap.Any("gateway_url", gateway_url), zap.Error(err))
	} else {
		// json 解析
		result := model.IotDeviceRelationAiModelResult{}
		err := utils.JSONDecode(bodyBytes, &result)
		if err != nil || result.Code != model.CodeSucc {
			Logger.Error("GetAllAiModelList json.Unmarshal error", zap.Any("gateway_url", gateway_url), zap.Error(err))
		} else {
			// Logger.Info("http.DefaultClient.Do success", zap.Any("gateway_url", gateway_url))
			return result.Result, nil
		}
	}
	return nil, fmt.Errorf("get ai model list error")
}

// 调用保存录制下的视频回放时间记录接口
func IpcPlaybackRecord(data model.ZLMRecordMp4Data) model.ApiResult {

	result := model.ApiResult{}

	// 获取网关地址
	gateway_url := fmt.Sprintf("http://%s%s", m.CMConfig.Gateway, OpenIpcPlaybackRecordURL)

	// 将结构体编码为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil || jsonData == nil {
		Logger.Error("json marshal error")
	} else {
		// 调用网关接口
		httpClient := middleware.GetHttpClient(m.CMConfig.OpenApi.ClientId, m.CMConfig.OpenApi.SecretKey)
		if httpClient != nil {
			resp, err := httpClient.Post(gateway_url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil || resp.StatusCode != http.StatusOK {
				Logger.Error("httpClient.Post error", zap.Any("gateway_url", gateway_url), zap.Error(err))
			} else {
				defer resp.Body.Close()
				// 读取 body 内容
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil || bodyBytes == nil {
					Logger.Error("io.ReadAll error", zap.Any("gateway_url", gateway_url), zap.Error(err))
				} else {
					// 转为字符串
					// bodyString := string(bodyBytes)
					// json 解析
					err := utils.JSONDecode(bodyBytes, &result)
					if err != nil || result.Code != model.CodeSucc {
						Logger.Error("IpcPlaybackRecord json.Unmarshal error", zap.Any("gateway_url", gateway_url), zap.Error(err))
					} else {
						// Logger.Info("http.DefaultClient.Do success", zap.Any("gateway_url", gateway_url), zap.Any("bodyString", bodyString))
						return result
					}
				}

			}
		}
	}
	result = model.ApiResult{
		Code:   model.CodeSysERR,
		Status: model.StatusSysERR,
		Result: "IpcPlaybackRecord error",
	}
	return result
}

// 调用网关接口，根据设备id获取所有ipc列表
func GetAllIpcList(deviceId string) ([]*model.IpcInfo, error) {
	if deviceId == "" {
		return nil, fmt.Errorf("deviceId is empty")
	}
	// 获取网关地址
	gateway_url := fmt.Sprintf("http://%s%s?deviceId=%s", m.CMConfig.Gateway, OpenGetIpcListURL, deviceId)

	// 调用网关接口
	httpClient := middleware.GetHttpClient(m.CMConfig.OpenApi.ClientId, m.CMConfig.OpenApi.SecretKey)
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient is nil")
	}
	resp, err := httpClient.Get(gateway_url)
	if err != nil {
		return nil, fmt.Errorf("http.Get error: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error("io.ReadAll error", zap.Any("gateway_url", gateway_url), zap.Error(err))
	} else {
		// json 解析
		result := model.IotIpcInfoResult{}
		err := utils.JSONDecode(bodyBytes, &result)
		if err != nil || result.Code != model.CodeSucc {
			Logger.Error("GetAllIpcList json.Unmarshal error", zap.Any("gateway_url", gateway_url), zap.Error(err))
		} else {
			Logger.Debug("GetAllIpcList success", zap.Any("gateway_url", gateway_url), zap.Any("result", result))
			return result.Result, nil
		}
	}
	return nil, fmt.Errorf("get not gb ipc list error")
}

// 调用网关接口，根据设备id获取非国标ipc列表
func GetNotGbIpcList(deviceId string) ([]*model.IotNotGbIpcInfo, error) {
	if deviceId == "" {
		return nil, fmt.Errorf("deviceId is empty")
	}
	// 获取网关地址
	gateway_url := fmt.Sprintf("http://%s%s?deviceId=%s", m.CMConfig.Gateway, OpenGetNotGbIpcListURL, deviceId)

	// 调用网关接口
	httpClient := middleware.GetHttpClient(m.CMConfig.OpenApi.ClientId, m.CMConfig.OpenApi.SecretKey)
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient is nil")
	}
	resp, err := httpClient.Get(gateway_url)
	if err != nil {
		return nil, fmt.Errorf("http.Get error: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error("io.ReadAll error", zap.Any("gateway_url", gateway_url), zap.Error(err))
	} else {
		// json 解析
		result := model.IotNotGbIpcInfoResult{}
		err := utils.JSONDecode(bodyBytes, &result)
		if err != nil || result.Code != model.CodeSucc {
			Logger.Error("GetNotGbIpcList json.Unmarshal error", zap.Any("gateway_url", gateway_url), zap.Error(err))
		} else {
			Logger.Debug("GetNotGbIpcList success", zap.Any("gateway_url", gateway_url))
			return result.Result, nil
		}
	}
	return nil, fmt.Errorf("get not gb ipc list error")
}

// 调用网关接口，更新非国标ipc信息
func IpcNotGbInfoUpdate(ipcId, status string) error {
	if ipcId == "" || status == "" {
		return fmt.Errorf("ipcId or status is empty")
	}
	if status != "ON" && status != "OFFLINE" && status != "ERROR" {
		return fmt.Errorf("status error")
	}
	// 获取网关地址
	gateway_url := fmt.Sprintf("http://%s%s?ipcId=%s&status=%s", m.CMConfig.Gateway, OpenGetNotGbIpcUpdateURL, ipcId, status)

	// 调用网关接口
	httpClient := middleware.GetHttpClient(m.CMConfig.OpenApi.ClientId, m.CMConfig.OpenApi.SecretKey)
	if httpClient == nil {
		return fmt.Errorf("httpClient is nil")
	}
	resp, err := httpClient.Get(gateway_url)
	if err != nil {
		return fmt.Errorf("http.Get error: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error("io.ReadAll error", zap.Any("gateway_url", gateway_url), zap.Error(err))
	} else {
		// json 解析
		result := model.ApiResult{}
		err := utils.JSONDecode(bodyBytes, &result)
		if err != nil || result.Code != model.CodeSucc {
			Logger.Error("IpcNotGbInfoUpdate json.Unmarshal error", zap.Any("gateway_url", gateway_url), zap.Error(err))
		} else {
			Logger.Debug("IpcNotGbInfoUpdate success", zap.Any("gateway_url", gateway_url))
			return nil
		}
	}
	return fmt.Errorf("IpcNotGbInfoUpdate error")
}
