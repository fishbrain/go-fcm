package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	messaging "firebase.google.com/go/v4/messaging"
	"github.com/fishbrain/go-fcm/utils"
	logging "github.com/fishbrain/logging-go"
)

const (
	// fcm_server_url fcm server url
	fcm_server_url = "https://fcm.googleapis.com/fcm/send"

	// MAX_TTL the default ttl for a notification
	MAX_TTL = 2419200
	// Priority_HIGH notification priority
	Priority_HIGH = "high"
	// Priority_NORMAL notification priority
	Priority_NORMAL = "normal"
	// retry_after_header header name
	retry_after_header = "Retry-After"
	// error_key readable error caching !
	error_key = "error"
)

var (
	// retreyableErrors whether the error is a retryable
	retreyableErrors = map[string]bool{
		"Unavailable":         true,
		"InternalServerError": true,
	}

	// fcmServerUrl for testing purposes
	fcmServerUrl = fcm_server_url
)

type MessagingClient interface {
	SendEachForMulticast(context.Context, *messaging.MulticastMessage) (*messaging.BatchResponse, error)
}

// FcmClient stores the key and the Message (FcmMsg)
type FcmClient struct {
	ApiKey  string
	Message FcmMsg
}

// FcmMsg represents fcm request message
type FcmMsg struct {
	Data                  interface{}          `json:"data,omitempty"`
	To                    string               `json:"to,omitempty"`
	RegistrationIds       []string             `json:"registration_ids,omitempty"`
	CollapseKey           string               `json:"collapse_key,omitempty"`
	Priority              string               `json:"priority,omitempty"`
	Notification          *NotificationPayload `json:"notification,omitempty"`
	ContentAvailable      bool                 `json:"content_available,omitempty"`
	DelayWhileIdle        bool                 `json:"delay_while_idle,omitempty"`
	TimeToLive            int                  `json:"time_to_live,omitempty"`
	RestrictedPackageName string               `json:"restricted_package_name,omitempty"`
	DryRun                bool                 `json:"dry_run,omitempty"`
	Condition             string               `json:"condition,omitempty"`
	MutableContent        bool                 `json:"mutable_content,omitempty"`
}

// FcmMsg represents fcm response message - (tokens and topics)
type FcmResponseStatus struct {
	Ok            bool
	StatusCode    int
	MulticastId   int64               `json:"multicast_id"`
	Success       int                 `json:"success"`
	Fail          int                 `json:"failure"`
	Canonical_ids int                 `json:"canonical_ids"`
	Results       []map[string]string `json:"results,omitempty"`
	MsgId         int64               `json:"message_id,omitempty"`
	Err           string              `json:"error,omitempty"`
	RetryAfter    string
}

// NotificationPayload notification message payload
type NotificationPayload struct {
	Title            string `json:"title,omitempty"`
	Body             string `json:"body,omitempty"`
	Icon             string `json:"icon,omitempty"`
	Sound            string `json:"sound,omitempty"`
	Badge            string `json:"badge,omitempty"`
	Image            string `json:"image,omitempty"`
	Tag              string `json:"tag,omitempty"`
	Color            string `json:"color,omitempty"`
	ClickAction      string `json:"click_action,omitempty"`
	BodyLocKey       string `json:"body_loc_key,omitempty"`
	BodyLocArgs      string `json:"body_loc_args,omitempty"`
	TitleLocKey      string `json:"title_loc_key,omitempty"`
	TitleLocArgs     string `json:"title_loc_args,omitempty"`
	AndroidChannelID string `json:"android_channel_id,omitempty"`
}

var authAndGetFcmClient = utils.AuthorizeAndGetfcmClientFromKey

// NewFcmClient init and create fcm client
func NewFcmClient(apiKey string) *FcmClient {
	fcmc := new(FcmClient)
	fcmc.ApiKey = apiKey

	return fcmc
}

// NewFcmTopicMsg sets the targeted token/topic and the data payload
func (this *FcmClient) NewFcmTopicMsg(to string, body map[string]string) *FcmClient {
	this.NewFcmMsgTo(to, body)

	return this
}

// NewFcmMsgTo sets the targeted token/topic and the data payload
func (this *FcmClient) NewFcmMsgTo(to string, body interface{}) *FcmClient {
	this.Message.To = to
	this.Message.Data = body

	return this
}

// SetMsgData sets data payload
func (this *FcmClient) SetMsgData(body interface{}) *FcmClient {
	this.Message.Data = body

	return this
}

// NewFcmRegIdsMsg gets a list of devices with data payload
func (this *FcmClient) NewFcmRegIdsMsg(list []string, body interface{}) *FcmClient {
	this.newDevicesList(list)
	this.Message.Data = body

	return this
}

