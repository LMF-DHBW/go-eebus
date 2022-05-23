package spine

import (
	"encoding/xml"
	"github.com/LMF-DHBW/go-eebus/ressources"
	"log"
)

func (conn *SpineConnection) sendSubscriptionRequest(EntityAddress int, FeatureAddress int, DestinationAddr *ressources.FeatureAddressType, FeatureType string) {
	clientAddr := ressources.MakeFeatureAddress(conn.OwnDevice.DeviceAddress, EntityAddress, FeatureAddress)
	conn.SendXML(
		conn.OwnDevice.MakeHeader(EntityAddress, FeatureAddress, DestinationAddr, "call", conn.MsgCounter, true),
		ressources.MakePayload("nodeManagementSubscriptionRequestCall", &ressources.NodeManagementSubscriptionRequestCall{
			&ressources.SubscriptionManagementRequestCallType{
				ClientAddress:     clientAddr,
				ServerAddress:     DestinationAddr,
				ServerFeatureType: FeatureType,
			},
		}))
	answer, ok := conn.RecieveTimeout(10)
	if ok {
		var Function *ressources.ResultElement
		err := xml.Unmarshal([]byte(answer.Payload.Cmd.Function), &Function)
		if err == nil {
			if Function.ErrorNumber == 0 {
				log.Println("Accepted subscription to: ", DestinationAddr.Device)

				newEntry := &ressources.BindSubscribeEntry{
					ClientAddress: *clientAddr,
					ServerAddress: *DestinationAddr,
				}

				conn.bindSubscribeInfo = append(conn.bindSubscribeInfo, &BindSubscribeInfo{
					"subscription", newEntry,
				})
				conn.bindSubscribeNotify("subscription", conn, newEntry)
			}
		}
	}
}

func (conn *SpineConnection) processSubscriptionRequest(datagram *ressources.DatagramType) {
	var Function *ressources.NodeManagementSubscriptionRequestCall
	err := xml.Unmarshal([]byte(datagram.Payload.Cmd.Function), &Function)

	entitiyAddr := Function.SubscriptionRequest.ServerAddress.Entity
	featureAddr := Function.SubscriptionRequest.ServerAddress.Feature
	isValidRequest := len(conn.OwnDevice.Entities) > entitiyAddr && len(conn.OwnDevice.Entities[entitiyAddr].Features) > featureAddr

	// Count the number of subscriptions
	numSubscriptions := conn.CountSubscriptions(*Function.SubscriptionRequest.ServerAddress)
	if err == nil && isValidRequest && conn.OwnDevice.Entities[entitiyAddr].Features[featureAddr].MaxSubscriptions > numSubscriptions {
		log.Println("Accept subscription from: ", Function.SubscriptionRequest.ClientAddress.Device)

		newEntry := &ressources.BindSubscribeEntry{
			ClientAddress: *Function.SubscriptionRequest.ClientAddress,
			ServerAddress: *Function.SubscriptionRequest.ServerAddress,
		}
		conn.bindSubscribeInfo = append(conn.bindSubscribeInfo, &BindSubscribeInfo{
			"subscription", newEntry,
		})
		conn.bindSubscribeNotify("subscription", conn, newEntry)
		serverAddr := Function.SubscriptionRequest.ServerAddress
		conn.SendXML(
			conn.OwnDevice.MakeHeader(serverAddr.Entity, serverAddr.Feature, Function.SubscriptionRequest.ClientAddress, "result", conn.MsgCounter, false),
			ressources.MakePayload("resultData", ressources.ResultData(0, "positive ackknowledgement for subscription request")))

		funct := conn.OwnDevice.Entities[serverAddr.Entity].Features[serverAddr.Feature].Functions[0]
		conn.SendXML(
			conn.OwnDevice.MakeHeader(serverAddr.Entity, serverAddr.Feature, Function.SubscriptionRequest.ClientAddress, "notify", conn.MsgCounter, false),
			ressources.MakePayload(funct.FunctionName, funct.Function))

		// Save binding
	} else {
		ownAddr := datagram.Header.AddressDestination
		conn.SendXML(
			conn.OwnDevice.MakeHeader(ownAddr.Entity, ownAddr.Feature, datagram.Header.AddressSource, "result", conn.MsgCounter, false),
			ressources.MakePayload("resultData", ressources.ResultData(1, "negative ackknowledgement for subscription request")))
	}
}
