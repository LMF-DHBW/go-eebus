package ship

import (
	"bufio"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"strings"

	"github.com/LMF-DHBW/go-eebus/ressources"
)

func ReadSkis() ([]string, []string) {
	// Create file if not exists
	if _, err := os.Stat("skis.txt"); errors.Is(err, os.ErrNotExist) {
		return make([]string, 0), make([]string, 0)
	}

	file, err := os.Open("skis.txt")
	ressources.CheckError(err)
	defer file.Close()

	var skis []string
	var devices []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ",")
		skis = append(skis, line[0])
		devices = append(devices, line[1])
	}
	ressources.CheckError(scanner.Err())
	return skis, devices
}

func WriteSkis(newSkis []string, newDevices []string) {
	file, err := os.OpenFile("skis.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	ressources.CheckError(err)
	defer file.Close()
	if len(newSkis) == len(newDevices) {
		result := ""

		for i := 0; i < len(newSkis); i++ {
			if newSkis[i] != "" {
				result += newSkis[i] + "," + newDevices[i]
				if i < (len(newSkis) - 1) {
					result += "\n"
				}
			}
		}

		_, err = file.WriteString(result)
		ressources.CheckError(err)
	}

}

func (shipNode *ShipNode) getSki() string {
	var file []byte
	var err error

	file, err = os.ReadFile(shipNode.CertName + ".crt")
	ressources.CheckError(err)

	crt := string(file)

	block, _ := pem.Decode([]byte(crt))
	var cert *x509.Certificate
	cert, _ = x509.ParseCertificate(block.Bytes)
	pubkey := cert.PublicKey.(*rsa.PublicKey)

	publicKey, err := x509.MarshalPKIXPublicKey(pubkey)
	ressources.CheckError(err)

	hasher := sha1.New()
	hasher.Write(publicKey)
	return hex.EncodeToString(hasher.Sum(nil))
}
