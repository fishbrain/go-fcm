package fcm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	messaging "firebase.google.com/go/v4/messaging"
	logging "github.com/fishbrain/logging-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fcmMock struct {
	mock.Mock
}

func TestMain(m *testing.M) {
	logging.Init(logging.LoggingConfig{})
	os.Exit(m.Run())
}

func (m *fcmMock) SendEachForMulticast(ctx context.Context, mm *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	args := m.Called(ctx, mm)
	return args.Get(0).(*messaging.BatchResponse), args.Error(1)
}

func TestTopicHandle_1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	c.NewFcmMsgTo("/topics/topicName", data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestImage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	notificationPayload := NotificationPayload{
		Title: "title - foo",
		Body:  "body - bar",
		Image: "https://example.com/img.jpg",
	}
	c.SetNotificationPayload(&notificationPayload)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestTopicHandle_2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	c.NewFcmTopicMsg("/topics/topicName", data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestTopicHandle_3(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	data2 := map[string]string{
		"msg": "Hello bits",
	}

	c.NewFcmTopicMsg("/topics/topicName", data)

	c.SetMsgData(data2)
	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestRegIdHandle_1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	ids := []string{
		"token0",
		"token1",
		"token2",
	}

	c.NewFcmRegIdsMsg(ids, data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}

	if res.Success != 2 || res.Fail != 1 {
		t.Error("Parsing Success or Fail error")
	}
}

func TestRegIdHandle_2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	ids := []string{
		"token0",
	}

	xds := []string{
		"token1",
		"token2",
	}

	c.newDevicesList(ids)

	c.SetMsgData(data)

	c.AppendDevices(xds)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}

	if res.Success != 2 || res.Fail != 1 {
		t.Error("Parsing Success or Fail error")
	}
}

func chgUrl(ts *httptest.Server) {
	fcmServerUrl = ts.URL
}

func topicHandle(w http.ResponseWriter, r *http.Request) {
	result := `{"message_id":6985435902064854329}`

	fmt.Fprintln(w, result)
}

func regIdHandle(w http.ResponseWriter, r *http.Request) {
	result := `{"multicast_id":1003859738309903334,"success":2,"failure":1,"canonical_ids":0,"results":[{"message_id":"0:1448128667408487%ecaaa23db3fd7efd"},{"message_id":"0:1468135657607438%ecafacddf9ff8ead"},{"error":"InvalidRegistration"}]}`
	fmt.Fprintln(w, result)

}

func TestSendFirebase(t *testing.T) {
	logging.Init(logging.LoggingConfig{})
	srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	ids := []string{
		"token0",
	}

	xds := []string{
		"token1",
		"token2",
	}

	c.newDevicesList(ids)

	c.SetMsgData(data)

	c.AppendDevices(xds)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}

	if res.Success != 2 || res.Fail != 1 {
		t.Error("Parsing Success or Fail error")
	}
}

func TestSendOnceFirebaseAdminGo_SuccessResponse(t *testing.T) {
	logging.Init(logging.LoggingConfig{})
	c := NewFcmClient("key")

	notificationPayload := NotificationPayload{
		Title: "title - foo",
		Body:  "body - bar",
		Image: "https://example.com/img.jpg",
	}
	c.SetNotificationPayload(&notificationPayload)

	messagingClientMock := new(fcmMock)
	mockCall := messagingClientMock.On("SendEachForMulticast", mock.Anything, mock.Anything).Return(
		&messaging.BatchResponse{
			SuccessCount: 1,
			FailureCount: 0,
			Responses: []*messaging.SendResponse{
				{
					Success:   true,
					MessageID: "123",
					Error:     nil,
				},
			},
		},
		nil,
	)

	fcmRespStatus, err := c.sendOnceFirebaseAdminGo(messagingClientMock)

	messagingClientMock.AssertExpectations(t)
	mockCall.Unset()

	require.Nil(t, err)
	require.Equal(t, &FcmResponseStatus{
		Ok:            true,
		StatusCode:    http.StatusOK,
		MulticastId:   0,
		Success:       1,
		Fail:          0,
		Canonical_ids: 0,
		Results: []map[string]string{
			{
				"messageID": "123",
				"success":   "true",
			},
		},
		MsgId: 0,
	}, fcmRespStatus)
}

