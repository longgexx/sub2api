// Package alipay provides Alipay API client for payment monitoring.
package alipay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// BillRecord 账单记录
type BillRecord struct {
	TransLogID     string  // 交易流水号
	TransType      string  // 交易类型
	TransTime      string  // 交易时间
	Amount         float64 // 交易金额
	Balance        float64 // 余额
	Memo           string  // 备注
	OtherAccount   string  // 对方账户
	TransDirection string  // 交易方向: in(收入) / out(支出)
}

// Client 支付宝客户端
type Client struct {
	config     config.AlipayConfig
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	httpClient *http.Client
}

// NewClient 创建支付宝客户端
func NewClient(cfg config.AlipayConfig) (*Client, error) {
	client := &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 解析私钥
	if cfg.PrivateKey != "" {
		privateKey, err := parsePrivateKey(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		client.privateKey = privateKey
	}

	// 解析支付宝公钥
	if cfg.AlipayPublicKey != "" {
		publicKey, err := parsePublicKey(cfg.AlipayPublicKey)
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
		client.publicKey = publicKey
	}

	return client, nil
}

// parsePrivateKey 解析私钥
func parsePrivateKey(privateKeyStr string) (*rsa.PrivateKey, error) {
	// 移除 PEM 头尾和空白字符
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "-----BEGIN RSA PRIVATE KEY-----", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "-----END RSA PRIVATE KEY-----", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "-----BEGIN PRIVATE KEY-----", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "-----END PRIVATE KEY-----", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "\n", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, "\r", "")
	privateKeyStr = strings.ReplaceAll(privateKeyStr, " ", "")

	// Base64 解码
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}

	// 尝试 PKCS8 格式
	key, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("not an RSA private key")
	}

	// 尝试 PKCS1 格式
	rsaKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return rsaKey, nil
}

// parsePublicKey 解析公钥
func parsePublicKey(publicKeyStr string) (*rsa.PublicKey, error) {
	// 移除 PEM 头尾和空白字符
	publicKeyStr = strings.ReplaceAll(publicKeyStr, "-----BEGIN PUBLIC KEY-----", "")
	publicKeyStr = strings.ReplaceAll(publicKeyStr, "-----END PUBLIC KEY-----", "")
	publicKeyStr = strings.ReplaceAll(publicKeyStr, "\n", "")
	publicKeyStr = strings.ReplaceAll(publicKeyStr, "\r", "")
	publicKeyStr = strings.ReplaceAll(publicKeyStr, " ", "")

	// Base64 解码
	keyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}

	// 解析公钥
	pub, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaKey, nil
}

// Sign RSA2 签名
func (c *Client) Sign(content string) (string, error) {
	if c.privateKey == nil {
		return "", errors.New("private key not configured")
	}

	h := sha256.New()
	h.Write([]byte(content))
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify 验证签名
func (c *Client) Verify(content, sign string) error {
	if c.publicKey == nil {
		return errors.New("public key not configured")
	}

	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write([]byte(content))
	hashed := h.Sum(nil)

	return rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, hashed, signBytes)
}

// buildCommonParams 构建公共参数
func (c *Client) buildCommonParams(method string) map[string]string {
	signType := strings.TrimSpace(c.config.SignType)
	if signType == "" {
		signType = "RSA2"
	}
	return map[string]string{
		"app_id":    c.config.AppID,
		"method":    method,
		"format":    "JSON",
		"charset":   "UTF-8",
		"sign_type": signType,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"version":   "1.0",
	}
}

// buildSignContent 构建签名内容
func (c *Client) buildSignContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	return strings.Join(parts, "&")
}

