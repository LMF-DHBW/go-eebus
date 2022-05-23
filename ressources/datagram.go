package ressources

import (
	"encoding/xml"
	"fmt"
	"time"
)

const SPECIFICATION_VERSION = "1.0.0"

type DeviceModel struct {
	DeviceType    string         `xml:"deviceType"`
	DeviceAddress string         `xml:"deviceAddress"`
	Description   string         `xml:"description"`
	Entities      []*EntityModel `xml:"entities"`
}

type NodeManagementBindingData struct {
	BindingEntries []*BindSubscribeEntry `xml:"bindingEntries"`
}

type BindSubscribeEntry struct {
	ClientAddress FeatureAddressType `xml:"clientAddress"`
	ServerAddress FeatureAddressType `xml:"serverAddress"`
}

type NodeManagementSubscriptionData struct {
	SubscriptionEntries []*BindSubscribeEntry `xml:"subscriptionEntries"`
}

type EntityModel struct {
	EntityType    string          `xml:"entityType"`
	EntityAddress int             `xml:"entityAddress"`
	Description   string          `xml:"description"`
	Features      []*FeatureModel `xml:"features"`
}

type FeatureModel struct {
	FeatureType      string           `xml:"featureType"`
	FeatureAddress   int              `xml:"featureAddress"`
	Role             string           `xml:"role"`
	Description      string           `xml:"description"`
	Functions        []*FunctionModel `xml:"functions"`
	BindingTo        []string
	SubscriptionTo   []string
	MaxBindings      int
	MaxSubscriptions int
}

type Notifier func(string, string, FeatureAddressType)

type FunctionModel struct {
	FunctionName string      `xml:"functionName"`
	ChangeNotify Notifier    `xml:"changeNotify"`
	Function     interface{} `xml:"function"`
}

type FunctionElement struct {
	Function string `xml:"function"`
}

type ResultElement struct {
	ErrorNumber int    `xml:"errorNumber"`
	Description string `xml:"description"`
}

type DescriptionElement struct {
	Label       string `xml:"label"`
	Description string `xml:"description"`
}

type NodeManagementSpecificationVersionListType struct {
	SpecificationVersion string `xml:"specificationVersion"`
}

type NodeManagementDetailedDiscoveryDeviceInformationType struct {
	Description *NetworkManagementDeviceDescriptionDataType `xml:"description"`
}

type NodeManagementDetailedDiscoveryEntityInformationType struct {
	Description *NetworkManagementEntityDescritpionDataType `xml:"description"`
}

type NodeManagementDetailedDiscoveryFeatureInformationType struct {
	Description *NetworkManagementFeatureInformationType `xml:"description"`
}

type NetworkManagementDeviceDescriptionDataType struct {
	DeviceAddress *DeviceAddressType `xml:"deviceAddress"`
	DeviceType    string             `xml:"deviceType"`
	Description   string             `xml:"description"`
}

type NetworkManagementEntityDescritpionDataType struct {
	EntityAddress *EntityAddressType `xml:"entityAddress"`
	EntityType    string             `xml:"entityType"`
	Description   string             `xml:"description"`
}

type NetworkManagementFeatureInformationType struct {
	FeatureAddress    *FeatureAddressType   `xml:"featureAddress"`
	FeatureType       string                `xml:"featureType"`
	Role              string                `xml:"role"`
	SupportedFunction *FunctionPropertyType `xml:"supportedFunction"`
	Description       string                `xml:"description"`
}

type NodeManagementDetailedDiscovery struct {
	SpecificationVersionList []*NodeManagementSpecificationVersionListType            `xml:"specificationVersionList"`
	DeviceInformation        *NodeManagementDetailedDiscoveryDeviceInformationType    `xml:"deviceInformation"`
	EntityInformation        []*NodeManagementDetailedDiscoveryEntityInformationType  `xml:"entityInformation"`
	FeatureInformation       []*NodeManagementDetailedDiscoveryFeatureInformationType `xml:"featureInformation"`
}

type FeatureAddressType struct {
	Device  string `xml:"device"`
	Entity  int    `xml:"entity"`
	Feature int    `xml:"feature"`
}

