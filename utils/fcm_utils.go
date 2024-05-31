package utils

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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
		
	secretName := "staging-bonito-google-identity-pool-credentials"
	region := "eu-west-1"

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		logging.Log.Errorf("Error loading AWS config: %v", err)
	}
	logging.Log.Infof("AWS config loaded")
	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
	logging.Log.Infof("AWS Secret loaded")
	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		logging.Log.Infof("Error loading secret value: %v", err)
	}

	// Decrypts secret using the associated KMS key.
	var secretString = *result.SecretString

	if err != nil {
		logging.Log.Infof("AuthorizeAndGetfcmClientFromKey: Error marshalling key: %s", err)
		return nil, err
	}

	opts := []option.ClientOption{option.WithCredentialsJSON([]byte(secretString))}

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