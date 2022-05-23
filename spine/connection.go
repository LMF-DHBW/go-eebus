package spine

import (
	"github.com/LMF-DHBW/go-eebus/ressources"
	"github.com/LMF-DHBW/go-eebus/ship"
	"log"
	"time"
)

type Notifier func(ressources.DatagramType, SpineConnection)
type BindSubscribeNotify func(string, *SpineConnection, *ressources.BindSubscribeEntry)

type SpineConnection struct {
	SME                  *ship.SMEInstance
	Address              string
	MsgCounter           int
	OwnDevice            *ressources.DeviceModel
	recieveChan          chan ressources.DatagramType
	DiscoveryInformation *ressources.NodeManagementDetailedDiscovery
	bindSubscribeNotify  BindSubscribeNotify
	bindSubscribeInfo    []*BindSubscribeInfo
	SubscriptionNofity   Notifier
	Features             []*Feature
}

type Feature struct {
	EntityType     string
	FeatureType    string
	FeatureAddress ressources.FeatureAddressType
	FunctionName   string
	CurrentState   string
}

type BindSubscribeInfo struct {
	BindSubscribe      string
	BindSubscribeEntry *ressources.BindSubscribeEntry
}

func NewSpineConnection(SME *ship.SMEInstance, ownDevice *ressources.DeviceModel, bindSubscribeNotify BindSubscribeNotify, SubscriptionNofity Notifier) *SpineConnection {
	return &SpineConnection{SME, "", 0, ownDevice, make(chan ressources.DatagramType), nil, bindSubscribeNotify, nil, SubscriptionNofity, make([]*Feature, 0)}
}

func (conn *SpineConnection) SendXML(header *ressources.HeaderType, payload *ressources.PayloadType) {
	log.Println("Sending: ", header.CmdClassifier)
	conn.SME.Send(ressources.DatagramType{header, payload})
}

func (conn *SpineConnection) StartRecieveHandler() {
	log.Println("Recieving")
	conn.SME.Recieve(func(datagram ressources.DatagramType) {
		entitiyAddr := datagram.Header.AddressDestination.Entity
		featureAddr := datagram.Header.AddressDestination.Feature
		deviceSource := datagram.Header.AddressSource.Device
		entitiySource := datagram.Header.AddressSource.Entity
		featureSource := datagram.Header.AddressSource.Feature
		isValidRequest := len(conn.OwnDevice.Entities) > entitiyAddr && len(conn.OwnDevice.Entities[entitiyAddr].Features) > featureAddr
		if isValidRequest {
			log.Println("Received: ", datagram.Header.CmdClassifier)
			feature := conn.OwnDevice.Entities[entitiyAddr].Features[featureAddr]
			var function *ressources.FunctionModel
			for _, v := range feature.Functions {
				if v.FunctionName == datagram.Payload.Cmd.FunctionName {
					function = v
					break
				}
			}
			switch datagram.Header.CmdClassifier {
			case "reply", "result":
				conn.recieveChan <- datagram
			case "read":
				if conn.requestAllowed("binding", datagram.Header) {
					conn.SendXML(
						conn.OwnDevice.MakeHeader(entitiyAddr, featureAddr, ressources.MakeFeatureAddress(deviceSource, entitiySource, featureSource), "result", conn.MsgCounter, false),
						ressources.MakePayload(function.FunctionName, function.Function))
				}
			case "write":
				if conn.requestAllowed("binding", datagram.Header) {
					function.ChangeNotify(datagram.Payload.Cmd.FunctionName, datagram.Payload.Cmd.Function, *datagram.Header.AddressDestination)
				}
			case "notify":
				if conn.requestAllowed("subscription", datagram.Header) {
					if len(conn.DiscoveryInformation.FeatureInformation) > featureSource && len(conn.DiscoveryInformation.EntityInformation) > entitiySource {
						featureInList := false
						for i := range conn.Features {
							if conn.Features[i].FeatureAddress == *datagram.Header.AddressSource && conn.Features[i].FunctionName == datagram.Payload.Cmd.FunctionName {
								featureInList = true
								break
							}
						}
						if !featureInList {
							conn.Features = append(conn.Features, &Feature{
								EntityType:     conn.DiscoveryInformation.EntityInformation[entitiySource].Description.EntityType,
								FeatureType:    conn.DiscoveryInformation.FeatureInformation[featureSource].Description.FeatureType,
								FeatureAddress: *datagram.Header.AddressSource,
								FunctionName:   datagram.Payload.Cmd.FunctionName,
								CurrentState:   datagram.Payload.Cmd.Function,
							})
						} else {
							for i := range conn.Features {
								if conn.Features[i].FeatureAddress == *datagram.Header.AddressSource && conn.Features[i].FunctionName == datagram.Payload.Cmd.FunctionName {
									conn.Features[i].CurrentState = datagram.Payload.Cmd.Function
									break
								}
							}
						}
						conn.SubscriptionNofity(datagram, *conn)
					}
				}
			case "call":
				if datagram.Payload.Cmd.FunctionName == "nodeManagementBindingRequestCall" {
					conn.processBindingRequest(&datagram)
				} else if datagram.Payload.Cmd.FunctionName == "nodeManagementSubscriptionRequestCall" {
					conn.processSubscriptionRequest(&datagram)
				}
			}
		}
	})
}

func (conn *SpineConnection) requestAllowed(bindSubscribe string, header *ressources.HeaderType) bool {
	entitiyAddr := header.AddressDestination.Entity
	featureAddr := header.AddressDestination.Feature
	entitiySource := header.AddressSource.Entity
	featureSource := header.AddressSource.Feature
	if entitiyAddr == 0 && featureAddr == 0 {
		return true
	}
	for _, info := range conn.bindSubscribeInfo {
		if info.BindSubscribe == bindSubscribe &&
			entitiyAddr == info.BindSubscribeEntry.ServerAddress.Entity && featureAddr == info.BindSubscribeEntry.ServerAddress.Feature &&
			entitiySource == info.BindSubscribeEntry.ClientAddress.Entity && featureSource == info.BindSubscribeEntry.ClientAddress.Feature {
			return true
		}
	}
	return false
}

func (conn *SpineConnection) RecieveTimeout(seconds int) (ressources.DatagramType, bool) {
	var res ressources.DatagramType
	err := false
	select {
	case res = <-conn.recieveChan:
		err = true
	case <-time.After(time.Duration(seconds) * time.Second):
		err = false
	}
	return res, err
}

func (conn *SpineConnection) CountBindings(serverAddr ressources.FeatureAddressType) int {
	numBindings := 0
	for _, bindSub := range conn.bindSubscribeInfo {
		if bindSub.BindSubscribe == "binding" && bindSub.BindSubscribeEntry.ServerAddress == serverAddr {
			numBindings++
		}
	}
	return numBindings
}

func (conn *SpineConnection) CountSubscriptions(serverAddr ressources.FeatureAddressType) int {
	numSubscriptions := 0
	for _, bindSub := range conn.bindSubscribeInfo {
		if bindSub.BindSubscribe == "subscription" && bindSub.BindSubscribeEntry.ServerAddress == serverAddr {
			numSubscriptions++
		}
	}
	return numSubscriptions
}
