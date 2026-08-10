package fcm

import (
	"errors"
	"testing"

	messaging "firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/require"
)

// NOTE: The tests below cover only the "no code available" side of
// fcmErrorCode and httpStatus, and that is all they can cover. Producing a real
// FCM error needs the SDK's internal error type, which cannot be constructed
// from outside the SDK, and the messaging endpoint is unexported with no
// emulator override, so a fake 404 cannot be driven through a real client
// either. The mapping from an FCM error to its code string is therefore only
// exercised in production - the code strings were instead checked by hand
// against the SDK constants in messaging.go. Change them with care.

func TestToFcmResponseResults(t *testing.T) {
	t.Run("successful send carries no error keys", func(t *testing.T) {
		responses := []*messaging.SendResponse{
			{Success: true, MessageID: "123"},
		}

		result := *toFcmResponseResults(&responses)

		require.Equal(t, []map[string]string{
			{"success": "true", "messageID": "123"},
		}, result)
	})

	t.Run("failed send keeps the error message", func(t *testing.T) {
		responses := []*messaging.SendResponse{
			{Success: false, Error: errors.New("Device unregistered.")},
		}

		result := *toFcmResponseResults(&responses)

		require.Equal(t, "Device unregistered.", result[0]["error"])
	})

	// An error that did not come from an FCM response has no error code and no
	// HTTP status to report. Both keys stay absent so that callers can tell
	// "no code available" from a real code and fall back to the message.
	t.Run("error without an FCM error code omits errorCode and httpStatus", func(t *testing.T) {
		responses := []*messaging.SendResponse{
			{Success: false, Error: errors.New("connection reset")},
		}

		result := *toFcmResponseResults(&responses)

		require.NotContains(t, result[0], "errorCode")
		require.NotContains(t, result[0], "httpStatus")
	})

	t.Run("results keep the order of the responses", func(t *testing.T) {
		responses := []*messaging.SendResponse{
			{Success: true, MessageID: "first"},
			{Success: false, Error: errors.New("boom")},
			{Success: true, MessageID: "third"},
		}

		result := *toFcmResponseResults(&responses)

		require.Len(t, result, 3)
		require.Equal(t, "first", result[0]["messageID"])
		require.Equal(t, "boom", result[1]["error"])
		require.Equal(t, "third", result[2]["messageID"])
	})
}

func TestFcmErrorCode(t *testing.T) {
	// The Admin SDK only recognises errors it built itself from an FCM response,
	// so anything else must not be classified as an FCM condition. Getting this
	// wrong would let a transport failure unregister a healthy token.
	t.Run("returns empty for errors that did not come from FCM", func(t *testing.T) {
		for _, err := range []error{
			errors.New("connection reset"),
			errors.New("Device unregistered."),
			errors.New("Requested entity was not found."),
		} {
			require.Equal(t, "", fcmErrorCode(err))
		}
	})
}

func TestHTTPStatus(t *testing.T) {
	t.Run("returns empty for errors without an HTTP response", func(t *testing.T) {
		require.Equal(t, "", httpStatus(errors.New("connection reset")))
	})
}
