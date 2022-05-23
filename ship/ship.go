package ship

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/LMF-DHBW/go-eebus/ressources"

	"github.com/grandcat/zeroconf"
	"github.com/phayes/freeport"
	"golang.org/x/net/websocket"
)

type ConnectionManager func(string, *websocket.Conn)
type ConnectionManagerSpine func(*SMEInstance)
type CloseHandler func(*SMEInstance)
type handler func([]byte)
type dataHandler func(ressources.DatagramType)

type ShipNode struct {
	serverPort            int
	isGateway             bool
	SME                   []*SMEInstance
	Requests              []*Request
	SpineConnectionNotify ConnectionManagerSpine
	SpineCloseHandler     CloseHandler
	CertName              string
}

type Request struct {
	Port string
	Id   string
	Ski  string
}

func NewShipNode(isGateway bool, certName string) *ShipNode {
	// Empty Ship node has empty list of clients and no server
	return &ShipNode{0, isGateway, make([]*SMEInstance, 0), make([]*Request, 0), nil, nil, certName}
}

func (shipNode *ShipNode) Start() {
	// ShipNode start -> assign port, create server
	port, err := freeport.GetFreePort()
	ressources.CheckError(err)
	shipNode.serverPort = port
	// Start server, Register Dns and search for other DNS entries
	if !shipNode.isGateway {
		go shipNode.StartServer()
		go shipNode.RegisterDns()

	}
	go shipNode.BrowseDns()
}

func (shipNode *ShipNode) handleFoundService(entry *zeroconf.ServiceEntry) {
	// If found service is not on same device
	if entry.Port != shipNode.serverPort {
		log.Println("Found new service", entry.HostName, entry.Port)
		//shipNode.Connect("localhost", strconv.Itoa(entry.Port))

		if shipNode.isGateway {
			shipNode.Requests = append(shipNode.Requests, &Request{
				Port: strconv.Itoa(entry.Port),
				Id:   strings.Split(entry.Text[1], "=")[1],
				Ski:  strings.Split(entry.Text[3], "=")[1],
			})
		} else {
			if ressources.StringInSlice(strings.Split(entry.Text[3], "=")[1], readSkis()) {
				// Device is trusted
				go shipNode.Connect("localhost", strconv.Itoa(entry.Port), strings.Split(entry.Text[3], "=")[1])
			}
		}
	}
}

/* Procedure for new conncetions
1. Create SME instance and append to list from SHIP node
2. Start CME handshake
3. Start Hello handshake -> Maybe skip for now, skip protocol handshake and pin exchange
4. Start data exchange -> notify spine
*/
func (shipNode *ShipNode) newConnection(role string, conn *websocket.Conn, ski string) {
	newSME := &SMEInstance{role, "INIT", conn, shipNode.SpineCloseHandler, ski}

	for _, e := range shipNode.SME {
		if e.Ski == ski {
			conn.Close()
			return
		}
	}

	shipNode.SME = append(shipNode.SME, newSME)

	newSME.StartCMI()
	shipNode.SpineConnectionNotify(newSME)
}

func (shipNode *ShipNode) Connect(host string, port string, ski string) {
	service := "wss://" + host + ":" + port

	conf, err := websocket.NewConfig(service, "http://localhosts")
	ressources.CheckError(err)

	// cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair(shipNode.CertName+".crt", shipNode.CertName+".key")
	ressources.CheckError(err)

	conf.TlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	conn, err := websocket.DialConfig(conf)
	ressources.CheckError(err)

	// publickey := conn.Request().TLS.PeerCertificates[0].RawSubjectPublicKeyInfo

	// hasher := sha1.New()
	// hasher.Write(publickey)
	// shipNode.newConnection("client", conn, hex.EncodeToString(hasher.Sum(nil)))
	shipNode.newConnection("client", conn, ski)
}

func (shipNode *ShipNode) StartServer() {

	server := &http.Server{
		Addr: ":" + strconv.Itoa(shipNode.serverPort),
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequestClientCert,
		},
		Handler: websocket.Handler(func(ws *websocket.Conn) {
			publickey := ws.Request().TLS.PeerCertificates[0].RawSubjectPublicKeyInfo

			hasher := sha1.New()
			hasher.Write(publickey)
			shipNode.newConnection("server", ws, hex.EncodeToString(hasher.Sum(nil)))
		}),
	}
	err := server.ListenAndServeTLS(shipNode.CertName+".crt", shipNode.CertName+".key")
	ressources.CheckError(err)
}
