package utils

import (
	"context"
	_ "embed"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	logging "github.com/fishbrain/logging-go"
	"google.golang.org/api/option"
)

var firebaseNewApp = firebase.NewApp

//go:embed workload_identity_pool_credentials_staging.json
var gcpCredentialsStaging []byte

//go:embed workload_identity_pool_credentials_production.json
var gcpCredentialsProduction []byte

func AuthorizeAndGetFirebaseMessagingClient() (*messaging.Client, error) {
	
	environment := os.Getenv("BONITO_ENV")
	
	var gcpCredentials []byte

	if environment == "staging" {
		gcpCredentials = gcpCredentialsStaging
	} else if environment == "production" {
		gcpCredentials = gcpCredentialsProduction
	}

	opts := []option.ClientOption{option.WithCredentialsJSON(gcpCredentials)}

	projectId := os.Getenv("GCP_PROD_PROJECT_ID")
	logging.Log.Infof("Initializing firebase app with project ID: %s", projectId)
	firebaseApp, err := firebaseNewApp(context.Background(), &firebase.Config{ProjectID: projectId}, opts...)

	if err != nil {
		logging.Log.Errorf("Error initializing firebase app: %s", err)
		return nil, err
	}
	logging.Log.Infof("App: %s", firebaseApp)

	fcmClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		logging.Log.Errorf("Error initializing FCM client: %s", err)
		return nil, err
	}
	logging.Log.Infof("FCM Client: %v", fcmClient)

	return fcmClient, err
}
