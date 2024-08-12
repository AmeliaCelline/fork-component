package freshdesk

import (
	"encoding/base64"
	"fmt"

	"github.com/instill-ai/component/internal/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(setup *structpb.Struct, logger *zap.Logger) *FreshdeskClient {
	basePath := fmt.Sprintf("https://%s.freshdesk.com/api", getDomain(setup))

	c := httpclient.New("Freshdesk", basePath+"/"+version,
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.Header.Set("Authorization", getAPIKey(setup))

	w := &FreshdeskClient{httpclient: c}

	return w
}

type errBody struct {
	Description string `json:"description"`
	Errors      []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"errors"`
}

func (e errBody) Message() string {
	var errReturn string
	fmt.Println("ERRRORRR", e.Errors)
	for index, err := range e.Errors {
		if index > 0 {
			errReturn += ". "
		}

		errReturn += err.Message
		if err.Field != "" {
			errReturn += ", field: " + err.Field
		}
		if err.Code != "" {
			errReturn += ", code: " + err.Code
		}
	}

	return errReturn
}

func getAPIKey(setup *structpb.Struct) string {
	apiKey := setup.GetFields()["api-key"].GetStringValue()

	// In Freshdesk, the format is api-key:X. Afterward, it needs to be encoded in base64
	encodedKey := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:X", apiKey)))
	return encodedKey
}

func getDomain(setup *structpb.Struct) string {
	return setup.GetFields()["domain"].GetStringValue()
}

type FreshdeskClient struct {
	httpclient *httpclient.Client
}

type FreshdeskInterface interface {
	GetTicket(ticketID int64) (*TaskGetTicketResponse, error)
	CreateTicket(req *TaskCreateTicketReq) (*TaskCreateTicketResponse, error)
}
