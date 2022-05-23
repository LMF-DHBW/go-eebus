package spine

import (
	"encoding/xml"
	"github.com/LMF-DHBW/go-eebus/ressources"
	"log"
)

func (conn *SpineConnection) sendBindingRequest(EntityAddress int, FeatureAddress int, DestinationAddr *ressources.FeatureAddressType, FeatureType string) {
	clientAddr := ressources.MakeFeatureAddress(conn.OwnDevice.DeviceAddress, EntityAddress, FeatureAddress)
	conn.SendXML(
		conn.OwnDevice.MakeHeader(EntityAddress, FeatureAddress, DestinationAddr, "call", conn.MsgCounter, true),
		ressources.MakePayload("nodeManagementBindingRequestCall", &ressources.NodeManagementBindingRequestCall{
			&ressources.BindingManagementRequestCallType{
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
				log.Println("Binding to: ", DestinationAddr.Device)

				newEntry := &ressources.BindSubscribeEntry{
					ClientAddress: *clientAddr,
					ServerAddress: *DestinationAddr,
				}
				conn.bindSubscribeInfo = append(conn.bindSubscribeInfo, &BindSubscribeInfo{
					"binding", newEntry,
				})
				conn.bindSubscribeNotify("binding", conn, newEntry)
			}
		}
	}
}

func (conn *SpineConnection) processBindingRequest(datagram *ressources.DatagramType) {
	var Function *ressources.NodeManagementBindingRequestCall
	err := xml.Unmarshal([]byte(datagram.Payload.Cmd.Function), &Function)

	entitiyAddr := Function.BindingRequest.ServerAddress.Entity
	featureAddr := Function.BindingRequest.ServerAddress.Feature
	isValidRequest := len(conn.OwnDevice.Entities) > entitiyAddr && len(conn.OwnDevice.Entities[entitiyAddr].Features) > featureAddr

	// Count the number of bindings
	numBindings := conn.CountBindings(*Function.BindingRequest.ServerAddress)

	if err == nil && isValidRequest && conn.OwnDevice.Entities[entitiyAddr].Features[featureAddr].MaxBindings > numBindings {
		log.Println("Binding to: ", Function.BindingRequest.ClientAddress.Device)
		newEntry := &ressources.BindSubscribeEntry{
			ClientAddress: *Function.BindingRequest.ClientAddress,
			ServerAddress: *Function.BindingRequest.ServerAddress,
		}
		conn.bindSubscribeInfo = append(conn.bindSubscribeInfo, &BindSubscribeInfo{
			"binding", newEntry,
		})
		conn.bindSubscribeNotify("binding", conn, newEntry)
		serverAddr := Function.BindingRequest.ServerAddress
		conn.SendXML(
			conn.OwnDevice.MakeHeader(serverAddr.Entity, serverAddr.Feature, Function.BindingRequest.ClientAddress, "result", conn.MsgCounter, false),
			ressources.MakePayload("resultData", ressources.ResultData(0, "positive ackknowledgement for binding request")))
	} else {
		ownAddr := datagram.Header.AddressDestination
		conn.SendXML(
			conn.OwnDevice.MakeHeader(ownAddr.Entity, ownAddr.Feature, datagram.Header.AddressSource, "result", conn.MsgCounter, false),
			ressources.MakePayload("resultData", ressources.ResultData(1, "negative ackknowledgement for binding request")))
	}
}
