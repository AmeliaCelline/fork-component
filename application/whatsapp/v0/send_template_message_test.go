package whatsapp

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	token = "test_token"
)

type MockWhatsappClientSendTemplate struct{}

func (c *MockWhatsappClientSendTemplate) SendMessageAPI(req interface{}, resp interface{}, PhoneNumberID string) (interface{}, error) {
	jsonData := `{
		"messaging_product": "whatsapp",
		"contacts": [
			{
				"input": "886901234567",
				"wa_id": "886901234567"
			}
		],
		"messages": [
			{
				"id": "wamid.HBgMODg2OTg3MTIyNjY4FQIAERgSMjREQjI0Q0FDQkZCQjU1QjYwAA==",
				"message_status": "accepted"
			}
		]
	}`

	err := json.Unmarshal([]byte(jsonData), resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func TestComponent_ExecuteSendTemplateMessageTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{}
	connector := Init(bc)

	wantOutput := TaskSendTemplateMessageOutput{
		WaID:          "886901234567",
		ID:            "wamid.HBgMODg2OTg3MTIyNjY4FQIAERgSMjREQjI0Q0FDQkZCQjU1QjYwAA==",
		MessageStatus: "accepted",
	}

	testcases := []struct {
		name       string
		task       string
		input      interface{}
		wantOutput TaskSendTemplateMessageOutput
		wantErr    string
	}{
		{
			name: "ok - send text-based template message",
			task: taskSendTextBasedTemplateMessage,
			input: TaskSendTextBasedTemplateMessageInput{
				PhoneNumberID:    "012345678901234",
				To:               "886901234567",
				TemplateName:     "random_text_template",
				LanguageCode:     "en_us",
				HeaderParameters: []string{"headerparameter1, headerparameter2"},
				BodyParameters:   []string{"bodyparameter1, bodyparameter2"},
				ButtonParameters: []string{"0;copy_code;randomvalue"},
			},
			wantOutput: wantOutput,
			wantErr:    "",
		},
		{
			name: "nok - send text-based template message: button format is incorrect",
			task: taskSendTextBasedTemplateMessage,
			input: TaskSendTextBasedTemplateMessageInput{
				PhoneNumberID:    "012345678901234",
				To:               "886901234567",
				TemplateName:     "random_text_template",
				LanguageCode:     "en_us",
				HeaderParameters: []string{"headerparameter1, headerparameter2"},
				BodyParameters:   []string{"bodyparameter1, bodyparameter2"},
				ButtonParameters: []string{"0;randomvalue"},
			},
			wantOutput: wantOutput,
			wantErr:    "format is wrong, it must be 'button_index;button_type;value_of_the_parameter'. Example: 0;quick_reply;randomvalue",
		},
		{
			name: "ok - send media-based template message",
			task: taskSendMediaBasedTemplateMessage,
			input: TaskSendMediaBasedTemplateMessageInput{
				PhoneNumberID:    "012345678901234",
				To:               "886901234567",
				TemplateName:     "random_document_template",
				LanguageCode:     "en_us",
				MediaType:        "document",
				IDOrLink:         "https://www.random.com/random.pdf",
				Filename:         "random.pdf",
				BodyParameters:   []string{"bodyparameter1, bodyparameter2"},
				ButtonParameters: []string{"0;url;websiteurl"},
			},
			wantOutput: wantOutput,
			wantErr:    "",
		},
		{
			name: "ok - send location-based template message",
			task: taskSendLocationBasedTemplateMessage,
			input: TaskSendLocationBasedTemplateMessageInput{
				PhoneNumberID:    "012345678901234",
				To:               "886901234567",
				TemplateName:     "random_location_template",
				LanguageCode:     "en_us",
				Latitude:         25.123456,
				Longitude:        121.123456,
				LocationName:     "A random location",
				Address:          "A random address",
				BodyParameters:   []string{"bodyparameter1, bodyparameter2"},
				ButtonParameters: []string{"0;quick_reply;randompayload"},
			},
			wantOutput: wantOutput,
			wantErr:    "",
		},
		{
			name: "ok - send authentication template message",
			task: taskSendAuthenticationTemplateMessage,
			input: TaskSendAuthenticationTemplateMessageInput{
				PhoneNumberID:   "012345678901234",
				To:              "886901234567",
				TemplateName:    "random_authentication_template",
				LanguageCode:    "en_us",
				OneTimePassword: "a12345",
			},
			wantOutput: wantOutput,
			wantErr:    "",
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {

			setup, err := structpb.NewStruct(map[string]any{
				"token": token,
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: tc.task},
				client:             &MockWhatsappClientSendTemplate{},
			}

			switch tc.task {
			case taskSendTextBasedTemplateMessage:
				e.execute = e.SendTextBasedTemplateMessage
			case taskSendMediaBasedTemplateMessage:
				e.execute = e.SendMediaBasedTemplateMessage
			case taskSendLocationBasedTemplateMessage:
				e.execute = e.SendLocationBasedTemplateMessage
			case taskSendAuthenticationTemplateMessage:
				e.execute = e.SendAuthenticationTemplateMessage
			}

			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			pbOut, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			c.Assert(err, qt.IsNil)

			outJson, err := protojson.Marshal(pbOut[0])
			c.Assert(err, qt.IsNil)

			c.Check(outJson, qt.JSONEquals, tc.wantOutput)
		})
	}
}
