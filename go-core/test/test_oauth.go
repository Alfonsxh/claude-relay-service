package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL = "http://localhost:3000"
)

func main() {
	fmt.Println("🧪 开始测试Claude Relay Service...")

	// 1. 测试健康检查
	fmt.Println("\n1️⃣ 测试健康检查...")
	if err := testHealth(); err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return
	}
	fmt.Println("✅ 健康检查通过")

	// 2. 生成OAuth授权URL
	fmt.Println("\n2️⃣ 生成OAuth授权URL...")
	authData, err := generateAuthURL()
	if err != nil {
		fmt.Printf("❌ 生成授权URL失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 授权URL: %s\n", authData.AuthURL)
	fmt.Printf("📋 State: %s\n", authData.State)

	// 3. 等待用户完成OAuth认证
	fmt.Println("\n3️⃣ 请按以下步骤完成OAuth认证:")
	fmt.Println("   1. 复制上面的授权URL到浏览器中打开")
	fmt.Println("   2. 登录Claude Code账号并授权")
	fmt.Println("   3. 授权完成后，复制地址栏中的authorization code")
	fmt.Println("   4. 在下方输入authorization code:")

	fmt.Print("请输入authorization code: ")
	var authCode string
	fmt.Scanln(&authCode)

	// 4. 交换token
	fmt.Println("\n4️⃣ 交换authorization code获取token...")
	accountName := "test_account"
	if err := exchangeToken(authCode, authData.State, accountName); err != nil {
		fmt.Printf("❌ token交换失败: %v\n", err)
		return
	}
	fmt.Printf("✅ OAuth认证成功，账户名: %s\n", accountName)

	// 5. 检查账户状态
	fmt.Println("\n5️⃣ 检查账户状态...")
	if err := checkAccountStatus(accountName); err != nil {
		fmt.Printf("❌ 检查账户状态失败: %v\n", err)
		return
	}

	// 6. 测试API转发
	fmt.Println("\n6️⃣ 测试API请求转发...")
	if err := testAPIRelay(accountName); err != nil {
		fmt.Printf("❌ API转发测试失败: %v\n", err)
		return
	}
	fmt.Println("✅ API转发测试成功")

	fmt.Println("\n🎉 所有测试完成！Claude Relay Service运行正常。")
}

type AuthData struct {
	AuthURL       string `json:"auth_url"`
	State         string `json:"state"`
	CodeChallenge string `json:"code_challenge"`
}

func testHealth() error {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result["status"] != "healthy" {
		return fmt.Errorf("服务状态不健康: %v", result)
	}

	return nil
}

func generateAuthURL() (*AuthData, error) {
	reqBody := map[string]interface{}{
		// "proxy_config": map[string]interface{}{
		// 	"type": "socks5",
		// 	"host": "127.0.0.1",
		// 	"port": 1080,
		// },
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/oauth/auth-url", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var authData AuthData
	if err := json.NewDecoder(resp.Body).Decode(&authData); err != nil {
		return nil, err
	}

	return &authData, nil
}

func exchangeToken(authCode, state, accountName string) error {
	reqBody := map[string]interface{}{
		"authorization_code": authCode,
		"state":              state,
		"account_name":       accountName,
		"proxy_config": map[string]interface{}{
			"type": "http",
			"host": "127.0.0.1",
			"port": 1097,
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/oauth/token", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Printf("Token交换结果: %v\n", result)
	return nil
}

func checkAccountStatus(accountName string) error {
	resp, err := http.Get(fmt.Sprintf("%s/oauth/accounts/%s/status", baseURL, accountName))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Printf("账户状态: %v\n", result)
	return nil
}

func testAPIRelay(accountName string) error {
	// 构建测试请求
	reqBody := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 100,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Hello! Please respond with 'Test successful' if you can see this message.",
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	// 创建请求
	url := fmt.Sprintf("%s/api/v1/messages?account=%s", baseURL, accountName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Printf("API响应: %v\n", result)
	return nil
}
