package fcm

import (
	"encoding/json"
	"net/http"
	"strconv"

	"firebase.google.com/go/v4/errorutils"
	messaging "firebase.google.com/go/v4/messaging"
)

func (this *FcmMsg) makeMulticastMessageData() (*map[string]string, bool) {
	if this.Data == nil {
		emptyMap := make(map[string]string)
		return &emptyMap, true
	}

	data, ok := this.Data.(map[string]interface{})
	dataMap := make(map[string]string)
	if !ok {
		return nil, false
	}

	titleField, ok := data["title"]
	if ok {
		dataMap["title"] = titleField.(string)
	}

	bodyField, ok := data["body"]
	if ok {
		dataMap["body"] = bodyField.(string)
	}

	itemTypeField, ok := data["item_type"]
	if ok {
		dataMap["item_type"] = itemTypeField.(string)
	}

	itemIdField, ok := data["item_id"]
	if ok {
		dataMap["item_id"] = itemIdField.(string)
	}

	deepLinkField, ok := data["deeplink"]
	if ok {
		dataMap["deeplink"] = deepLinkField.(string)
	}

	imageUrlField, ok := data["image_url"]
	if ok {
		dataMap["image_url"] = imageUrlField.(string)
	}

	soundField, ok := data["sound"]
	if ok {
		dataMap["sound"] = soundField.(string)
	}

	actorNicknameField, ok := data["actor_nickname"]
	if ok {
		dataMap["actor_nickname"] = actorNicknameField.(string)
	}

	badgeCountField, ok := data["badge_count"]
	if ok {
		dataMap["badge_count"] = strconv.Itoa(int(badgeCountField.(float64)))
	}

	actionsField, ok := data["actions"]
	if ok {
		bytes, err := json.Marshal(actionsField)
		if err == nil {
			dataMap["actions"] = string(bytes)
		}
	}

	trackingPayloadField, ok := data["tracking_payload"]
	if ok {
		bytes, err := json.Marshal(trackingPayloadField)
		if err == nil {
			dataMap["tracking_payload"] = string(bytes)
		}
	}

	return &dataMap, true
}

func addImageURLToMulticastMessage(multicastMessage *messaging.MulticastMessage, imageURL string) (*messaging.MulticastMessage) {	
	multicastMessage.APNS.FCMOptions = &messaging.APNSFCMOptions{
		ImageURL: imageURL,
	}
	multicastMessage.APNS.Payload.Aps.MutableContent = true
	multicastMessage.Android = &messaging.AndroidConfig{
		Notification: &messaging.AndroidNotification{
			ImageURL: imageURL,
		},
	}
	return multicastMessage
}

func toFcmRespStatus(resp *messaging.BatchResponse) *FcmResponseStatus {
	var ok bool
	var statusCode int = http.StatusInternalServerError

	if resp.SuccessCount > 0 {
		ok = true
		statusCode = http.StatusOK
	}

	if resp.FailureCount > 0 {
		// NOTE: With Ok set to false bonito will try to inspect responses and handle errors.
		ok = false
	}

	status := FcmResponseStatus{
		Ok:            ok,
		StatusCode:    statusCode,
		Success:       resp.SuccessCount,
		Fail:          resp.FailureCount,
		Canonical_ids: 0, // TODO Where does it come from? Is it needed?
		Results:       *toFcmResponseResults(&resp.Responses),
	}

	return &status
}

func toFcmResponseResults(original *[]*messaging.SendResponse) *[]map[string]string {
	var result []map[string]string
	var elem map[string]string

	for _, resp := range *original {
		elem = map[string]string{
			"success":   strconv.FormatBool(resp.Success),
			"messageID": resp.MessageID,
		}

		optErr := resp.Error
		if optErr != nil {
			elem["error"] = optErr.Error()

			if code := fcmErrorCode(optErr); code != "" {
				elem["errorCode"] = code
			}

			if status := httpStatus(optErr); status != "" {
				elem["httpStatus"] = status
			}
		}

		result = append(result, elem)
	}

	return &result
}

// fcmErrorCode maps a send error to the stable FCM v1 error code that FCM
// returns in the FcmError details of its error response.
//
// The "error" entry above carries the error message instead, which is free-form
// text chosen by the server - it has no stable contract and callers must not
// classify errors by matching on it.
//
// Returns an empty string when the error carries no FCM error code, for example
// a transport failure or a response without FcmError details. Callers should
// fall back to the message in that case.
//
// err must be the error exactly as the SDK produced it. The predicates below
// type-assert on the SDK's concrete error type rather than using errors.As, so
// wrapping err anywhere on its way here turns every code into an empty string
// with no compile error and no test failure.
func fcmErrorCode(err error) string {
	// These are mutually exclusive - each predicate compares against a single
	// error code, so the order of the cases does not matter.
	switch {
	case messaging.IsUnregistered(err):
		return "UNREGISTERED"
	case messaging.IsSenderIDMismatch(err):
		return "SENDER_ID_MISMATCH"
	case messaging.IsQuotaExceeded(err):
		return "QUOTA_EXCEEDED"
	case messaging.IsUnavailable(err):
		return "UNAVAILABLE"
	case messaging.IsInternal(err):
		return "INTERNAL"
	case messaging.IsInvalidArgument(err):
		return "INVALID_ARGUMENT"
	case messaging.IsThirdPartyAuthError(err):
		// Covers both THIRD_PARTY_AUTH_ERROR and the older APNS_AUTH_ERROR,
		// which the SDK treats as the same condition.
		return "THIRD_PARTY_AUTH_ERROR"
	}

	return ""
}

// httpStatus returns the status code of the response that produced err, or an
// empty string if err did not come from an HTTP response. It gives callers a
// coarse signal when no FCM error code is available.
//
// As with fcmErrorCode, err must not be wrapped - errorutils.HTTPResponse also
// type-asserts on the SDK's concrete error type.
func httpStatus(err error) string {
	if resp := errorutils.HTTPResponse(err); resp != nil {
		return strconv.Itoa(resp.StatusCode)
	}

	return ""
}