func TestSendOnceFirebaseAdminGo_SuccessResponseWhenNoNotificationPayload(t *testing.T) {
	logging.Init(logging.LoggingConfig{})
	c := NewFcmClient("key")

	messagingClientMock := new(fcmMock)
	mockCall := messagingClientMock.On("SendEachForMulticast", mock.Anything, mock.Anything).Return(
		&messaging.BatchResponse{
			SuccessCount: 1,
			FailureCount: 0,
			Responses: []*messaging.SendResponse{
				{
					Success:   true,
					MessageID: "123",
					Error:     nil,
				},
			},
		},
		nil,
	)

	fcmRespStatus, err := c.sendOnceFirebaseAdminGo(messagingClientMock)

	messagingClientMock.AssertExpectations(t)
	mockCall.Unset()

	require.Nil(t, err)
	require.Equal(t, &FcmResponseStatus{
		Ok:            true,
		StatusCode:    http.StatusOK,
		MulticastId:   0,
		Success:       1,
		Fail:          0,
		Canonical_ids: 0,
		Results: []map[string]string{
			{
				"messageID": "123",
				"success":   "true",
			},
		},
		MsgId: 0,
	}, fcmRespStatus)
}

func TestSendOnceFirebaseAdminGo_BadResponse(t *testing.T) {
	c := NewFcmClient("key")

	notificationPayload := NotificationPayload{
		Title: "title - foo",
		Body:  "body - bar",
		Image: "https://example.com/img.jpg",
	}
	c.SetNotificationPayload(&notificationPayload)

	messagingClientMock := new(fcmMock)
	mockCall := messagingClientMock.On("SendEachForMulticast", mock.Anything, mock.Anything).Return(
		&messaging.BatchResponse{
			SuccessCount: 0,
			FailureCount: 1,
			Responses: []*messaging.SendResponse{
				{
					Success:   false,
					MessageID: "123",
					Error:     errors.New("something went wrong"),
				},
			},
		},
		nil,
	)

	fcmRespStatus, err := c.sendOnceFirebaseAdminGo(messagingClientMock)

	messagingClientMock.AssertExpectations(t)
	mockCall.Unset()

	require.Nil(t, err)
	require.Equal(t, &FcmResponseStatus{
		Ok:            false,
		StatusCode:    http.StatusInternalServerError,
		MulticastId:   0,
		Success:       0,
		Fail:          1,
		Canonical_ids: 0,
		Results: []map[string]string{
			{
				"messageID": "123",
				"success":   "false",
				"error":     "something went wrong",
			},
		},
		MsgId: 0,
	}, fcmRespStatus)
}

func TestSendOnceFirebaseAdminGo_MixedResponse(t *testing.T) {
	c := NewFcmClient("key")

	notificationPayload := NotificationPayload{
		Title: "title - foo",
		Body:  "body - bar",
		Image: "https://example.com/img.jpg",
	}
	c.SetNotificationPayload(&notificationPayload)

	messagingClientMock := new(fcmMock)
	mockCall := messagingClientMock.On("SendEachForMulticast", mock.Anything, mock.Anything).Return(
		&messaging.BatchResponse{
			SuccessCount: 1,
			FailureCount: 1,
			Responses: []*messaging.SendResponse{
				{
					Success:   false,
					MessageID: "123",
					Error:     errors.New("something went wrong"),
				},
				{
					Success:   true,
					MessageID: "123",
					Error:     nil,
				},
			},
		},
		nil,
	)

	fcmRespStatus, err := c.sendOnceFirebaseAdminGo(messagingClientMock)

	messagingClientMock.AssertExpectations(t)
	mockCall.Unset()

	require.Nil(t, err)
	require.Equal(t, &FcmResponseStatus{
		Ok:            false,
		StatusCode:    http.StatusOK,
		MulticastId:   0,
		Success:       1,
		Fail:          1,
		Canonical_ids: 0,
		Results: []map[string]string{
			{
				"messageID": "123",
				"success":   "false",
				"error":     "something went wrong",
			},
			{
				"messageID": "123",
				"success":   "true",
			},
		},
		MsgId: 0,
	}, fcmRespStatus)
}

func TestMakeMulticastMessageData_Nil(t *testing.T) {
	msg := FcmMsg{}
	res, ok := msg.makeMulticastMessageData()

	require.Equal(t, true, ok)
	require.Equal(t, &map[string]string{}, res)
}

func TestMakeMulticastMessageData_NotNil(t *testing.T) {
	type Action struct {
		Type  string
		Value string
	}

	msg := FcmMsg{
		Data: map[string]interface{}{
			"actions":   []Action{{Type: "Like", Value: "like"}},
			"body":      "example body",
			"item_type": "Post",
			"item_id":   "123",
		},
	}

	res, ok := msg.makeMulticastMessageData()

	require.Equal(t, true, ok)
	require.Equal(t, &map[string]string{
		"body":             "example body",
		"item_type":        "Post",
		"item_id":          "123",
		"actions":          `[{"Type":"Like","Value":"like"}]`,
		"actor_nickname":   "",
		"badge_count":      "0",
		"deeplink":         "",
		"image_url":        "",
		"sound":            "",
		"title":            "",
		"tracking_payload": "",
	}, res)
}
