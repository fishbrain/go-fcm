package fcm

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		}

		result = append(result, elem)
	}

	return &result
}
