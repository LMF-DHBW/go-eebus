package go_eebus

import (
	"github.com/LMF-DHBW/go-eebus/ressources"
	"github.com/LMF-DHBW/go-eebus/spine"
)

type EebusNode struct {
	isGateway       bool
	SpineNode       *spine.SpineNode
	DeviceStructure *ressources.DeviceModel
	Update          Updater
}

type Updater func(ressources.DatagramType, spine.SpineConnection)

func NewEebusNode(isGateway bool, certName string, devId string, brand string, devType string) *EebusNode {
	deviceModel := &ressources.DeviceModel{}
	newEebusNode := &EebusNode{isGateway, nil, deviceModel, nil}
	newEebusNode.SpineNode = spine.NewSpineNode(isGateway, deviceModel, newEebusNode.SubscriptionNofity, certName, devId, brand, devType)
	return newEebusNode
}

func (eebusNode *EebusNode) SubscriptionNofity(datagram ressources.DatagramType, conn spine.SpineConnection) {
	if eebusNode.Update != nil {
		eebusNode.Update(datagram, conn)
	}
}

func (eebusNode *EebusNode) Start() {
	eebusNode.SpineNode.Start()
}
