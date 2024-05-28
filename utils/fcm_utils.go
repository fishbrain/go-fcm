package utils

import (
	"context"
	"io"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/fishbrain/go-fcm/config"
	logging "github.com/fishbrain/logging-go"
	"google.golang.org/api/option"
)

var osOpen = os.Open
var ioReadAll = io.ReadAll
var firebaseNewApp = firebase.NewApp

func AuthorizeAndGetFirebaseMessagingClient() (*messaging.Client, error) {
	
	fileName := "workload_identity_pool_credentials_" + config.Config.Environment + ".json"

	logging.Log.Infof("Opening file: %s", fileName)
	
	file, err := osOpen("../data/gcp/" + fileName)
	if err != nil {
		logging.Log.Errorf("Error opening file: %s", err)
		return nil, err
	}
	defer file.Close()

	gcpCredentials, err := ioReadAll(file)
	if err != nil {
		logging.Log.Errorf("Error reading file: %s", err)
		return nil, err
	}
	logging.Log.Infof("GCP credentials: %v", gcpCredentials)

	opts := []option.ClientOption{option.WithCredentialsJSON(gcpCredentials)}

	firebaseApp, err := firebaseNewApp(context.Background(), nil, opts...)

	if err != nil {
		logging.Log.Infof("Error initializing firebase app: %s", err)
		return nil, err
	}
	logging.Log.Infof("App: %s", firebaseApp)

	fcmClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		logging.Log.Infof("Error initializing FCM client: %s", err)
		return nil, err
	}

	return fcmClient, err
}
