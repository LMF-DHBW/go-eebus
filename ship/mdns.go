package ship

import (
	"bufio"
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/LMF-DHBW/go-eebus/ressources"

	"github.com/grandcat/zeroconf"
)

func (shipNode *ShipNode) BrowseDns() {
	log.Println("Browsing for entries")
	// Discover all ship services on the network
	resolver, err := zeroconf.NewResolver(nil)
	ressources.CheckError(err)

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			go shipNode.handleFoundService(entry)
		}
	}(entries)

	ctx, _ := context.WithCancel(context.Background())

	err = resolver.Browse(ctx, "_ship._tcp", "local.", entries)
	ressources.CheckError(err)

	<-ctx.Done()
}

func (shipNode *ShipNode) RegisterDns() {
	// Define values for DNS entry
	port := strconv.Itoa(shipNode.serverPort)
	id := "DEVICE-EEB01-0001.local."
	ski := shipNode.getSki()

	txtRecord := []string{"txtvers=1", "id=" + id, "path=wss://localhost:" + port, "SKI=" + ski, "register=true"}
	log.Println("Registering: ", txtRecord)
	server, err := zeroconf.Register("Device "+port, "_ship._tcp", "local.", shipNode.serverPort, txtRecord, nil)
	ressources.CheckError(err)

	defer server.Shutdown()
	defer log.Println("Registering stopped")
	// Shutdown server after 2 minutes
	<-time.After(time.Second * 120)
}

/************ Comminssioning ************/

func readSkis() []string {
	file, err := os.Open("skis.txt")
	ressources.CheckError(err)
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	ressources.CheckError(scanner.Err())
	return lines
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
