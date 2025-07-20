package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"

	"claude-relay-core/internal/config"
)

// GenerateCodeVerifier 生成PKCE code verifier
func GenerateCodeVerifier() (string, error) {
	// 生成32字节随机数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成code verifier失败: %w", err)
	}
	
	// base64url编码
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateCodeChallenge 生成PKCE code challenge
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GenerateState 生成随机state参数
func GenerateState() (string, error) {
	// 生成32字节随机数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成state失败: %w", err)
	}
	
	// hex编码
	return hex.EncodeToString(bytes), nil
}

// GenerateAuthURL 生成OAuth授权URL
func GenerateAuthURL(cfg *config.Config, codeChallenge, state string) string {
	params := url.Values{
		"code":                  {"true"},
		"client_id":             {cfg.OAuth.ClientID},
		"response_type":         {"code"},
		"redirect_uri":          {cfg.OAuth.RedirectURI},
		"scope":                 {cfg.OAuth.Scopes},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}

	return fmt.Sprintf("%s?%s", cfg.OAuth.AuthorizeURL, params.Encode())
}

// GenerateOAuthParams 生成完整的OAuth参数
func GenerateOAuthParams(cfg *config.Config) (*PKCEData, error) {
	// 生成code verifier
	codeVerifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("生成code verifier失败: %w", err)
	}

	// 生成code challenge
	codeChallenge := GenerateCodeChallenge(codeVerifier)

	// 生成state
	state, err := GenerateState()
	if err != nil {
		return nil, fmt.Errorf("生成state失败: %w", err)
	}

	// 生成授权URL
	authURL := GenerateAuthURL(cfg, codeChallenge, state)

	return &PKCEData{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		State:         state,
		AuthURL:       authURL,
	}, nil
}