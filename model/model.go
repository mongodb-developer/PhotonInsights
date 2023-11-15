package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Module struct {
	Id              int32         `bson:"_id"`
	Payload         ModulePayload `bson:"payload"`
	Deployments     []Deployment  `bson:"deployments"`
	SubscribedUsers []string      `bson:"subscribedUsers,omitempty"`
	Tags            []Tag         `bson:"tags,omitempty"`
	Instances       []string      `bson:"instances"`
}

type Instances struct {
	Instances []string `bson:"instances"`
}

type Tag struct {
	TagName  string `bson:"tagName"`
	TagValue string `bson:"tagValue"`
}

type ModulePayload struct {
	Name            string             `bson:"name"`
	LOB             string             `bson:"LOB"`
	ApplicationID   int32              `bson:"applicationID"`
	PhotonVersion   string             `bson:"photonVersion"`
	MonetaVersion   string             `bson:"monetaVersion"`
	ProductName     string             `bson:"productName"`
	ProductLine     string             `bson:"productLine"`
	ProductOwner    string             `bson:"productOwner"`
	ProductOwnerSID string             `bson:"productOwnerSID"`
	TechPartner     string             `bson:"techPartner"`
	TechPartnerSID  string             `bson:"techPartnerSID"`
	CreatedOn       primitive.DateTime `bson:"createdOn"`
}

type ModulePayloadInstance struct {
	Platform       string `bson:"platform"`
	Environment    string `bson:"environment"`
	DeploymentName string `bson:"deploymentName"`
	InstanceId     string `bson:"InstanceId"`
}

type Deployment struct {
	DeploymentID string            `bson:"deploymentId"`
	Payload      DeploymentPayload `bson:"payload"`
	Instances    []interface{}     `bson:"instances"`
}

type DeploymentPayload struct {
	DeploymentName string `bson:"deploymentName"`
	Env            string `bson:"env"`
	Platform       string `bson:"platform"`
	Version        string `bson:"version"`
	InstanceCount  int    `bson:"instanceCount"`
	HasAnomaly     bool   `bson:"hasAnomoly"`
}

type GaiaInstance struct {
	InstanceId string
}

type Anomaly struct {
	AnomalyId     primitive.ObjectID `bson:"anomalyId"`
	PropertyName  string             `bson:"propertyName"`
	ExpectedValue string             `bson:"expectedValue"`
	CurrentValue  string             `bson:"currentValue"`
	Description   string             `bson:"description"`
	RuleId        int                `bson:"ruleId"`
	Remediated    bool               `bson:"remediated"`
}

type Anomalies struct {
	DeploymentId string              `bson:"_id"`
	Instances    []InstanceAnomalies `bson:"Instances"`
}

type InstanceAnomalies struct {
	InstanceId string    `bson:"instanceId"`
	Anomalies  []Anomaly `bson:"anomalies"`
}
