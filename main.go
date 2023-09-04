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
	
	secretKey     = "IFBEGRCFIZDTCMRT"    // BASE32 encoding Key
	timeout       = 30 * time.Second      // Timeout
)

var (
	accountName     = ""
	//path          = ""
	//arguments     = ""
	emergencyCodes []string
	emergencyFile = "emergencycode"  // Emergency code save path
)

func main() {
	if isRunning() {
		fmt.Println("The program is already running")
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
	emergencyFile = currentUser.HomeDir + "/" + emergencyFile

	if isFirstRun() {
		fmt.Println("This is your first time to run this program. Please scan the following QR code and save your key")
		showQRCode()
		fmt.Printf("Please enter verification code: ")
		code := readCode()
		if verifyCode(code) {
			fmt.Println("Your emergency scratch codes are:")
			generateEmergencyCodes(6)
			saveToFile()
			showEmergencyCodes()
			os.Exit(0)
		} else {
			fmt.Println("Authenticate failed")
		}
	} else {
		readFromFile()
		fmt.Printf("Please enter verification code: ")
		code := readCodeWithTimeout()
		if verifyCode(code) {
			fmt.Println("Authenticate successfully")
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
				fmt.Println("Authenticate failed")
			}
		}
	}
	os.Exit(1)
}

func isFirstRun() bool {
	if _, err := os.Stat(emergencyFile); err == nil {
		return false
	}
	return true
}

func isRunning() bool {
	cmd := exec.Command("pgrep", "-f", os.Args[0])
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	count := len(output) / 5 // Each process ID occupies 5 bytes
	return count > 1         // If it is greater than 1, it indicates that the program is running
}

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

func readCode() string {
	var code string
	fmt.Scanln(&code)
	return code
}

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

func isDigit(str string) bool {
	for i := 0; i < len(str); i++ {
		ascii := int(str[i])
		if ascii < 48 || ascii > 57 {
			return false
		}
	}
	return true
}

func showEmergencyCodes() {
	for _, c := range emergencyCodes {
		fmt.Println(c)
	}
}
/*
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
