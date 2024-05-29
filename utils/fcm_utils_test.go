package utils

import (
	"context"
	"fmt"
	"os"
	"testing"

	firebase "firebase.google.com/go/v4"
	logging "github.com/fishbrain/logging-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/option"
)

// Mock types
type MockFile struct {
	mock.Mock
}

type MockFirebaseApp struct {
	mock.Mock
}

func TestMain(m *testing.M) {
	logging.Init(logging.LoggingConfig{})
	os.Exit(m.Run())
}

func TestAuthorizeAndGetFirebaseMessagingClient_ErrorInitializingFirebaseApp(t *testing.T) {

	// Mock firebase.NewApp to return an error
	firebaseNewApp = func(ctx context.Context, config *firebase.Config, opts ...option.ClientOption) (*firebase.App, error) {
		return nil, fmt.Errorf("error initializing firebase app")
	}
	defer func() { firebaseNewApp = firebase.NewApp }()

	_, err := AuthorizeAndGetFirebaseMessagingClient()
	assert.Error(t, err)
}

func TestAuthorizeAndGetFirebaseMessagingClient_ErrorInitializingFCMClient(t *testing.T) {
	
	// Mock firebase.NewApp to return a mocked firebase app
	firebaseNewApp = func(ctx context.Context, config *firebase.Config, opts ...option.ClientOption) (*firebase.App, error) {
		mockFirebaseApp := new(MockFirebaseApp)
		mockFirebaseApp.On("Messaging", context.Background()).Return(nil, fmt.Errorf("error initializing FCM client"))
		return &firebase.App{}, nil
	}
	defer func() { firebaseNewApp = firebase.NewApp }()

	_, err := AuthorizeAndGetFirebaseMessagingClient()
	assert.Error(t, err)
}