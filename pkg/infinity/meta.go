package infinity

import (
	"context"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/yesoreyeram/grafana-infinity-datasource/pkg/models"
)

type CustomMeta struct {
	Query                  models.Query  `json:"query"`
	Data                   any           `json:"data"`
	ResponseCodeFromServer int           `json:"responseCodeFromServer"`
	Duration               time.Duration `json:"duration"`
	Error                  string        `json:"error"`
}

func GetDummyFrame(query models.Query) *data.Frame {
	frameName := query.RefID
	if frameName == "" {
		frameName = "response"
	}
	frame := data.NewFrame(frameName)
	frame.Meta = &data.FrameMeta{
		ExecutedQueryString: "This feature is not available for this type of query yet",
		Custom: &CustomMeta{
			Query:                  query,
			Data:                   query.Data,
			ResponseCodeFromServer: 0,
			Error:                  "",
		},
	}
	return frame
}

func WrapMetaForInlineQuery(frame *data.Frame, err error, query models.Query) (*data.Frame, error) {
	if frame == nil {
		frame = data.NewFrame(query.RefID)
	}
	customMeta := &CustomMeta{Query: query, Data: query.Data, ResponseCodeFromServer: 0}
	if err != nil {
		customMeta.Error = err.Error()
	}
	frame.Meta = &data.FrameMeta{
		ExecutedQueryString: "This feature is not available for this type of query yet",
		Custom:              customMeta,
	}
	return frame, err
}

func WrapMetaForRemoteQuery(ctx context.Context, frame *data.Frame, err error, query models.Query) (*data.Frame, error) {
	if frame == nil {
		frame = data.NewFrame(query.RefID)
	}
	meta := frame.Meta
	if meta == nil {
		customMeta := &CustomMeta{Query: query, Data: query.Data, ResponseCodeFromServer: 0}
		if err != nil {
			customMeta.Error = err.Error()
		}
		frame.Meta = &data.FrameMeta{Custom: customMeta}
	}
	return frame, err
}