type EntityAddressType struct {
	Device string `xml:"device"`
	Entity int    `xml:"entity"`
}

type DeviceAddressType struct {
	Device string `xml:"device"`
}

type FunctionPropertyType struct {
	Function           string `xml:"function"`
	PossibleOperations string `xml:"possibleOperations"`
}

type DatagramType struct {
	Header  *HeaderType  `xml:"header"`
	Payload *PayloadType `xml:"payload"`
}

type PayloadType struct {
	Cmd *CmdType `xml:"cmd"`
}

type CmdType struct {
	FunctionName string `xml:"functionName"`
	Function     string `xml:"function"`
}

type HeaderType struct {
	SpecificationVersion string              `xml:"specificationVersion"`
	AddressSource        *FeatureAddressType `xml:"addressSource"`
	AddressDestination   *FeatureAddressType `xml:"addressDestination"`
	MsgCounter           int                 `xml:"msgCounter"`
	CmdClassifier        string              `xml:"cmdClassifier"`
	Timestamp            string              `xml:"timestamp"`
	AckRequest           bool                `xml:"ackRequest"`
}

type NodeManagementBindingRequestCall struct {
	BindingRequest *BindingManagementRequestCallType `xml:"bindingRequest"`
}

type BindingManagementRequestCallType struct {
	ClientAddress     *FeatureAddressType `xml:"clientAddress"`
	ServerAddress     *FeatureAddressType `xml:"serverAddress"`
	ServerFeatureType string              `xml:"serverFeatureType"`
}

type NodeManagementSubscriptionRequestCall struct {
	SubscriptionRequest *SubscriptionManagementRequestCallType `xml:"subscriptionRequest"`
}

type SubscriptionManagementRequestCallType struct {
	ClientAddress     *FeatureAddressType `xml:"clientAddress"`
	ServerAddress     *FeatureAddressType `xml:"serverAddress"`
	ServerFeatureType string              `xml:"serverFeatureType"`
}

func ResultData(errorNumber int, description string) *ResultElement {
	return &ResultElement{
		ErrorNumber: errorNumber,
		Description: description,
	}
}

type TimePeriodType struct {
	StartTime string `xml:"startTime"`
	EndTime   string `xml:"endTime"`
}

type MeasurementDataType struct {
	ValueType        string         `xml:"valueType"`
	Timestamp        string         `xml:"timestamp"`
	Value            float64        `xml:"value"`
	EvaluationPeriod TimePeriodType `xml:"evaluationPeriod"`
	ValueSource      string         `xml:"valueSource"`
	ValueTendency    string         `xml:"valueTendency"`
	ValueState       string         `xml:"valueState"`
}

type MeasurementDescriptionDataType struct {
	MeasurementType string `xml:"measurementType"`
	Unit            string `xml:"unit"`
	ScopeType       string `xml:"scopeType"`
	Label           string `xml:"label"`
	Description     string `xml:"description"`
}

func Measurement(MeasurementType string, Unit string, ScopeType string, Label string, Description string) []*FunctionModel {
	return []*FunctionModel{
		{
			FunctionName: "measurementData",
			Function:     &MeasurementDataType{},
		},
		{
			FunctionName: "measurementDescription",
			Function: &MeasurementDescriptionDataType{
				MeasurementType,
				Unit,
				ScopeType,
				Label,
				Description,
			},
		},
	}
}

func ActuatorSwitch(role string, label string, description string, ChangeNotify Notifier) []*FunctionModel {
	return []*FunctionModel{
		{
			FunctionName: "actuatorSwitchData",
			Function: &FunctionElement{
				Function: "off",
			},
			ChangeNotify: ChangeNotify,
		},
		{
			FunctionName: "actuatorSwitchDescriptionData",
			Function: &DescriptionElement{
				label,
				description,
			},
		},
	}
}

