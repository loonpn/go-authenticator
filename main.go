package main

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

const (
	length       = 6                  // Password length
	timeout      = 30 * time.Second   // Timeout
	algorithm    = otp.AlgorithmSHA1  // Encryption algorithm
)

var (
	emergencyFile  string           // The path where the emergency code is saved
	emergencyCodes []string         // Emergency code
	secretKeyLen = 16               // Key length
)

func main() {
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error: Unable to get user information")
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknowhost"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error: Unable to get user directory")
		os.Exit(-3)
	} else {
		emergencyFile = home + "/.authenticator"
	}
	accountName := currentUser.Username + "@" + hostname

	issuer := hostname
	if len(os.Args) == 2 {
		issuer = os.Args[1]
	}
	if isFirstRun(secretKeyLen) {
		// Generate a key randomly
		secretKey, err := randBase32String(length *2)
		if err == nil {
			secretKey = secretKey[:secretKeyLen]
		}
		url, err := showQRCode(accountName, issuer, secretKey)
		if err != nil {
			fmt.Println("Please generate QR code and scan:", url)
		} else {
			fmt.Println("Please scan the QR code using a password verifier app")
		}
		fmt.Printf("Please enter a random password: ")
		code := readCode()
		if verifyCode(code, secretKey) {
			err = generateEmergencyCodes(6) // Randomly generate 6 emergency codes
			if err != nil {
				fmt.Println("Error: Unable to generate emergency code")
				os.Exit(-1)
			}
			err = saveToFile(secretKey)
			if err != nil {
				fmt.Println("Error: Unable to save file", emergencyFile)
				os.Exit(-2)
			}
			fmt.Println("Validation successful, your key is:", secretKey, "\nPlease save the following emergency code:")
			showEmergencyCodes()
			os.Exit(0)
		}
		fmt.Println("Validation Failed")
	} else {
		secretKey, err := readFromFile(secretKeyLen)
		if err != nil {
			fmt.Println("Error: Cannot read file properly", emergencyFile)
			os.Exit(-3)
		}
		fmt.Printf("Please enter a random password: ")
		code := readCodeWithTimeout()
		if verifyCode(code, secretKey) {
			os.Exit(0)
		}
		if code != "" {
			fmt.Println("Validation Failed")
		} else {
			fmt.Println()
		}
	}
	os.Exit(1)
}

// Determine whether it is run for the first time
func isFirstRun(secretKeyLen int) bool {
	if f, err := os.Stat(emergencyFile); err == nil {
		if f.IsDir() || f.Size() <= int64(secretKeyLen) || f.Size() > 1024 {
			// The path is a directory, an empty file, or a file size that exceeds 1024 bytes
			os.RemoveAll(emergencyFile)
		} else {
			return false
		}
	}
	return true
}

// Display the QR code character pattern
func showQRCode(accountName, issuer, secretKey string) (string, error) {
	url := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s&algorithm=%s&digits=%d", accountName, secretKey, issuer, algorithm, length)
	_, err := totp.GenerateCodeCustom(secretKey, time.Now(), totp.ValidateOpts{
		Digits:    length,
		Algorithm: algorithm,
	})
	if err != nil {
		return url, err
	}
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return url, err
	}
	fmt.Println(qr.ToSmallString(false))
	return url, nil
}

// Read the password entered by the user
func readCode() string {
	var code string
	fmt.Scanln(&code)
	return code
}

// Read the password entered by the user with a timeout function
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

// Verify that the password entered by the user is correct
func verifyCode(code, secretKey string) bool {
	if !isDigitOrUpper(code) {
		return false
	}
	if len(code) == length {
		valid, err := totp.ValidateCustom(code, secretKey, time.Now(), totp.ValidateOpts{
			Digits:    length,
			Algorithm: algorithm,
		})
		if err != nil {
			return false
		}
		return valid
	} else if len(code) == length * 2 {
		for i, c := range emergencyCodes {
			if code == c {
				emergencyCodes[i] = ""
				saveToFile(secretKey)
				return true
			}
		}
	}
	return false
}

// Randomly generate Base32 strings
func randBase32String(n int) (string, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return base32.StdEncoding.EncodeToString(b), nil
}

// Generate emergency codes randomly
func generateEmergencyCodes(n int) error {
	emergencyCodes = make([]string, n)
	for i := n; i < 2 * n; i++ {
		code, err := randBase32String(length * 2)
		if err != nil {
			return err
		}
		emergencyCodes[i - n] = code[:length *2]
	}
	return nil
}

// Save the emergency code to a file
func saveToFile(secretKey string) error {
	str := "# Your secret key is:\n" + secretKey + "\n# Your emergency codes are:\n"
	for _, i := range emergencyCodes {
		if i != "" {
			str = str + i + "\n"
		}
	}
	os.Remove(emergencyFile)
	f,err := os.Create(emergencyFile)
	defer f.Close()
	if err !=nil {
		return err
	}
	_,err=f.Write([]byte(str))
	if err != nil{
		return err
	}
	return nil
}

// Read the file
func readFromFile(secretKeyLen int) (string, error) {
	content, err := ioutil.ReadFile(emergencyFile)
	if err != nil {
		return "", err
	}
	secretKey := ""
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		linelen := len(line)
		if linelen > 0 && int(line[0]) == 35 || linelen == 0 {
			// The line is either a comment line or a blank line
			continue
		}
		if i < 32 && isDigitOrUpper(line) {
			// The line is a combination of numbers and uppercase letters
			if secretKey == "" && linelen == secretKeyLen {
				// The line is the key
				secretKey = line
			} else if secretKey != "" && linelen == length * 2 {
				// This line is the emergency code and follows the key line
				emergencyCodes = append(emergencyCodes, line)
			} else {
				// Illegal format: The key line is in the middle or after the emergency code line; Multiple keys
				secretKey = ""
				break
			}
		} else {
			// Illegal format: The string is not a comment or a combination of numbers and uppercase letters; The number of rows exceeds the specified range
			secretKey = ""
			break
		}
	}
	if secretKey != "" {
		return secretKey, nil
	} else {
		return "", fmt.Errorf("Illegal format");
	}
}

// Determines whether a string is a combination of numbers and uppercase letters
func isDigitOrUpper(str string) bool {
	// Iterate through each byte of the string
	for i := 0; i < len(str); i++ {
		// Gets the ASCII code corresponding to the byte
		ascii := int(str[i])
		// Determine if the ASCII code is not in the range of 0-9 and A-Z
		if ascii < 48 || ascii > 57 && ascii < 65 || ascii > 90 {
			// If not, return false
			return false
		}
	}
	return true
}

// Displays the emergency code
func showEmergencyCodes() {
	for _, c := range emergencyCodes {
		fmt.Println(c)
	}
}
