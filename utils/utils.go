package utils

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	. "go-sip/logger"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"hash/crc32"

	"encoding/hex"
	"sort"
	"sync"
)

// Error Error
type Error struct {
	err    error
	params []interface{}
}

func (err *Error) Error() string {
	if err == nil {
		return "<nil>"
	}
	str := fmt.Sprint(err.params...)
	if err.err != nil {
		str += fmt.Sprintf(" err:%s", err.err.Error())
	}
	return str
}

// NewError NewError
func NewError(err error, params ...interface{}) error {
	return &Error{err, params}
}

// JSONEncode JSONEncode
func JSONEncode(data interface{}) []byte {
	d, err := json.Marshal(data)
	if err != nil {
		Logger.Error("JSONEncode error:", zap.Error(err))
	}
	return d
}

// JSONDecode JSONDecode
func JSONDecode(data []byte, obj interface{}) error {
	return json.Unmarshal(data, obj)
}

func RandInt(min, max int) int {
	if max < min {
		return 0
	}
	max++
	max -= min
	rand.Seed(time.Now().UnixNano())
	r := rand.Int()
	return r%max + min
}

const (
	letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// RandString https://github.com/kpbird/golang_random_string
func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	output := make([]byte, n)
	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)
	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)
	// fill output
	for pos := range output {
		// get random item
		random := randomness[pos]
		// random % 64
		randomPos := random % uint8(l)
		// put into output
		output[pos] = letterBytes[randomPos]
	}

	return string(output)
}

func timeoutClient() *http.Client {
	connectTimeout := time.Duration(20 * time.Second)
	readWriteTimeout := time.Duration(30 * time.Second)
	return &http.Client{
		Transport: &http.Transport{
			DialContext:         timeoutDialer(connectTimeout, readWriteTimeout),
			MaxIdleConnsPerHost: 200,
			DisableKeepAlives:   true,
		},
	}
}
func timeoutDialer(cTimeout time.Duration,
	rwTimeout time.Duration) func(ctx context.Context, net, addr string) (c net.Conn, err error) {
	return func(ctx context.Context, netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

// PostRequest PostRequest
func PostRequest(url string, bodyType string, body io.Reader) ([]byte, error) {
	client := timeoutClient()
	resp, err := client.Post(url, bodyType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respbody, nil
}

// PostJSONRequest PostJSONRequest
func PostJSONRequest(url string, data interface{}) ([]byte, error) {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return PostRequest(url, "application/json;charset=UTF-8", bytes.NewReader(bytesData))
}

// GetRequest GetRequest
func GetRequest(url string) ([]byte, error) {
	client := timeoutClient()
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respbody, nil
}

// GetMD5 GetMD5
func GetMD5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// XMLDecode XMLDecode
func XMLDecode(data []byte, v interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(v)
}

// Max Max
func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ResolveSelfIP ResolveSelfIP
func ResolveSelfIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip, nil
		}
	}
	return nil, errors.New("server not connected to any network")
}

// GBK 转 UTF-8
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// UTF-8 转 GBK
func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func GetInfoCseq() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(900_000_000) + 100_000_000 // [100000000, 999999999]
}

// Hash 函数
func HashString(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}

// 删除list中的元素
func RemoveListByValue(slice []string, value string) []string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in RemoveStringValue:", r)
		}
	}()

	if slice == nil {
		panic("input slice is nil")
	}

	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}

var (
	rnd  *rand.Rand
	once sync.Once
)

// EncodeMD5 md5 encryption
func EncodeMD5(value string) string {
	m := md5.New()
	m.Write([]byte(value))

	return hex.EncodeToString(m.Sum(nil))
}

// 过滤list中不符合条件的元素
func Filter[T any](input []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range input {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// 稳定排序：按指定 string 字段升序或降序
func SortByFieldStable[T any](list []T, keyExtractor func(T) string, asc bool) {
	sort.SliceStable(list, func(i, j int) bool {
		ki := keyExtractor(list[i])
		kj := keyExtractor(list[j])
		if asc {
			return ki < kj
		}
		return ki > kj
	})
}

type JSONTime time.Time

const timeLayout = "2006-01-02 15:04:05"

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format(timeLayout))
}

func (t *JSONTime) UnmarshalJSON(b []byte) error {
	// 去掉引号
	str := string(b)
	str = str[1 : len(str)-1]

	// 解析时间
	parsedTime, err := time.ParseInLocation(timeLayout, str, time.Local)
	if err != nil {
		return err
	}

	*t = JSONTime(parsedTime)
	return nil
}

func (t JSONTime) String() string {
	return time.Time(t).Format(timeLayout)
}

// Contains 判断一个 slice 是否包含目标元素
func Contains[T comparable](list []T, target T) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

// NowInCn 返回当前上海时间（Asia/Shanghai）
func NowInCn() time.Time {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如有错误，默认使用系统本地时间（可以根据需要选择 panic 或记录错误）
		return time.Now()
	}
	return time.Now().In(loc)
}

// 分片函数：平均分成 n 份
func SplitSlice[T any](list []T, n int) [][]T {
	if n <= 0 {
		panic("n 必须 > 0")
	}
	var res [][]T
	total := len(list)
	chunkSize := (total + n - 1) / n

	for i := 0; i < total; i += chunkSize {
		end := i + chunkSize
		if end > total {
			end = total
		}
		res = append(res, list[i:end])
	}
	return res
}

