package fcm

import (
	logging "github.com/fishbrain/logging-go"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// Function to get temporary AWS credentials
func getTemporaryAWSCredentials() (*sts.Credentials, error) {
	// Create a new session with AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		return nil, err
	}

	// Create a new STS client
	svc := sts.New(sess)

	// Get the caller identity (in case you need it for logging/debugging)
	callerIdentity, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	logging.Log.Info("Caller Identity: %s, Account: %s, ARN: %s", *callerIdentity.UserId, *callerIdentity.Account, *callerIdentity.Arn)

	// GetSessionToken to get temporary credentials
	input := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(3600), // Set the duration for temporary credentials
	}

	result, err := svc.GetSessionToken(input)
	if err != nil {
		return nil, err
	}

	return result.Credentials, nil
}
