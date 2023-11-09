package infinity_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yesoreyeram/grafana-infinity-datasource/pkg/infinity"
	"github.com/yesoreyeram/grafana-infinity-datasource/pkg/models"
)

func Test_getQueryURL(t *testing.T) {
	tests := []struct {
		name          string
		settings      models.InfinitySettings
		query         models.Query
		excludeSecret bool
		want          string
	}{
		{
			settings: models.InfinitySettings{},
			query: models.Query{
				URL: "0.0.0.0",
			},
			want: "0.0.0.0",
		},
		{
			settings: models.InfinitySettings{},
			query: models.Query{
				URL: "https://foo.com?key=val",
			},
			want: "https://foo.com?key=val",
		},
		{
			settings: models.InfinitySettings{
				URL: "https://foo.com",
			},
			query: models.Query{
				URL: "/hello?key=val",
			},
			want: "https://foo.com/hello?key=val",
		},
		{
			name: "should merge all the query parameters",
			settings: models.InfinitySettings{
				URL: "https://foo.com",
			},
			query: models.Query{
				URL: "/hello?key=val10&foo=bar",
				URLOptions: models.URLOptions{
					Params: []models.URLOptionKeyValuePair{
						{Key: "key", Value: "val11"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
			want: "https://foo.com/hello?foo=bar&key=val10&key=val11&key2=value2",
		},
		{
			name: "should always respect the URL in the settings",
			settings: models.InfinitySettings{
				URL: "https://one.com",
			},
			query: models.Query{
				URL: "https://two.com/hello?key=val",
			},
			want: "https://one.comhttps://two.com/hello?key=val",
		},
		{
			settings: models.InfinitySettings{
				URL: "https://foo.com",
			},
			query: models.Query{
				URL: "https://foo.com/hello?key=val",
			},
			want: "https://foo.com/hello?key=val",
		},
		{
			settings: models.InfinitySettings{
				URL: "https://foo.com",
				SecureQueryFields: map[string]string{
					"key_one": "val_one",
					"key_two": "val_two",
				},
			},
			query: models.Query{
				URL: "/hello?key=val&key_one=${__qs.key_one}&key_two=${__qs.key_two}&foo=bar",
			},
			want: "https://foo.com/hello?foo=bar&key=val&key_one=val_one&key_two=val_two",
		},
		{
			settings: models.InfinitySettings{
				URL: "https://foo.com",
				SecureQueryFields: map[string]string{
					"key_one": "val_one",
					"key_two": "val_two",
				},
			},
			query: models.Query{
				URL: "/hello?key=val&foo=bar&key_one=foo",
			},
			want: "https://foo.com/hello?foo=bar&key=val&key_one=val_one&key_two=val_two",
		},
		{
			settings: models.InfinitySettings{
				URL: "https://foo.com/${__qs.key_three}",
				SecureQueryFields: map[string]string{
					"key_one":   "val_one",
					"key_two":   "val_two",
					"key_three": "val_three",
				},
				AuthenticationMethod: "apiKey",
				ApiKeyKey:            "hello",
				ApiKeyValue:          "world",
				ApiKeyType:           "query",
			},
			query: models.Query{
				URL: "/hello?key=val&foo=bar&key_one=foo",
				URLOptions: models.URLOptions{
					Params: []models.URLOptionKeyValuePair{
						{Key: "mee", Value: "too"},
					},
				},
			},
			want: "https://foo.com/val_three/hello?foo=bar&hello=world&key=val&key_one=val_one&key_three=val_three&key_two=val_two&mee=too",
		},
		{
			settings: models.InfinitySettings{
				URL: "https://foo.com/${__qs.key_three}",
				SecureQueryFields: map[string]string{
					"key_one":   "val_one",
					"key_two":   "val_two",
					"key_three": "val_three",
				},
				AuthenticationMethod: "apiKey",
				ApiKeyKey:            "hello",
				ApiKeyValue:          "world",
				ApiKeyType:           "query",
			},
			query: models.Query{
				URL: "/hello?key=val&foo=bar&key_one=foo",
				URLOptions: models.URLOptions{
					Params: []models.URLOptionKeyValuePair{
						{Key: "mee", Value: "too"},
					},
				},
			},
			excludeSecret: true,
			want:          "https://foo.com/xxxxxxxx/hello?foo=bar&hello=xxxxxxxx&key=val&key_one=xxxxxxxx&key_three=xxxxxxxx&key_two=xxxxxxxx&mee=too",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := infinity.GetQueryURL(context.TODO(), tt.settings, tt.query, !tt.excludeSecret)
			assert.Equal(t, err, nil)
			assert.Equal(t, tt.want, u)
		})
	}
}
func TestClient_GetExecutedURL(t *testing.T) {
	tests := []struct {
		name     string
		settings models.InfinitySettings
		url      string
		command  string
		query    models.Query
	}{
		{
			query:   models.Query{URL: "https://foo.com"},
			url:     "https://foo.com",
			command: "curl -X 'GET' 'https://foo.com'",
		},
		{
			settings: models.InfinitySettings{UserName: "hello", Password: "world", BasicAuthEnabled: true},
			query:    models.Query{URL: "https://foo.com"},
			url:      "https://foo.com",
			command:  "curl -X 'GET' -H 'Authorization: Basic xxxxxxxx' 'https://foo.com'",
		},
		{
			settings: models.InfinitySettings{AuthenticationMethod: "bearerToken", BearerToken: "world2"},
			query:    models.Query{URL: "https://foo.com"},
			url:      "https://foo.com",
			command:  "curl -X 'GET' -H 'Authorization: Bearer xxxxxxxx' 'https://foo.com'",
		},
		/*
			{
				settings: models.InfinitySettings{AuthenticationMethod: "zcap", ZCapJsonPath: "https://localhost:53443/mercury/hello"},
				query:    models.Query{URL: "https://foo.com"},
				url:      "https://foo.com",
				command:  "curl -X 'GET' -H '/mercury/hello: test' 'https://foo.com'",
			},*/
		{
			settings: models.InfinitySettings{AuthenticationMethod: "zcap"},
			query:    models.Query{URL: "https://foo.com"},
			url:      "https://foo.com",
			command:  "curl -X 'GET' -H 'Key: xxxxxxxx' 'https://foo.com'",
		},

		{
			settings: models.InfinitySettings{AuthenticationMethod: "apiKey", ApiKeyType: "header", ApiKeyKey: "hello", ApiKeyValue: "world"},
			query:    models.Query{URL: "https://foo.com"},
			url:      "https://foo.com",
			command:  "curl -X 'GET' -H 'Hello: xxxxxxxx' 'https://foo.com'",
		},
		{
			settings: models.InfinitySettings{ForwardOauthIdentity: true},
			query:    models.Query{URL: "https://foo.com"},
			url:      "https://foo.com",
			command:  "curl -X 'GET' -H 'Authorization: xxxxxxxx' 'https://foo.com'",
		},
		{
			settings: models.InfinitySettings{CustomHeaders: map[string]string{"good": "bye"}, SecureQueryFields: map[string]string{"me": "too"}},
			query:    models.Query{URL: "https://foo.com?something=${__qs.me}", Type: "json", URLOptions: models.URLOptions{Method: "POST", Body: "my request body with ${__qs.me} value", Headers: []models.URLOptionKeyValuePair{{Key: "hello", Value: "world"}}}},
			url:      "https://foo.com?me=xxxxxxxx&something=xxxxxxxx",
			command:  "curl -X 'POST' -d 'my request body with ${__qs.me} value' -H 'Accept: application/json;q=0.9,text/plain' -H 'Content-Type: application/json' -H 'Good: xxxxxxxx' -H 'Hello: xxxxxxxx' 'https://foo.com?me=xxxxxxxx&something=xxxxxxxx'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &infinity.Client{Settings: tt.settings}
			got := client.GetExecutedURL(context.TODO(), tt.query)
			assert.Equal(t, fmt.Sprintf("###############\n## URL\n###############\n\n%s\n\n###############\n## Curl Command\n###############\n\n%s", tt.url, tt.command), got)
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name string
		u    string
		want string
	}{
		{u: "https://github.com/yesoreyeram/grafana-infinity-datasource/blob/main/testdata/users.csv", want: "https://raw.githubusercontent.com/yesoreyeram/grafana-infinity-datasource/main/testdata/users.csv"},
		{u: "https://github.com/yesoreyeram/grafana-infinity-datasource/blob/main/testdata/blob/users.csv", want: "https://raw.githubusercontent.com/yesoreyeram/grafana-infinity-datasource/main/testdata/blob/users.csv"},
		{u: "https://raw.githubusercontent.com/yesoreyeram/grafana-infinity-datasource/main/testdata/users.csv", want: "https://raw.githubusercontent.com/yesoreyeram/grafana-infinity-datasource/main/testdata/users.csv"},
		{u: "https://foo.com/channels/38629/feed.json", want: "https://foo.com/channels/38629/feed.json"},
		{u: "https://gitlab.com/restcountries/restcountries/-/blob/master/src/main/resources/countriesV3.json", want: "https://gitlab.com/restcountries/restcountries/-/raw/master/src/main/resources/countriesV3.json"},
		{u: "https://bitbucket.org/yesoreyeram/test_repo/src/master/data/countries.json", want: "https://bitbucket.org/yesoreyeram/test_repo/raw/master/data/countries.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, infinity.NormalizeURL(tt.u))
		})
	}
}