func DecodeURLFromJSON(raw string) (string, error) {
	var decoded string
	err := json.Unmarshal([]byte(`"`+raw+`"`), &decoded)
	if err != nil {
		Logger.Error("DecodeURLFromJSON json反序列化失败", zap.Error(err))
		return "", err
	}
	return decoded, nil
}

// FromReader 计算 io.Reader 的 MD5 值（边读边计算，节省内存）
func MD5FromReader(r io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CheckOrCreateDirFast 高效判断目录是否存在并具备读写权限，不存在则尝试创建
func CheckOrCreateDir(dir string) error {
	// 检查路径状态
	if fi, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			// 不存在则创建
			if err := os.MkdirAll(dir, 0755); err != nil {
				return errors.New("目录不存在且创建失败: " + err.Error())
			}
		} else {
			return errors.New("目录状态检查失败: " + err.Error())
		}
	} else if !fi.IsDir() {
		return errors.New("路径已存在但不是目录")
	}

	// 合并读写/创建权限判断：创建并删除一个临时文件
	tmpFile := filepath.Join(dir, ".zlm_perm_test")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		return errors.New("目录不可写或无法创建文件: " + err.Error())
	}
	_ = os.Remove(tmpFile) // 尽量清理，但不阻止成功返回

	return nil
}

// 检测目录是否存在
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir() // 也可以改为 true，如果你只想判断路径存在
}

func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}
	// double check（可选）
	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		return fmt.Errorf("删除失败，文件仍然存在")
	}
	return nil
}

// 删除某个目录下的所有.rknn文件
func DeleteRKNNFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue // 忽略子目录
		}
		if strings.HasSuffix(entry.Name(), ".rknn") {
			filePath := filepath.Join(dir, entry.Name())
			err := os.Remove(filePath)
			if err != nil {
				fmt.Printf("删除失败: %s, 错误: %v\n", filePath, err)
			} else {
				fmt.Printf("已删除: %s\n", filePath)
			}
		}
	}
	return nil
}

// 拆分a,b,c,a型的字符串，并检查是否有重复的元素
func StrArrCheckDuplicates(str string) ([]string, error) {
	if str == "" {
		return nil, fmt.Errorf("参数不能为空")
	}
	var result = []string{}
	parts := strings.Split(str, ",")
	seen := make(map[string]bool)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		result = append(result, part)
	}
	return result, nil
}

// PingIP 尝试 ping 一个 IP 一次，返回 true 表示可达
func PingIP(ip string) bool {
	// -c 1 = 发送 1 个包, -W 1 = 等待 1 秒超时
	out, err := exec.Command("ping", "-c", "1", "-W", "1", ip).CombinedOutput()
	if err != nil {
		return false
	}
	// 一般成功的返回中会包含 "ttl="
	return strings.Contains(string(out), "ttl=")
}

// PingWithRetry 对某个 IP ping 5 次，成功 >= 3 返回 true
func PingWithRetry(ip string) bool {
	success := 0
	for i := 0; i < 5; i++ {
		if PingIP(ip) {
			success++
		}
		if i < 4 { // 最后一次不需要再 sleep
			time.Sleep(1 * time.Second)
		}
	}
	return success >= 3
}

// 根据网卡名获取 IP
func getIPByInterface(c, ifname string) string {
	// 使用 ip addr 命令
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s show %s | grep 'inet ' | awk '{print $2}' | cut -d/ -f1", c, ifname))
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	ip := strings.TrimSpace(string(out))
	return ip
}

// 自动 fallback 获取 IP
func GetPreferredIP() string {
	interfaces := []string{"eth0", "wlan0", "wlan1"} // 按优先级排列
	for _, iface := range interfaces {
		ip := getIPByInterface("ip addr", iface)
		if ip != "" {
			return ip
		}
	}
	return "" // 没有可用 IP
}

// CheckPort 单个IP端口检测
func CheckPort(ip, port string, timeout time.Duration) bool {
	address := net.JoinHostPort(ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// CopyFile 将一个文件复制到指定目录下（覆盖已有文件）
func CopyFile(srcFile, dstDir string) error {
	info, err := os.Stat(srcFile)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("源路径 %s 是目录，不是文件", srcFile)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return err
	}

	// 拼接目标文件路径：dstDir + basename(srcFile)
	dstFile := filepath.Join(dstDir, filepath.Base(srcFile))

	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstFile) // 覆盖模式
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	// 保持文件权限
	return os.Chmod(dstFile, info.Mode())
}

// EqualStringSliceSet 比较两个 string 切片是否相等（不考虑顺序，忽略空字符串）
func EqualStringSliceSet(a, b []string) bool {
	count := make(map[string]int)

	// 统计 a 中非空元素
	for _, v := range a {
		if v != "" {
			count[v]++
		}
	}

	// 用 b 抵消计数
	for _, v := range b {
		if v != "" {
			count[v]--
		}
	}

	// 检查是否都抵消为 0
	for _, v := range count {
		if v != 0 {
			return false
		}
	}
	return true
}

// StrToInt 尝试将字符串转为 int，失败时返回 defaultVal
func StrToInt(s string, defaultVal int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return defaultVal
}
