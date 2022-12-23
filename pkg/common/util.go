package common

import (
	"crypto/ed25519"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"strings"

	"github.com/handewo/gojump/pkg/log"
	"golang.org/x/crypto/bcrypt"

	uuid "github.com/satori/go.uuid"
)

func MakeSignature(key, date string) string {
	s := strings.Join([]string{key, date}, "\n")
	return Base64Encode(MD5Encode([]byte(s)))
}

func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func MD5Encode(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}

func IgnoreErrWriteString(writer io.Writer, s string) {
	_, _ = io.WriteString(writer, s)
}

const (
	ColorEscape = "\033["
	Green       = "32m"
	Red         = "31m"
	ColorEnd    = ColorEscape + "0m"
	Bold        = "1"
)

const (
	CharClear     = "\x1b[H\x1b[2J"
	CharTab       = "\t"
	CharNewLine   = "\r\n"
	CharCleanLine = '\x15'
)

func WrapperString(text string, color string, meta ...bool) string {
	wrapWith := make([]string, 0)
	metaLen := len(meta)
	switch metaLen {
	case 1:
		wrapWith = append(wrapWith, Bold)
	}
	wrapWith = append(wrapWith, color)
	return fmt.Sprintf("%s%s%s%s", ColorEscape, strings.Join(wrapWith, ";"), text, ColorEnd)
}

func WrapperTitle(text string) string {
	return WrapperString(text, Green, true)
}

func WrapperWarn(text string) string {
	text += "\n\r"
	return WrapperString(text, Red)
}

func IgnoreErrWriteWindowTitle(writer io.Writer, title string) {
	// OSC Ps ; Pt BEL
	// OSC Ps ; Pt ST
	// Ps = 2  ⇒  Change Window Title to Pt.
	_, _ = writer.Write([]byte(fmt.Sprintf("\x1b]2;%s\x07", title)))
}

func LongestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	isCommonPrefix := func(length int) bool {
		str0, count := strs[0][:length], len(strs)
		for i := 1; i < count; i++ {
			if strs[i][:length] != str0 {
				return false
			}
		}
		return true
	}

	minLength := len(strs[0])
	for _, s := range strs {
		if len(s) < minLength {
			minLength = len(s)
		}

	}

	low, high := 0, minLength
	for low < high {
		mid := (high-low+1)/2 + low
		if isCommonPrefix(mid) {
			low = mid
		} else {
			high = mid - 1
		}

	}
	return strs[0][:low]
}

func FilterPrefix(strs []string, s string) (r []string) {
	for _, v := range strs {
		if len(v) >= len(s) {
			if v[:len(s)] == s {
				r = append(r, v)
			}
		}
	}

	return r
}

func LongestStr(strs []string) string {
	longestStr := ""
	for _, str := range strs {
		if len(str) >= len(longestStr) {
			longestStr = str
		}
	}

	return longestStr
}

func Pretty(strs []string, width int) (s string) {
	longestStr := LongestStr(strs)
	length := len(longestStr) + 4
	lineCount := width / length

	for index, str := range strs {
		if index == 0 {
			s += fmt.Sprintf(fmt.Sprintf("%%-%ds", length), str)
		} else {
			if index%lineCount == 0 {
				s += fmt.Sprintf(fmt.Sprintf("\n%%-%ds", length), str)
			} else {
				s += fmt.Sprintf(fmt.Sprintf("%%-%ds", length), str)
			}
		}
	}

	return s
}

func UUID() string {
	return uuid.NewV4().String()
}

func ValidUUIDString(sid string) bool {
	_, err := uuid.FromString(sid)
	return err == nil
}

func Sum(i []int) int {
	sum := 0
	for _, v := range i {
		sum += v
	}
	return sum
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func GenerateRSAPem() string {
	bitSize := 2048

	privateKey, err := generateRSAPrivateKey(bitSize)
	if err != nil {
		log.Fatal.Print(err.Error())
	}

	privateKeyBytes := encodeRSAPrivateKeyToPEM(privateKey)
	return string(privateKeyBytes)
}

func GenerateEd25519Pem() string {
	// Generate a new private/public keypair for OpenSSH
	_, privKey, _ := ed25519.GenerateKey(rand.Reader)
	pemKey := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: MarshalED25519PrivateKey(privKey),
	}
	privateKey := pem.EncodeToMemory(pemKey)
	return string(privateKey)
}

// generateRSAPrivateKey creates a RSA Private Key of specified byte size
func generateRSAPrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// encodeRSAPrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodeRSAPrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
// func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
// 	publicRsaKey, err := ssh.NewPublicKey(privatekey)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)
//
// 	return pubKeyBytes, nil
// }

// Hash password using the bcrypt hashing algorithm
func HashPassword(password string) (string, error) {
	// Convert password string to byte slice
	var passwordBytes = []byte(password)

	// Hash password with bcrypt's min cost
	hashedPasswordBytes, err := bcrypt.
		GenerateFromPassword(passwordBytes, bcrypt.MinCost)

	return string(hashedPasswordBytes), err
}
