package wvp

import (
	"encoding/json"
	"errors"
	"go-sip/db/redis"
	redis_util "go-sip/db/redis/redis_wvp_util"
	. "go-sip/logger"
	"go-sip/m"
	"go-sip/model"
	"go-sip/utils"
	"strconv"

	"net/url"
	"strings"

	"go.uber.org/zap"

	"fmt"
	"io"
	"net/http"
)

// 根据ipc_id调用sip服务接口
func WvpIpcGetRequestHandler(stream_id, api_url string, params url.Values) *m.Response {
	stream_id_arr := strings.Split(stream_id, "_")
	ipc_id := stream_id_arr[0]

	// 查询redis获取ipc列表
	device_ipc_info_str, err := redis_util.HGet_2(redis.DEVICE_IPC_INFO_KEY, ipc_id)
	if err != nil || device_ipc_info_str == "" {
		return nil
	}

	ipc_info := model.IpcInfo{}
	// 反序列化
	err = json.Unmarshal([]byte(device_ipc_info_str), &ipc_info)
	if err != nil {
		Logger.Error("json反序列化失败", zap.Error(err))
		return nil
	}

	return wvpInvokeSipServer(ipc_info.SipId, api_url, params)
}

// 根据device_id调用sip服务接口
func WvpDeviceGetRequestHandler(device_id, api_url string, params url.Values) *m.Response {
	// 获取sip_id
	sip_id, err := redis_util.HGet_2(redis.DEVICE_SIP_KEY, device_id)
	if err != nil {
		return nil
	}
	return wvpInvokeSipServer(sip_id, api_url, params)
}

func wvpInvokeSipServer(sip_id, api_url string, params url.Values) *m.Response {
	// 根据sip_id获取sip_ip:sip_port
	sip_url, err := redis_util.HGet_2(redis.SIP_SERVER_HOST, sip_id)
	if err != nil {
		Logger.Error("ipc_id未注册，请检查摄像头是否正常")
		return nil
	}
	// 使用sip_url调用sip服务接口
	base_request_url := fmt.Sprintf("http://%s%s", sip_url, api_url)

	full_url := fmt.Sprintf("%s?%s", base_request_url, params.Encode())

	// 调用sip接口
	Logger.Info("wvpInvokeSipServer full_url", zap.Any("full_url", full_url))
	req, err := http.NewRequest("GET", full_url, nil)
	if err != nil {
		Logger.Error("参数格式错误，json序列化失败", zap.Any("error", err))
		return nil
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Logger.Error("调用sip接口失败", zap.Any("error", err))
		return nil
	}
	// 读取响应体
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		Logger.Error("调用sip接口响应状态异常", zap.Any("status", resp.StatusCode))
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Error("调用sip接口读取body失败", zap.Any("error", err))
		return nil
	}
	var response m.Response
	err2 := json.Unmarshal(body, &response)
	if err2 != nil {
		return nil
	}
	return &response
}

// 根据设备id查询关联的zlm，如果没有关联，则随机获取一个，获取后记录需要选择的zlm
func WvpGetZlmInfo(device_id string) (*model.ZlmInfo, error) {
	if device_id == "" {
		return nil, errors.New("设备id不能为空")
	}

	zlmInfo := &model.ZlmInfo{}

	// 根据设备id查询关联地区码
	region_code, err := redis_util.HGet_4(redis.IOT_DEVICE_REGION_KEY, device_id)
	if err != nil || region_code == "" {
		Logger.Error("获取region_code失败", zap.Error(err))
		return nil, errors.New("获取region_code失败")
	}
	// 根据地区码获取zlm信息列表
	zlm_info_list_str, err := redis_util.HGet_2(redis.WVP_REGION_RELATION_ZLM_INFO, region_code)
	if err != nil || zlm_info_list_str == "" {
		Logger.Error("根据地区码获取zlm信息列表失败", zap.Error(err))
		return nil, errors.New("根据地区码获取zlm信息列表失败")
	}
	// 反序列化
	var zlmAndRegionInfoList []model.ZlmAndRegionInfo
	err = json.Unmarshal([]byte(zlm_info_list_str), &zlmAndRegionInfoList)
	if err != nil {
		Logger.Error("zlm信息列表json字符串反序列化失败", zap.Error(err))
		return nil, errors.New("zlm信息列表json字符串反序列化失败")
	}
	if len(zlmAndRegionInfoList) == 0 {
		Logger.Error("zlm信息列表为空")
		return nil, errors.New("zlm信息列表为空")
	}

	zlmDomain, _ := redis_util.HGet_2(redis.DEVICE_ZLM_KEY, device_id)
	var isEnableZlm = false
	var zlmAndRegionInfoNewList []model.ZlmAndRegionInfo
	for _, zlmAndRegionInfo := range zlmAndRegionInfoList {
		if zlmAndRegionInfo.ZlmNodeStatus == "enable" {
			zlmAndRegionInfoNewList = append(zlmAndRegionInfoNewList, zlmAndRegionInfo)
			if zlmAndRegionInfo.ZlmDomain == zlmDomain {
				isEnableZlm = true
			}
		}
	}
	if len(zlmAndRegionInfoNewList) == 0 {
		Logger.Error("zlm信息列表为空")
		return nil, errors.New("zlm信息列表为空")
	}
	if isEnableZlm {
		zlm_info_str, err := redis_util.HGet_2(redis.WVP_ZLM_NODE_INFO, zlmDomain)
		if err != nil || zlm_info_str == "" {
			Logger.Error("获取zlm信息失败", zap.Error(err))
			return nil, errors.New("获取zlm信息失败")
		}
		// 反序列化 JSON 字符串
		err = json.Unmarshal([]byte(zlm_info_str), &zlmInfo)
		if err != nil {
			Logger.Error("获取zlm信息失败", zap.Error(err))
			return nil, errors.New("获取zlm信息失败")
		}
	} else {
		// 从 zlm_info_list 中使用hash算法选择一个zlm服务
		zlmAndRegionInfo, err := selectZlmConfig(zlmAndRegionInfoNewList, device_id)
		if err != nil || zlmAndRegionInfo == nil {
			Logger.Error("获取zlm服务信息失败", zap.Error(err))
			return nil, errors.New("获取zlm服务信息失败")
		}
		zlmInfo = &model.ZlmInfo{
			ZlmIp:     zlmAndRegionInfo.ZlmIp,
			ZlmPort:   strconv.Itoa(zlmAndRegionInfo.ZlmPort),
			ZlmSecret: zlmAndRegionInfo.ZlmSecret,
			ZlmDomain: zlmAndRegionInfo.ZlmDomain,
		}
		redis_util.HSet_2(redis.DEVICE_ZLM_KEY, device_id, zlmAndRegionInfo.ZlmDomain)
	}
	Logger.Debug("GetZlmInfo", zap.Any("zlmInfo", zlmInfo))
	return zlmInfo, nil
}

// 从 configs 中使用hash取余选出一个元素
func selectZlmConfig(configs []model.ZlmAndRegionInfo, key string) (*model.ZlmAndRegionInfo, error) {
	if len(configs) == 0 {
		return nil, errors.New("配置列表为空")
	}
	if key == "" {
		return nil, errors.New("key不能为空")
	}
	index := int(utils.HashString(key)) % len(configs)
	return &configs[index], nil
}
