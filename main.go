package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

const (
	
	secretKey     = "IFBEGRCFIZDTCMRT"         // 密钥,BASE32编码
	timeout       = 30 * time.Second           // 超时时间
	emergencyFile = "/etc/emergencycode" // 紧急密码保存路径
)

var (
	accountName     = ""        // 账号名称
	//path          = ""        // 程序路径
	//arguments     = ""        // 程序运行参数
	emergencyCodes []string     // 紧急密码
)

func main() {
	if isRunning() {
		fmt.Println("程序已经在运行，请勿重复启动")
		os.Exit(1)
	}
	currentUser, err := user.Current()
	if err != nil {
	   panic(err)
	}
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	accountName = currentUser.Username + "@" + name

	if isFirstRun() {
		fmt.Println("首次运行程序，请扫描以下二维码并保存账号和密钥")
		showQRCode()
		fmt.Printf("请输入随机密码: ")
		code := readCode()
		if verifyCode(code) {
			fmt.Println("验证成功，以下是紧急密码，请妥善保存")
			generateEmergencyCodes(6) // 随机生成6个紧急密码
			saveToFile()
			showEmergencyCodes()
			os.Exit(0)
		} else {
			fmt.Println("验证失败")
		}
	} else {
		readFromFile()
		fmt.Printf("请在30秒内输入临时密码: ")
		code := readCodeWithTimeout()
		if verifyCode(code) {
			fmt.Println("验证成功")
			/*for k, v := range os.Args {
				if k == 1 {
					path = v
				} else if k > 1 {
					if arguments == "" {
						arguments = v
					} else {
						arguments = arguments + " " + v
					}
				}
			}*/
			//runPathProgram()
			os.Exit(0)
		} else {
			if code != "" {
				fmt.Println("验证失败")
			}
		}
	}
	os.Exit(1)
}

// 判断是否首次运行
func isFirstRun() bool {
	if _, err := os.Stat(emergencyFile); err == nil {
		return false
	}
	return true
}

// 判断是否已经在运行
func isRunning() bool {
	cmd := exec.Command("pgrep", "-f", os.Args[0])
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	count := len(output) / 5 // 每个进程ID占5个字节
	return count > 1         // 如果大于1说明有多个进程实例
}

// 显示二维码字符图案
func showQRCode() {
	key, err := totp.GenerateCodeCustom(secretKey, time.Now(), totp.ValidateOpts{
		Digits:    6,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		panic(err)
	}
	url := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s", accountName, secretKey, key)
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		panic(err)
	}
	fmt.Println(qr.ToSmallString(false))
}

// 读取用户输入的密码
func readCode() string {
	var code string
	fmt.Scanln(&code)
	return code
}

// 读取用户输入的密码，带超时功能
func readCodeWithTimeout() string {
	ch := make(chan string)
	go func() {
		var code string
		fmt.Scanln(&code)
		ch <- code
	}()
	select {
	case code := <-ch:
		return code
	case <-time.After(timeout):
		return ""
	}
}

// 验证用户输入的密码是否正确
func verifyCode(code string) bool {
	if len(code) != 6 && len(code) != 12 || !isDigit(code) {
		return false
	}
	for i, c := range emergencyCodes {
		if code == c {
			emergencyCodes[i] = ""
			saveToFile()
			return true
		}
	}
	valid, err := totp.ValidateCustom(code, secretKey, time.Now(), totp.ValidateOpts{
		Digits:    6,
        Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false
	}
	return valid
}

// 随机生成紧急密码
func generateEmergencyCodes(n int) {
	emergencyCodes = make([]string, n)
	for i := 0; i < n; i++ {
		d := time.Duration(-i-1) * time.Minute
		code, err := totp.GenerateCodeCustom(secretKey + secretKey, time.Now().Add(d), totp.ValidateOpts{
			Digits:    12,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err != nil {
			panic(err)
		}
		emergencyCodes[i] = code
	}
}

// 保存紧急代码到文件
func saveToFile() {
	str := ""
	for _, i := range emergencyCodes {
		if i != "" {
			str = str + i + "\n"
		}
	}
	os.Remove(emergencyFile)
	f,err := os.Create(emergencyFile)
	defer f.Close()
	if err !=nil {
		panic(err)
	}
	_,err=f.Write([]byte(str))
	if err != nil{
		panic(err)
	}
}

// 读取文件
func readFromFile() {
	content, err := ioutil.ReadFile(emergencyFile)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if len(line) == 12 && isDigit(line) {
			emergencyCodes = append(emergencyCodes, line)
		}
	}
}

// 判断字符串是否为数字
func isDigit(str string) bool {
	// 遍历字符串的每个字节
	for i := 0; i < len(str); i++ {
		// 获取字节对应的ASCII码
		ascii := int(str[i])
		// 判断ASCII码是否在0-9的范围内
		if ascii < 48 || ascii > 57 {
			// 如果不是，返回false
			return false
		}
	}
	// 如果都是，返回true
	return true
}

// 显示紧急密码
func showEmergencyCodes() {
	for _, c := range emergencyCodes {
		fmt.Println(c)
	}
}
/*
// 执行指定的路径程序
func runPathProgram() {
	if path == "" {
		return
	}
	cmd := exec.Command(path, arguments)
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
}
*/
