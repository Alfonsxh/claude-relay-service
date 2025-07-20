package oauth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Storage OAuth数据存储
type Storage struct {
	dataDir string
}

// NewStorage 创建存储实例
func NewStorage(dataDir string) *Storage {
	return &Storage{
		dataDir: dataDir,
	}
}

// SaveOAuthData 保存OAuth数据到文件
func (s *Storage) SaveOAuthData(accountName string, data *OAuthData) error {
	// 确保数据目录存在
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 序列化数据
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化OAuth数据失败: %w", err)
	}

	// 写入文件
	filename := filepath.Join(s.dataDir, fmt.Sprintf("oauth_%s.json", accountName))
	if err := os.WriteFile(filename, jsonData, 0600); err != nil {
		return fmt.Errorf("写入OAuth数据文件失败: %w", err)
	}

	fmt.Printf("✅ OAuth数据已保存: %s\n", filename)
	return nil
}

// LoadOAuthData 从文件加载OAuth数据
func (s *Storage) LoadOAuthData(accountName string) (*OAuthData, error) {
	filename := filepath.Join(s.dataDir, fmt.Sprintf("oauth_%s.json", accountName))
	
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("OAuth数据文件不存在: %s", filename)
	}

	// 读取文件
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取OAuth数据文件失败: %w", err)
	}

	// 反序列化数据
	var data OAuthData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("反序列化OAuth数据失败: %w", err)
	}

	return &data, nil
}

// SavePKCEData 保存PKCE数据到临时文件
func (s *Storage) SavePKCEData(state string, data *PKCEData) error {
	// 确保数据目录存在
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 序列化数据
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化PKCE数据失败: %w", err)
	}

	// 写入临时文件
	filename := filepath.Join(s.dataDir, fmt.Sprintf("pkce_%s.json", state))
	if err := os.WriteFile(filename, jsonData, 0600); err != nil {
		return fmt.Errorf("写入PKCE数据文件失败: %w", err)
	}

	return nil
}

// LoadPKCEData 从文件加载PKCE数据
func (s *Storage) LoadPKCEData(state string) (*PKCEData, error) {
	filename := filepath.Join(s.dataDir, fmt.Sprintf("pkce_%s.json", state))
	
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("PKCE数据文件不存在: %s", filename)
	}

	// 读取文件
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取PKCE数据文件失败: %w", err)
	}

	// 反序列化数据
	var data PKCEData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("反序列化PKCE数据失败: %w", err)
	}

	return &data, nil
}

// DeletePKCEData 删除PKCE临时数据
func (s *Storage) DeletePKCEData(state string) error {
	filename := filepath.Join(s.dataDir, fmt.Sprintf("pkce_%s.json", state))
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除PKCE数据文件失败: %w", err)
	}
	return nil
}

// ListAccounts 列出所有已存储的账户
func (s *Storage) ListAccounts() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(s.dataDir, "oauth_*.json"))
	if err != nil {
		return nil, fmt.Errorf("扫描OAuth数据文件失败: %w", err)
	}

	var accounts []string
	for _, file := range files {
		basename := filepath.Base(file)
		// 提取账户名：oauth_accountname.json -> accountname
		if len(basename) > 12 && basename[:6] == "oauth_" && basename[len(basename)-5:] == ".json" {
			accountName := basename[6 : len(basename)-5]
			accounts = append(accounts, accountName)
		}
	}

	return accounts, nil
}