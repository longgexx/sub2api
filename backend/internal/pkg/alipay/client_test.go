//go:build unit

package alipay

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// 生成测试用的 RSA 密钥对
func generateTestKeyPair(t *testing.T) (string, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// 导出私钥 (PKCS1)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)

	// 导出公钥
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	return privateKeyBase64, publicKeyBase64
}

func TestNewClient_NoKeys(t *testing.T) {
	cfg := config.AlipayConfig{
		AppID:     "test_app_id",
		ServerURL: "https://openapi.alipay.com/gateway.do",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Nil(t, client.privateKey)
	require.Nil(t, client.publicKey)
}

func TestNewClient_WithKeys(t *testing.T) {
	privateKeyBase64, publicKeyBase64 := generateTestKeyPair(t)

	cfg := config.AlipayConfig{
		AppID:           "test_app_id",
		ServerURL:       "https://openapi.alipay.com/gateway.do",
		PrivateKey:      privateKeyBase64,
		AlipayPublicKey: publicKeyBase64,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.privateKey)
	require.NotNil(t, client.publicKey)
}

func TestNewClient_InvalidPrivateKey(t *testing.T) {
	cfg := config.AlipayConfig{
		AppID:      "test_app_id",
		PrivateKey: "invalid-key",
	}

	_, err := NewClient(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse private key")
}

func TestNewClient_InvalidPublicKey(t *testing.T) {
	cfg := config.AlipayConfig{
		AppID:           "test_app_id",
		AlipayPublicKey: "invalid-key",
	}

	_, err := NewClient(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse public key")
}

func TestClient_Sign(t *testing.T) {
	privateKeyBase64, _ := generateTestKeyPair(t)

	cfg := config.AlipayConfig{
		PrivateKey: privateKeyBase64,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	content := "app_id=test&biz_content={}"
	signature, err := client.Sign(content)
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	// 签名应该是 base64 编码的
	_, err = base64.StdEncoding.DecodeString(signature)
	require.NoError(t, err)
}

func TestClient_Sign_NoPrivateKey(t *testing.T) {
	cfg := config.AlipayConfig{}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	_, err = client.Sign("test content")
	require.Error(t, err)
	require.Contains(t, err.Error(), "private key not configured")
}

func TestClient_Verify(t *testing.T) {
	privateKeyBase64, publicKeyBase64 := generateTestKeyPair(t)

	cfg := config.AlipayConfig{
		PrivateKey:      privateKeyBase64,
		AlipayPublicKey: publicKeyBase64,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	content := "app_id=test&biz_content={}"

	// 签名
	signature, err := client.Sign(content)
	require.NoError(t, err)

	// 验证
	err = client.Verify(content, signature)
	require.NoError(t, err)
}

func TestClient_Verify_InvalidSignature(t *testing.T) {
	_, publicKeyBase64 := generateTestKeyPair(t)

	cfg := config.AlipayConfig{
		AlipayPublicKey: publicKeyBase64,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	err = client.Verify("test content", "invalid-signature")
	require.Error(t, err)
}

func TestClient_Verify_NoPublicKey(t *testing.T) {
	cfg := config.AlipayConfig{}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	err = client.Verify("test content", "signature")
	require.Error(t, err)
	require.Contains(t, err.Error(), "public key not configured")
}

func TestClient_BuildCommonParams(t *testing.T) {
	cfg := config.AlipayConfig{
		AppID:    "test_app_id",
		SignType: "RSA2",
	}

	client, _ := NewClient(cfg)
	params := client.buildCommonParams("alipay.trade.query")

	require.Equal(t, "test_app_id", params["app_id"])
	require.Equal(t, "alipay.trade.query", params["method"])
	require.Equal(t, "JSON", params["format"])
	require.Equal(t, "UTF-8", params["charset"])
	require.Equal(t, "RSA2", params["sign_type"])
	require.Equal(t, "1.0", params["version"])
	require.NotEmpty(t, params["timestamp"])
}

func TestClient_BuildCommonParams_DefaultSignType(t *testing.T) {
	cfg := config.AlipayConfig{
		AppID: "test_app_id",
	}

	client, _ := NewClient(cfg)
	params := client.buildCommonParams("alipay.trade.query")

	require.Equal(t, "RSA2", params["sign_type"])
}

func TestClient_BuildSignContent(t *testing.T) {
	cfg := config.AlipayConfig{}
	client, _ := NewClient(cfg)

	params := map[string]string{
		"app_id":    "test_app",
		"method":    "test.method",
		"sign":      "should_be_excluded",
		"timestamp": "2024-01-01 00:00:00",
		"empty":     "",
	}

	content := client.buildSignContent(params)

	// sign 和空值应该被排除
	require.NotContains(t, content, "sign=")
	require.NotContains(t, content, "empty=")

	// 应该包含其他参数，且按字母排序
	require.Contains(t, content, "app_id=test_app")
	require.Contains(t, content, "method=test.method")
	require.Contains(t, content, "timestamp=2024-01-01 00:00:00")

	// 验证排序
	require.True(t, content[:6] == "app_id") // app_id 在 method 和 timestamp 之前
}

func TestParsePrivateKey_PKCS1(t *testing.T) {
	// 生成 PKCS1 格式的私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)

	parsed, err := parsePrivateKey(privateKeyBase64)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestParsePrivateKey_PKCS8(t *testing.T) {
	// 生成 PKCS8 格式的私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)

	parsed, err := parsePrivateKey(privateKeyBase64)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestParsePrivateKey_WithPEMHeaders(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)

	// 添加 PEM 头尾
	pemKey := "-----BEGIN RSA PRIVATE KEY-----\n" + privateKeyBase64 + "\n-----END RSA PRIVATE KEY-----"

	parsed, err := parsePrivateKey(pemKey)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestParsePublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	parsed, err := parsePublicKey(publicKeyBase64)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestParsePublicKey_WithPEMHeaders(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	// 添加 PEM 头尾
	pemKey := "-----BEGIN PUBLIC KEY-----\n" + publicKeyBase64 + "\n-----END PUBLIC KEY-----"

	parsed, err := parsePublicKey(pemKey)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestBillRecord_Model(t *testing.T) {
	record := BillRecord{
		TransLogID:     "20240101001",
		TransType:      "收款",
		TransTime:      "2024-01-01 12:00:00",
		Amount:         10.5,
		Balance:        100.0,
		Memo:           "转账备注",
		OtherAccount:   "user@example.com",
		TransDirection: "in",
	}

	require.Equal(t, "20240101001", record.TransLogID)
	require.Equal(t, "收款", record.TransType)
	require.Equal(t, 10.5, record.Amount)
	require.Equal(t, "in", record.TransDirection)
}

func TestClient_IsConfigured(t *testing.T) {
	privateKeyBase64, _ := generateTestKeyPair(t)

	t.Run("已配置", func(t *testing.T) {
		cfg := config.AlipayConfig{
			AppID:      "test_app_id",
			PrivateKey: privateKeyBase64,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		require.True(t, client.IsConfigured())
	})

	t.Run("未配置 AppID", func(t *testing.T) {
		cfg := config.AlipayConfig{
			PrivateKey: privateKeyBase64,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		require.False(t, client.IsConfigured())
	})

	t.Run("未配置私钥", func(t *testing.T) {
		cfg := config.AlipayConfig{
			AppID: "test_app_id",
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		require.False(t, client.IsConfigured())
	})
}

func TestClient_GetConfigStatus(t *testing.T) {
	privateKeyBase64, publicKeyBase64 := generateTestKeyPair(t)

	cfg := config.AlipayConfig{
		AppID:           "1234567890123456",
		ServerURL:       "https://openapi.alipay.com/gateway.do",
		PrivateKey:      privateKeyBase64,
		AlipayPublicKey: publicKeyBase64,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	status := client.GetConfigStatus()

	require.True(t, status["configured"].(bool))
	require.True(t, status["app_id_set"].(bool))
	require.True(t, status["private_key_set"].(bool))
	require.True(t, status["public_key_set"].(bool))
	require.Equal(t, "https://openapi.alipay.com/gateway.do", status["server_url"])
	require.Equal(t, "1234****3456", status["app_id_masked"])
}

func TestClient_GetConfigStatus_NotConfigured(t *testing.T) {
	cfg := config.AlipayConfig{}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	status := client.GetConfigStatus()

	require.False(t, status["configured"].(bool))
	require.False(t, status["app_id_set"].(bool))
	require.False(t, status["private_key_set"].(bool))
}
