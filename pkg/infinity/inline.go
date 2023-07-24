package infinity

import (
	"context"
	"errors"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/yesoreyeram/grafana-infinity-datasource/pkg/models"
	"github.com/yesoreyeram/grafana-plugins/lib/go/jsonframer"
)

func GetFrameForInlineSources(query models.Query) (*data.Frame, error) {
	frame := GetDummyFrame(query)
	if query.Type == models.QueryTypeGROQ || query.Type == models.QueryTypeUQL {
		return frame, nil
	}
	if query.Parser != "backend" {
		return frame, nil
	}
	switch query.Type {
	case models.QueryTypeCSV, models.QueryTypeTSV:
		frame, err := GetCSVBackendResponse(query.Data, query)
		if err != nil {
			return frame, err
		}
		return PostProcessFrame(context.Background(), frame, query)
	case models.QueryTypeXML, models.QueryTypeHTML:
		frame, err := GetXMLBackendResponse(query.Data, query)
		if err != nil {
			return frame, err
		}
		return PostProcessFrame(context.Background(), frame, query)
	case models.QueryTypeJSON, models.QueryTypeGraphQL:
		columns := []jsonframer.ColumnSelector{}
		for _, c := range query.Columns {
			columns = append(columns, jsonframer.ColumnSelector{
				Selector:   c.Selector,
				Alias:      c.Text,
				Type:       c.Type,
				TimeFormat: c.TimeStampFormat,
			})
		}
		newFrame, err := jsonframer.ToFrame(query.Data, jsonframer.FramerOptions{
			FrameName:    query.RefID,
			RootSelector: query.RootSelector,
			Columns:      columns,
		})
		if err != nil {
			return frame, err
		}
		if newFrame != nil {
			frame.Fields = append(frame.Fields, newFrame.Fields...)
		}
		return PostProcessFrame(context.Background(), frame, query)
	default:
		return frame, errors.New("unknown backend query type")
	}
}