// Execute 执行 API 请求
func (c *Client) Execute(method string, bizContent map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()
	params := c.buildCommonParams(method)

	// 序列化业务内容
	if bizContent != nil {
		bizContentBytes, err := json.Marshal(bizContent)
		if err != nil {
			return nil, err
		}
		params["biz_content"] = string(bizContentBytes)
	}

	// 签名
	signContent := c.buildSignContent(params)
	sign, err := c.Sign(signContent)
	if err != nil {
		log.Printf("[Alipay] API %s sign failed: %v", method, err)
		return nil, fmt.Errorf("sign failed: %w", err)
	}
	params["sign"] = sign

	// 构建请求 URL
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	// 发送请求
	resp, err := c.httpClient.PostForm(c.config.ServerURL, values)
	if err != nil {
		log.Printf("[Alipay] API %s request failed: %v (took %v)", method, err, time.Since(startTime))
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Alipay] API %s read response failed: %v (took %v)", method, err, time.Since(startTime))
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	// 支付宝 API 返回 GBK 编码（忽略 charset 参数），需要转换为 UTF-8
	body, err = simplifiedchinese.GBK.NewDecoder().Bytes(body)
	if err != nil {
		log.Printf("[Alipay] API %s GBK to UTF-8 conversion failed: %v (took %v)", method, err, time.Since(startTime))
		return nil, fmt.Errorf("encoding conversion failed: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[Alipay] API %s parse response failed: %v, body: %s (took %v)", method, err, string(body), time.Since(startTime))
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// 记录 API 调用结果
	duration := time.Since(startTime)
	respKey := strings.ReplaceAll(method, ".", "_") + "_response"
	if respData, ok := result[respKey].(map[string]interface{}); ok {
		code, _ := respData["code"].(string)
		msg, _ := respData["msg"].(string)
		if code == "10000" {
			log.Printf("[Alipay] API %s success (took %v)", method, duration)
		} else {
			subCode, _ := respData["sub_code"].(string)
			subMsg, _ := respData["sub_msg"].(string)
			log.Printf("[Alipay] API %s failed: code=%s, msg=%s, sub_code=%s, sub_msg=%s (took %v)", method, code, msg, subCode, subMsg, duration)
		}
	} else {
		// 响应格式异常
		log.Printf("[Alipay] API %s unexpected response format (took %v)", method, duration)
	}

	return result, nil
}

// QueryBills 查询账单
func (c *Client) QueryBills(startTime, endTime string, pageNo, pageSize int) ([]BillRecord, error) {
	bizContent := map[string]interface{}{
		"start_time": startTime,
		"end_time":   endTime,
		"page_no":    fmt.Sprintf("%d", pageNo),
		"page_size":  fmt.Sprintf("%d", pageSize),
	}

	result, err := c.Execute("alipay.data.bill.accountlog.query", bizContent)
	if err != nil {
		return nil, err
	}

	// 解析响应
	respKey := "alipay_data_bill_accountlog_query_response"
	respData, ok := result[respKey].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	// 检查错误码
	code, _ := respData["code"].(string)
	if code != "10000" {
		msg, _ := respData["msg"].(string)
		subMsg, _ := respData["sub_msg"].(string)
		return nil, fmt.Errorf("API error: %s - %s", msg, subMsg)
	}

	// 解析账单记录
	var bills []BillRecord
	detailList, ok := respData["detail_list"].([]interface{})
	if !ok {
		return bills, nil // 没有账单记录
	}

	for _, item := range detailList {
		if itemMap, ok := item.(map[string]interface{}); ok {
			bill := BillRecord{
				TransLogID:     getString(itemMap, "account_log_id"),
				TransType:      getString(itemMap, "type"),
				TransTime:      getString(itemMap, "trans_dt"),
				Amount:         getFloat64(itemMap, "trans_amount"),
				Balance:        getFloat64(itemMap, "balance"),
				Memo:           getString(itemMap, "trans_memo"),
				OtherAccount:   getString(itemMap, "other_account"),
				TransDirection: getString(itemMap, "direction"),
			}
			bills = append(bills, bill)
		}
	}

	return bills, nil
}

// QueryTodayBills 查询今日账单
func (c *Client) QueryTodayBills() ([]BillRecord, error) {
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endTime := now

	return c.QueryBills(
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		1, 2000,
	)
}

// QueryRecentBills 查询最近指定分钟内的账单
func (c *Client) QueryRecentBills(minutes int) ([]BillRecord, error) {
	now := time.Now()
	startTime := now.Add(-time.Duration(minutes) * time.Minute)

	return c.QueryBills(
		startTime.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"),
		1, 2000,
	)
}

// IsConfigured 检查是否已配置
func (c *Client) IsConfigured() bool {
	return c.config.AppID != "" && c.privateKey != nil
}

// ValidateConfig 校验配置是否正确（尝试调用账单查询接口）
// 返回 nil 表示配置正确，否则返回错误信息
func (c *Client) ValidateConfig() error {
	if c.config.AppID == "" {
		return errors.New("app_id is not configured")
	}
	if c.privateKey == nil {
		return errors.New("private_key is not configured or invalid")
	}
	if c.config.ServerURL == "" {
		return errors.New("server_url is not configured")
	}

	// 尝试查询最近 1 分钟的账单来验证配置
	now := time.Now()
	startTime := now.Add(-1 * time.Minute)

	_, err := c.QueryBills(
		startTime.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"),
		1, 1,
	)
	if err != nil {
		return fmt.Errorf("API validation failed: %w", err)
	}

	return nil
}

// GetConfigStatus 获取配置状态摘要
func (c *Client) GetConfigStatus() map[string]interface{} {
	status := map[string]interface{}{
		"configured":          c.IsConfigured(),
		"app_id_set":          c.config.AppID != "",
		"private_key_set":     c.privateKey != nil,
		"public_key_set":      c.publicKey != nil,
		"server_url":          c.config.ServerURL,
	}
	if c.config.AppID != "" {
		// 只显示 app_id 的前4位和后4位
		appID := c.config.AppID
		if len(appID) > 8 {
			status["app_id_masked"] = appID[:4] + "****" + appID[len(appID)-4:]
		} else {
			status["app_id_masked"] = "****"
		}
	}
	return status
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	}
	return 0
}