func (device *DeviceModel) CreateNodeManagement(isGateway bool) *FeatureModel {
	subscriptions := []string{}
	bindings := []string{}
	if isGateway {
		subscriptions = append(subscriptions, []string{"ActuatorSwitch", "Measurement", "NodeManagement"}...)
		bindings = append(bindings, "ActuatorSwitch")
	}
	return &FeatureModel{
		FeatureType:      "NodeManagement",
		FeatureAddress:   0,
		Role:             "special",
		SubscriptionTo:   subscriptions,
		BindingTo:        bindings,
		MaxBindings:      128,
		MaxSubscriptions: 128,
		Functions: []*FunctionModel{
			{
				FunctionName: "nodeManagementDetailedDiscoveryData",
				Function: &NodeManagementDetailedDiscovery{
					SpecificationVersionList: []*NodeManagementSpecificationVersionListType{
						{
							SpecificationVersion: SPECIFICATION_VERSION,
						},
					},
					DeviceInformation: &NodeManagementDetailedDiscoveryDeviceInformationType{
						Description: &NetworkManagementDeviceDescriptionDataType{
							DeviceAddress: &DeviceAddressType{
								Device: device.DeviceAddress,
							},
							DeviceType:  device.DeviceType,
							Description: device.Description,
						},
					},
					EntityInformation:  makeEntities(device),
					FeatureInformation: makeFeatures(device),
				},
			},
			{
				FunctionName: "nodeManagementBindingData",
				Function:     &NodeManagementBindingData{},
			},
			{
				FunctionName: "nodeManagementSubscriptionData",
				Function:     &NodeManagementSubscriptionData{},
			},
		},
	}
}

func makeEntities(device *DeviceModel) []*NodeManagementDetailedDiscoveryEntityInformationType {
	var all []*NodeManagementDetailedDiscoveryEntityInformationType
	entities := device.Entities
	for _, entity := range entities {
		all = append(all, &NodeManagementDetailedDiscoveryEntityInformationType{
			Description: &NetworkManagementEntityDescritpionDataType{
				EntityAddress: &EntityAddressType{
					Device: device.DeviceAddress,
					Entity: entity.EntityAddress,
				},
				EntityType:  entity.EntityType,
				Description: entity.Description,
			},
		})
	}
	return all
}

func makeFeatures(device *DeviceModel) []*NodeManagementDetailedDiscoveryFeatureInformationType {
	var all []*NodeManagementDetailedDiscoveryFeatureInformationType
	entities := device.Entities
	for _, entity := range entities {
		for _, feature := range entity.Features {
			all = append(all, &NodeManagementDetailedDiscoveryFeatureInformationType{
				Description: &NetworkManagementFeatureInformationType{
					FeatureAddress: &FeatureAddressType{
						Device:  device.DeviceAddress,
						Entity:  entity.EntityAddress,
						Feature: feature.FeatureAddress,
					},
					FeatureType: feature.FeatureType,
					Description: feature.Description,
					Role:        feature.Role,
				},
			})
		}
	}
	return all
}

func (device *DeviceModel) MakeHeader(entity int, feature int, addressDestination *FeatureAddressType, cmdClassifier string, msgCounter int, ackRequest bool) *HeaderType {
	return &HeaderType{
		SpecificationVersion: SPECIFICATION_VERSION,
		AddressSource: &FeatureAddressType{
			Device:  device.DeviceAddress,
			Entity:  entity,
			Feature: feature,
		},
		AddressDestination: addressDestination,
		MsgCounter:         msgCounter,
		CmdClassifier:      cmdClassifier,
		Timestamp:          timestampNow(),
		AckRequest:         ackRequest,
	}
}

func MakePayload(FunctionName string, Function interface{}) *PayloadType {
	return &PayloadType{
		Cmd: &CmdType{
			FunctionName,
			xmlToString(Function),
		},
	}
}

func MakeFeatureAddress(device string, entity int, feature int) *FeatureAddressType {
	return &FeatureAddressType{
		device,
		entity,
		feature,
	}
}

func timestampNow() string {
	current_time := time.Now()

	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d.0Z",
		current_time.Year(), current_time.Month(), current_time.Day(),
		current_time.Hour(), current_time.Minute(), current_time.Second())
}

func xmlToString(in interface{}) string {
	bytes, err := xml.Marshal(in)
	CheckError(err)
	return string(bytes)
}
