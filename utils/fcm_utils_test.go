package utils

import (
	"context"
	"fmt"
	"io"
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

func TestAuthorizeAndGetFirebaseMessagingClient_ErrorOpeningFile(t *testing.T) {
	// Mock os.Open to return an error
	osOpen = func(name string) (*os.File, error) {
		return nil, fmt.Errorf("error opening file")
	}
	defer func() { osOpen = os.Open }()

	_, err := AuthorizeAndGetFirebaseMessagingClient()
	assert.Error(t, err)
}

func TestAuthorizeAndGetFirebaseMessagingClient_ErrorReadingFile(t *testing.T) {
	// Mock os.Open to return a file
	osOpen = func(name string) (*os.File, error) {
		return &os.File{}, nil
	}
	defer func() { osOpen = os.Open }()

	// Mock io.ReadAll to return an error
	ioReadAll = func(r io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("error reading file")
	}
	defer func() { ioReadAll = io.ReadAll }()

	_, err := AuthorizeAndGetFirebaseMessagingClient()
	assert.Error(t, err)
}

func TestAuthorizeAndGetFirebaseMessagingClient_ErrorInitializingFirebaseApp(t *testing.T) {
	// Mock os.Open to return a file
	osOpen = func(name string) (*os.File, error) {
		return &os.File{}, nil
	}
	defer func() { osOpen = os.Open }()

	// Mock io.ReadAll to return some credentials
	ioReadAll = func(r io.Reader) ([]byte, error) {
		return []byte("mocked credentials"), nil
	}
	defer func() { ioReadAll = io.ReadAll }()

	// Mock firebase.NewApp to return an error
	firebaseNewApp = func(ctx context.Context, config *firebase.Config, opts ...option.ClientOption) (*firebase.App, error) {
		return nil, fmt.Errorf("error initializing firebase app")
	}
	defer func() { firebaseNewApp = firebase.NewApp }()

	_, err := AuthorizeAndGetFirebaseMessagingClient()
	assert.Error(t, err)
}

func TestAuthorizeAndGetFirebaseMessagingClient_ErrorInitializingFCMClient(t *testing.T) {
	// Mock os.Open to return a file
	osOpen = func(name string) (*os.File, error) {
		mockFile := new(MockFile)
		mockFile.On("Close").Return(nil)
		return &os.File{}, nil
	}
	defer func() { osOpen = os.Open }()

	// Mock io.ReadAll to return some credentials
	ioReadAll = func(r io.Reader) ([]byte, error) {
		return []byte("mocked credentials"), nil
	}
	defer func() { ioReadAll = io.ReadAll }()

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