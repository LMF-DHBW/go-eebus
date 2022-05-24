package ship

import (
	"context"
	"log"
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