// newDevicesList init the devices list
func (this *FcmClient) newDevicesList(list []string) *FcmClient {
	this.Message.RegistrationIds = make([]string, len(list))
	copy(this.Message.RegistrationIds, list)

	return this
}

// AppendDevices adds more devices/tokens to the Fcm request
func (this *FcmClient) AppendDevices(list []string) *FcmClient {
	this.Message.RegistrationIds = append(this.Message.RegistrationIds, list...)

	return this
}

// apiKeyHeader generates the value of the Authorization key
func (this *FcmClient) apiKeyHeader() string {
	return fmt.Sprintf("key=%v", this.ApiKey)
}

// sendOnce send a single request to fcm
func (this *FcmClient) sendOnce() (*FcmResponseStatus, error) {
	jsonByte, err := json.Marshal(this.Message)
	if err != nil {
		return &FcmResponseStatus{}, err
	}

	request, err := http.NewRequest("POST", fcmServerUrl, bytes.NewBuffer(jsonByte))
	request.Header.Set("Authorization", this.apiKeyHeader())
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return &FcmResponseStatus{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return &FcmResponseStatus{}, err
	}

	fcmRespStatus := &FcmResponseStatus{
		StatusCode: response.StatusCode,
		RetryAfter: response.Header.Get(retry_after_header),
	}

	if response.StatusCode != 200 {
		fcmRespStatus.Err = string(body)
		return fcmRespStatus, nil
	}

	err = fcmRespStatus.parseStatusBody(body)
	if err != nil {
		return fcmRespStatus, err
	}

	fcmRespStatus.Ok = true

	return fcmRespStatus, nil
}

// Send to fcm
func (this *FcmClient) Send() (*FcmResponseStatus, error) {

	if this.Message.DryRun {
		logging.Log.Info("Dry run mode enabled")

		client, err := utils.AuthorizeAndGetFirebaseMessagingClient()
		if err != nil {
			logging.Log.Infof("Error getting messaging client with embedded key: %s", err)
		}
		logging.Log.Infof("FCM Client with embedded key: %v", client)

		response, err := this.sendOnceFirebaseAdminGo(client)
		if response.Ok {
			logging.Log.Infof("Success sending message with embedded key")
			return response, err
		}else {
			logging.Log.Infof("Error sending message with embedded key %v", response)
		}

		client, err = utils.AuthorizeAndGetfcmClientFromIdPoolKey()
		if err != nil {
			logging.Log.Infof("Error getting messaging client with key as a map: %s", err)
		}
		logging.Log.Infof("FCM Client with key as a map: %v", client)

		response, err = this.sendOnceFirebaseAdminGo(client)
		
		if response.Ok{
			logging.Log.Infof("Success sending message with key as a map")
			return response, err
		}else {
			logging.Log.Infof("Error sending message with key as a map %v", response)
		}
	}

	client, err := authAndGetFcmClient()
		if err != nil {
			logging.Log.Errorf("Error getting messaging client: %s", err)
			return &FcmResponseStatus{}, err
		}

	return this.sendOnceFirebaseAdminGo(client)
}

func (fcmClient *FcmClient) sendOnceFirebaseAdminGo(client MessagingClient) (*FcmResponseStatus, error) {
	multicastMessage, ok := fcmClient.Message.makeMulticastMessageData()
	if !ok {
		return nil, errors.New("error building multicast message for Firebase Admin Go library")
	}

	message := &messaging.MulticastMessage{
		Data:   *multicastMessage,
		Tokens: fcmClient.Message.RegistrationIds,
	}
	
	if fcmClient.Message.Notification != nil {
		message.Notification = &messaging.Notification{
			Title: fcmClient.Message.Notification.Title,
			Body:  fcmClient.Message.Notification.Body,
		}

		message.APNS = &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: fcmClient.Message.Notification.asAPS(),
			},
		} 

		imageUrlField := fcmClient.Message.Notification.Image
		if imageUrlField != "" {
			message = addImageURLToMulticastMessage(message, imageUrlField)
			}
	}

	batchResponse, err := client.SendEachForMulticast(context.Background(), message)
	if err != nil {
		logging.Log.Errorf("Error sending message: %s", err)
		return &FcmResponseStatus{}, err
	}

	fcmRespStatus := toFcmRespStatus(batchResponse)

	return fcmRespStatus, nil
}

func (n *NotificationPayload) asAPS() *messaging.Aps {
	badge, err := strconv.Atoi(n.Badge)
	if err != nil {
		return &messaging.Aps{
			Sound: n.Sound,
			Alert: &messaging.ApsAlert{
				Body: n.Body,
				Title: n.Title,
			},
		}
	}
	return &messaging.Aps{
		Badge: &badge,
		Sound: n.Sound,
		Alert: &messaging.ApsAlert{
			Body: n.Body,
			Title: n.Title,
		},
	}
}

