# PhotonInsights
Go project to create test data for the JPMC Photon Insights proposed schema

The 'env_atlas' file should be updated with the connection URI for your target Atlas Cluster
and the location of a JSON file containing an array of sample phone-home messages.

The name of the MongoDB database and collections can be updated on lines 55, 57, 59 and 61
photoninsights.go

The number of messages to process can be updated in the for-loop definition on line 64 of photoninsights.go

The data generates values for up to 4500 appplications. Each applications can have one or more deployments.
A deployment is tied to a combination of 1 of 3 cloud providers (GAP, GKP, LOCAL) and 1 of 3 environments 
(DEV, UAT, PROD) - so 9 combinations in total. Each combination can have up to 3 associated deployments.
A deployment can have between one and three instances. This means an application can have up to 81 
instances associated with it. These values can be modified on lines 111 through 124 of datagen/datagen.go

Line 127 of datagen/datagen.go should be modified to reflect the number of sample JSON phone-home messages
in the provided file. The default code assumes an array of 5 messages (indexed 0 through 4) in the call
to generateRandomInt.

To build the project, download and install Go 1.21.1 or later. Then, from a terminal in the project
directory, get the required libraries:

go get github.com/joho/godotenv
go get github.com/wI2L/jsondiff
go get github.com/google/uuid
go get go.mongodb.org/mongo-driver/mongo

Finally, to run the project, from a terminal in the project directory type:

go run photoninsights.go

To compile the project to an executable, type:

go build

Note - the current code processes messages sequentially and is pretty slow as a result. This code could be improved
by running multiple threads to process messages in parallel and/or using the MongoDB updateMany and insertMany methods
rather than updateOne / insertOne methods respectively as much of the porcessing time is network latency / overhead.
