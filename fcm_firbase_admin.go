package fcm

import (
	"encoding/json"
	"net/http"
	"strconv"

	messaging "firebase.google.com/go/v4/messaging"
	logging "github.com/fishbrain/logging-go"
)

func (this *FcmMsg) makeMulticastMessageData() (*map[string]string, bool) {
	if this.Data == nil {
		emptyMap := make(map[string]string)
		return &emptyMap, true
	}

	data, ok := this.Data.(map[string]interface{})
	if !ok {
		return nil, false
	}

	var (
		title                     string
		body                      string
		itemType                  string
		itemID                    string
		deepLink                  string
		imageURL                  string
		sound                     string
		actorNickname             string
		badgeCount                int
		actionsSerialized         string
		trackingPayloadSerialized string
	)

	titleField, ok := data["title"]
	if ok {
		title = titleField.(string)
	}

	bodyField, ok := data["body"]
	if ok {
		body = bodyField.(string)
	}

	itemTypeField, ok := data["item_type"]
	if ok {
		itemType = itemTypeField.(string)
	}

	itemIdField, ok := data["item_id"]
	if ok {
		itemID = itemIdField.(string)
	}

	deepLinkField, ok := data["deeplink"]
	if ok {
		deepLink = deepLinkField.(string)
	}

	imageUrlField, ok := data["image_url"]
	if ok {
		imageURL = imageUrlField.(string)
	}

	soundField, ok := data["sound"]
	if ok {
		sound = soundField.(string)
	}

	actorNicknameField, ok := data["actor_nickname"]
	if ok {
		actorNickname = actorNicknameField.(string)
	}

	badgeCountField, ok := data["badge_count"]
	if ok {
		badgeCount = int(badgeCountField.(float64))
	}

	actionsField, ok := data["actions"]
	if ok {
		bytes, err := json.Marshal(actionsField)
		if err == nil {
			actionsSerialized = string(bytes)
		}
	}

	trackingPayloadField, ok := data["tracking_payload"]
	if ok {
		bytes, err := json.Marshal(trackingPayloadField)
		if err == nil {
			trackingPayloadSerialized = string(bytes)
		}
	}

	dataMap := make(map[string]string)
	dataMap["title"] = title
	dataMap["body"] = body
	dataMap["item_type"] = itemType
	dataMap["item_id"] = itemID
	dataMap["deeplink"] = deepLink
	dataMap["image_url"] = imageURL
	dataMap["sound"] = sound
	dataMap["actor_nickname"] = actorNickname
	dataMap["badge_count"] = strconv.Itoa(badgeCount)
	dataMap["actions"] = actionsSerialized
	dataMap["tracking_payload"] = trackingPayloadSerialized

	return &dataMap, true
}

func toFcmRespStatus(resp *messaging.BatchResponse) *FcmResponseStatus {
	var ok bool
	var statusCode int = http.StatusInternalServerError

	if resp.SuccessCount > 0 {
		ok = true
		statusCode = http.StatusOK
	}
	logging.Log.Infof("Batch response: %v", resp)

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
			"Success":   strconv.FormatBool(resp.Success),
			"MessageID": resp.MessageID,
		}

		optErr := resp.Error
		if optErr != nil {
			elem["Error"] = optErr.Error()
		}

		result = append(result, elem)
	}

	return &result
}