// parseStatusBody parse FCM response body
func (this *FcmResponseStatus) parseStatusBody(body []byte) error {
	if err := json.Unmarshal([]byte(body), &this); err != nil {
		return err
	}

	return nil
}

// SetPriority Sets the priority of the message.
// Priority_HIGH or Priority_NORMAL
func (this *FcmClient) SetPriority(p string) *FcmClient {
	if p == Priority_HIGH {
		this.Message.Priority = Priority_HIGH
	} else {
		this.Message.Priority = Priority_NORMAL
	}

	return this
}

// SetCollapseKey This parameter identifies a group of messages
// (e.g., with collapse_key: "Updates Available") that can be collapsed,
// so that only the last message gets sent when delivery can be resumed.
// This is intended to avoid sending too many of the same messages when the
// device comes back online or becomes active (see delay_while_idle).
func (this *FcmClient) SetCollapseKey(val string) *FcmClient {
	this.Message.CollapseKey = val

	return this
}

// SetNotificationPayload sets the notification payload based on the specs
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
func (this *FcmClient) SetNotificationPayload(payload *NotificationPayload) *FcmClient {
	this.Message.Notification = payload

	return this
}

// SetContentAvailable On iOS, use this field to represent content-available
// in the APNS payload. When a notification or message is sent and this is set
// to true, an inactive client app is awoken. On Android, data messages wake
// the app by default. On Chrome, currently not supported.
func (this *FcmClient) SetContentAvailable(isContentAvailable bool) *FcmClient {
	this.Message.ContentAvailable = isContentAvailable

	return this
}

// SetDelayWhileIdle When this parameter is set to true, it indicates that
// the message should not be sent until the device becomes active.
// The default value is false.
func (this *FcmClient) SetDelayWhileIdle(isDelayWhileIdle bool) *FcmClient {
	this.Message.DelayWhileIdle = isDelayWhileIdle

	return this
}

// SetTimeToLive This parameter specifies how long (in seconds) the message
// should be kept in FCM storage if the device is offline. The maximum time
// to live supported is 4 weeks, and the default value is 4 weeks.
// For more information, see
// https://firebase.google.com/docs/cloud-messaging/concept-options#ttl
func (this *FcmClient) SetTimeToLive(ttl int) *FcmClient {
	if ttl > MAX_TTL {
		this.Message.TimeToLive = MAX_TTL
	} else {
		this.Message.TimeToLive = ttl
	}

	return this
}

// SetRestrictedPackageName This parameter specifies the package name of the
// application where the registration tokens must match in order to
// receive the message.
func (this *FcmClient) SetRestrictedPackageName(pkg string) *FcmClient {
	this.Message.RestrictedPackageName = pkg

	return this
}

// SetDryRun This parameter, when set to true, allows developers to test
// a request without actually sending a message.
// The default value is false
func (this *FcmClient) SetDryRun(drun bool) *FcmClient {
	this.Message.DryRun = drun

	return this
}

// SetMutableContent Currently for iOS 10+ devices only. On iOS,
// use this field to represent mutable-content in the APNs payload.
// When a notification is sent and this is set to true, the content
// of the notification can be modified before it is displayed,
// using a Notification Service app extension.
// This parameter will be ignored for Android and web.
func (this *FcmClient) SetMutableContent(mc bool) *FcmClient {
	this.Message.MutableContent = mc

	return this
}

// PrintResults prints the FcmResponseStatus results for fast using and debugging
func (this *FcmResponseStatus) PrintResults() {
	fmt.Println("Status Code   :", this.StatusCode)
	fmt.Println("Success       :", this.Success)
	fmt.Println("Fail          :", this.Fail)
	fmt.Println("Canonical_ids :", this.Canonical_ids)
	fmt.Println("Topic MsgId   :", this.MsgId)
	fmt.Println("Topic Err     :", this.Err)
	for i, val := range this.Results {
		fmt.Printf("Result(%d)> \n", i)
		for k, v := range val {
			fmt.Println("\t", k, " : ", v)
		}
	}
}

// IsTimeout check whether the response timeout based on http response status
// code and if any error is retryable
func (this *FcmResponseStatus) IsTimeout() bool {
	if this.StatusCode >= 500 {
		return true
	} else if this.StatusCode == 200 {
		for _, val := range this.Results {
			for k, v := range val {
				if k == error_key && retreyableErrors[v] {
					return true
				}
			}
		}
	}

	return false
}

// GetRetryAfterTime converts the retry after response header
// to a time.Duration
func (this *FcmResponseStatus) GetRetryAfterTime() (t time.Duration, e error) {
	t, e = time.ParseDuration(this.RetryAfter)

	return
}

// SetCondition to set a logical expression of conditions that determine the message target
func (this *FcmClient) SetCondition(condition string) *FcmClient {
	this.Message.Condition = condition

	return this
}
