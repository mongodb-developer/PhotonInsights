package datagen

import (
	"context"
	"encoding/json"

	//"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wI2L/jsondiff"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"photoninsights/model"
)

var baseMessage bson.A = nil
var modules = make(map[int32][]string)
var ModulesColl *mongo.Collection
var MessagesColl *mongo.Collection
var ModificationsColl *mongo.Collection

func readMessageFile() {

	// Read the JSON file
	filePath := os.Getenv("MESSAGE_FILE")
	if filePath == "" {
		log.Fatal("You must set your 'MESSAGE_FILE' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Parse the JSON file directly into bson.M
	if err := bson.UnmarshalExtJSON(fileData, true, &baseMessage); err != nil {
		log.Fatal(err.Error())
	}

}

func getModulePayload(appId int32) (model.ModulePayload, bool) {

	filter := bson.D{{Key: "_id", Value: appId}}
	projection := bson.D{
		{Key: "payload", Value: 1},
		{Key: "_id", Value: 0}, // exclude the _id field
	}
	findOptions := options.FindOne().SetProjection(projection)
	var payload model.ModulePayload
	err := ModulesColl.FindOne(context.TODO(), filter, findOptions).Decode(&payload)
	if err == mongo.ErrNoDocuments {
		return payload, false
	}
	if err != nil {
		log.Fatal(err)
	}
	return payload, true

}

func getModuleInstances(appId int32) ([]string, bool) {

	filter := bson.D{{Key: "_id", Value: appId}}
	projection := bson.D{
		{Key: "instances", Value: 1},
		{Key: "_id", Value: 0}, // exclude the _id field
	}
	findOptions := options.FindOne().SetProjection(projection)
	var instances model.Instances
	err := ModulesColl.FindOne(context.TODO(), filter, findOptions).Decode(&instances)
	if err == mongo.ErrNoDocuments {
		return instances.Instances, false
	}
	if err != nil {
		log.Fatal(err)
	}
	return instances.Instances, true

}

// ConvertBsonDToM converts bson.D to bson.M, including handling nested bson.D fields
func ConvertBsonDToM(doc bson.D) bson.M {
	m := make(bson.M)
	for _, element := range doc {
		// Check if the value is also a bson.D and convert recursively
		if nestedDoc, ok := element.Value.(bson.D); ok {
			m[element.Key] = ConvertBsonDToM(nestedDoc)
		} else {
			m[element.Key] = element.Value
		}
	}
	return m
}

func GenerateTestMessage() {

	if baseMessage == nil {
		readMessageFile()
	}

	envs := [...]string{"DEV", "UAT", "PROD"}
	platforms := [...]string{"GAP", "GKP", "LOCAL"}

	//Generate the application Id:
	appId := generateRandomInt(1, 4500)
	appName := "application_" + strconv.Itoa(appId)

	//Generate the deployment information
	env := envs[generateRandomInt(0, 2)]
	platform := platforms[generateRandomInt(0, 2)]
	deploymentName := "deployment_" + strconv.Itoa(generateRandomInt(1, 3))

	//Generate the Instance Data
	instId := "instance_" + strconv.Itoa(generateRandomInt(1, 3))

	//Create a copy of base message
	newMessage := ConvertBsonDToM(baseMessage[generateRandomInt(0, 4)].(bson.D))

	//Modify the copy with the new values
	newMessage["cloud"].(bson.M)["appId"] = appId
	newMessage["cloud"].(bson.M)["appName"] = appName + "#" + deploymentName
	newMessage["cloud"].(bson.M)["instanceId"] = strconv.Itoa(appId) + ":" + platform + ":" + env + ":" + deploymentName + ":" + instId
	newMessage["cloud"].(bson.M)["space_id"] = platform
	newMessage["cloud"].(bson.M)["space_name"] = env
	newMessage["application"].(bson.M)["name"] = appName
	newMessage["applicationConfig"].(bson.M)["spring.application.name"] = appName

	//Convert the message to a JSON byte array
	jsonMessage, err := json.MarshalIndent(newMessage, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	//Call the message processor
	ProcessMessage(jsonMessage)
}

func ProcessMessage(jsonMessage []byte) {

	newModule := false
	newDeployment := false
	newInstance := false

	//Convert the message to bson.M
	var data bson.M
	if err := bson.UnmarshalExtJSON(jsonMessage, true, &data); err != nil {
		log.Fatal(err)
	}
	data["baseLine"] = false

	var appName, instanceId, platform, env, deploymentName string
	var appId int32

	//Get the application data
	if cloud, ok := data["cloud"].(bson.M); ok {
		appId = cloud["appId"].(int32)
		appName = cloud["appName"].(string)
		instanceId = cloud["instanceId"].(string)
		platform = cloud["space_id"].(string)
		env = cloud["space_name"].(string)

		//Split the app name and deployment ID
		parts := strings.Split(appName, "#")
		appName = parts[0]
		deploymentName = parts[1]
	} else {
		log.Fatal("Couldn't parse cloud section of received message")
	}

	//Do we already have this app? If not, generate an module struct for it
	var modInstances []string
	var mod model.Module
	modInstances, ok := modules[appId]
	if !ok {
		//We don't have local copy - attempt to get it from the database
		modInstances, ok = getModuleInstances(appId)
		if !ok {
			//Module is not in the database either - so create a new one
			mod.Id = appId
			mod.Payload.Name = appName
			mod.Payload.ApplicationID = appId
			//The following should be read from the message rather than hardcoded
			mod.Payload.LOB = "CCB"
			mod.Payload.PhotonVersion = "2.6.3-SNAPSHOT"
			mod.Payload.MonetaVersion = "2.6.4"
			mod.Payload.ProductName = "Business Banking - SB Lending"
			mod.Payload.ProductLine = "SB Lending"
			mod.Payload.ProductOwner = "Henry Smith"
			mod.Payload.ProductOwnerSID = "HS5678"
			mod.Payload.TechPartner = "Martin Baker"
			mod.Payload.TechPartnerSID = "MB1234"
			startDate, err := time.Parse("2006-01-02T15:04:05-0700", data["startTime"].(string))
			if err != nil {
				log.Fatal("Couldn't parse start date in message")
			}
			mod.Payload.CreatedOn = primitive.DateTime(startDate.UnixNano() / int64(time.Millisecond))
			mod.Instances = make([]string, 0)
			mod.Deployments = make([]model.Deployment, 0)
			newModule = true
		}
	} else {
		//log.Print("Found instance in local cache")
	}

	//Generate the Instance Data
	instFound := false
	var inst model.GaiaInstance
	var modPayloadInst model.ModulePayloadInstance
	modPayloadInst.Platform = platform
	modPayloadInst.Environment = env
	modPayloadInst.DeploymentName = deploymentName
	modPayloadInst.InstanceId = instanceId
	for _, instance := range modInstances {
		if instance == instanceId {
			instFound = true
			break
		}
	}
	if !instFound {
		inst.InstanceId = instanceId
		mod.Instances = append(mod.Instances, instanceId)
		newInstance = true
	}

	//If this is a previously unseen deployment, add it.
	depFound := false
	var dep model.Deployment
	for _, instance := range modInstances {
		instanceParts := strings.Split(instance, ":")
		if instanceParts[3] == deploymentName && instanceParts[2] == env && instanceParts[1] == platform {
			depFound = true
			break
		}
	}
	if !depFound {
		dep.DeploymentID = uuid.NewString()
		dep.Payload.DeploymentName = deploymentName
		dep.Payload.Env = env
		dep.Payload.Platform = platform
		dep.Payload.Version = "2.6.3-SNAPSHOT"
		dep.Payload.HasAnomaly = false
		dep.Payload.InstanceCount = 1
		dep.Instances = make([]interface{}, 0)
		dep.Instances = append(dep.Instances, inst)
		mod.Deployments = append(mod.Deployments, dep)
		newDeployment = true
	}

	//If this is a new instance, add the message to messages collection
	if newInstance {
		data["baseLine"] = false
		data["_id"] = instanceId
		_, err := MessagesColl.InsertOne(context.TODO(), data)
		if err != nil {
			log.Fatal(err)
		} else {
			//log.Printf("Created message document with _id: %s", res.InsertedID.(string))
		}
		//If this is also the first message for this deployment, save the message again - once with a flag indicating it is the baseline message for this deployment
		if newDeployment {
			data["baseLine"] = true
			data["_id"] = instanceId + ":" + "baseLine"
			_, err = MessagesColl.InsertOne(context.TODO(), data)
			if err != nil {
				log.Fatal(err)
			} else {
				//log.Printf("Created baseline message document with _id: %s", res.InsertedID.(string))
			}
		}
	} else {
		//This is an exisitng instance - save any deltas
		//Serarching by the cloud values is only necessary in our test data - in a real environment you'd save the instance ID GUID as _id and search by that
		filter := bson.M{"_id": instanceId, "baseLine": false}
		var lastMessage bson.M
		err := MessagesColl.FindOne(context.TODO(), filter).Decode(&lastMessage)
		if err != nil {
			log.Fatal(err)
		}
		jsonDataOrig, err := json.MarshalIndent(lastMessage, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		jsonDataNew, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Fatal(err)
		}

		patch, err := jsondiff.CompareJSON(jsonDataOrig, jsonDataNew)
		if err != nil {
			log.Fatal(err)
		}
		//mods will hold the modification documents
		var mods bson.A

		//pipline will hold an aggregation pipleing to apply modifications to the existing message document
		// we need to do it this way to handle fields with dots in their names.
		var pipeline bson.A
		plMatch := bson.M{"$match": filter}
		pipeline = append(pipeline, plMatch)

		for _, op := range patch {
			//Ignore removal of _id
			if !(op.Type == jsondiff.OperationRemove && op.Path == "/_id") {
				//fmt.Printf("%s\n", op)
				modification := make(bson.M)
				modification["instanceId"] = instanceId
				modification["operation"] = op.Type
				modification["oldValue"] = op.OldValue
				modification["newValue"] = op.Value
				modification["path"] = op.Path
				modification["date"] = time.Now()
				mods = append(mods, modification)

				//Construct the agg pipline stage to apply the change to the message doc
				pathArray := strings.Split(op.Path, "/")
				path := ""
				field := pathArray[len(pathArray)-1]
				for i := 0; i < len(pathArray)-1; i++ {
					if path != "" {
						path += "."
					}
					path += pathArray[i]
				}
				var plReplace bson.M
				if op.Type == jsondiff.OperationRemove {
					plReplace = bson.M{"$set": bson.M{path: bson.M{"$setField": bson.M{"field": field, "input": "$" + path, "value": "$$REMOVE"}}}}
				} else {
					plReplace = bson.M{"$set": bson.M{path: bson.M{"$setField": bson.M{"field": field, "input": "$" + path, "value": op.Value}}}}
				}
				pipeline = append(pipeline, plReplace)
			}
		}
		if len(pipeline) > 1 {
			//Update the message
			//Add the merge stage
			plMerge := bson.M{"$merge": bson.M{"into": "Messages", "on": "_id", "whenMatched": "merge"}}
			pipeline = append(pipeline, plMerge)
			_, err := MessagesColl.Aggregate(context.TODO(), pipeline)
			if err != nil {
				log.Fatal(err)
			} else {
				//log.Printf("Updated message document: %s", filter)
			}
			//Record the changes
			_, err = ModificationsColl.InsertMany(context.TODO(), mods)
			if err != nil {
				log.Fatal(err)
			} else {
				//log.Printf("Inserted %d new mdofication docuemnts", len(res.InsertedIDs))
			}

		}
	}

	//Do any necessary updates to the module doc
	//Is this a new instance?
	if newModule {
		_, err := ModulesColl.InsertOne(context.TODO(), mod)
		if err != nil {
			log.Fatal(err)
		}
		//log.Printf("Inserted module document with id: %d", res.InsertedID.(int32))
	} else if newDeployment {
		filter := bson.M{"_id": appId}
		update := bson.M{
			"$push": bson.M{"deployments": dep, "instances": instanceId},
		}
		res, err := ModulesColl.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			log.Fatal(err)
		}
		if res.ModifiedCount == 1 {
			//log.Printf("Updated module document with id: %d", appId)
		} else {
			log.Printf("Failed to update module document with id: %d", appId)
		}
	} else if newInstance {
		filter := bson.M{"_id": appId}
		update := bson.M{
			"$push": bson.M{"deployments.$[elem].instances": inst, "instances": instanceId},
			"$inc":  bson.M{"deployments.$[elem].payload.instanceCount": 1},
		}
		arrayFilters := options.ArrayFilters{
			Filters: bson.A{
				bson.M{"$and": bson.A{
					bson.M{"elem.payload.deploymentName": deploymentName},
					bson.M{"elem.payload.env": env},
					bson.M{"elem.payload.platform": platform},
				}},
			},
		}
		updateOptions := options.Update().SetArrayFilters(arrayFilters)
		res, err := ModulesColl.UpdateOne(context.TODO(), filter, update, updateOptions)
		if err != nil {
			log.Fatal(err)
		}
		if res.ModifiedCount == 1 {
			//log.Printf("Updated module document with id: %d", appId)
		} else {
			log.Printf("Failed to update module document with id: %d", appId)
		}
	}
	//Module data has been saved to the datbase if necessary so we can include it in our local cache.
	if newModule || newDeployment || newInstance {
		modInstances = append(modInstances, instanceId)
		modules[appId] = modInstances
	}
}

// generateRandomInt generates a random integer between min and max (inclusive).
func generateRandomInt(min, max int) int {
	// Generate a random number in the range [0, max-min] and shift it by min.
	return rand.Intn(max-min+1) + min
}
