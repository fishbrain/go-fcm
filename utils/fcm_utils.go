package utils

import (
	"context"
	_ "embed"
	"encoding/json"
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

func AuthorizeAndGetfcmClientFromKey() (*messaging.Client, error) {

	var secretString = os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")

	opts := []option.ClientOption{option.WithCredentialsJSON([]byte(secretString))}

	firebaseApp, err := firebaseNewApp(context.Background(), nil, opts...)

	if err != nil {
		logging.Log.Infof("Error initializing firebase app: %s", err)
		return nil, err
	}

	fcmClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		logging.Log.Infof("Error initializing FCM client: %s", err)
		return nil, err
	}

	return fcmClient, err
}

func AuthorizeAndGetfcmClientFromIdPoolKey() (*messaging.Client, error) {
		
	key := map[string]interface{}{
		"type": "external_account",
		"audience": "//iam.googleapis.com/projects/10207772235/locations/global/workloadIdentityPools/bonito-staging-fcm/providers/bonito-staging-fcm-aws",
		"subject_token_type": "urn:ietf:params:aws:token-type:aws4_request",
		"service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/firebase-adminsdk-3karf@api-project-10207772235.fishbrain.com.iam.gserviceaccount.com:generateAccessToken",
		"token_url": "https://sts.googleapis.com/v1/token",
		"credential_source": map[string]string{
		  "environment_id": "aws1",
		  "region_url": "http://169.254.169.254/latest/meta-data/placement/availability-zone",
		  "url": "http://169.254.169.254/latest/meta-data/iam/security-credentials",
		  "regional_cred_verification_url": "https://sts.{region}.amazonaws.com?Action=GetCallerIdentity&Version=2011-06-15",
		},
	  }
	
	gcpCredentials, err := json.Marshal(key)

	if err != nil {
		logging.Log.Infof("AuthorizeAndGetfcmClientFromKey: Error marshalling key: %s", err)
		return nil, err
	}

	opts := []option.ClientOption{option.WithCredentialsJSON(gcpCredentials)}

	projectId := os.Getenv("GCP_PROD_PROJECT_ID")
	logging.Log.Infof("AuthorizeAndGetfcmClientFromKey: Initializing firebase app with project ID: %s", projectId)
	firebaseApp, err := firebaseNewApp(context.Background(), nil, opts...)

	if err != nil {
		logging.Log.Infof("AuthorizeAndGetfcmClientFromKey: Error initializing firebase app: %s", err)
		return nil, err
	}
	logging.Log.Infof("App: %s", firebaseApp)

	fcmClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		logging.Log.Infof("AuthorizeAndGetfcmClientFromKey: Error initializing FCM client: %s", err)
		return nil, err
	}
	logging.Log.Infof("AuthorizeAndGetfcmClientFromKey: FCM Client: %v", fcmClient)

	return fcmClient, err
}