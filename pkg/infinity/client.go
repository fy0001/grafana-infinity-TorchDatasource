package infinity

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/yesoreyeram/grafana-infinity-datasource/pkg/mercury"
	"github.com/yesoreyeram/grafana-infinity-datasource/pkg/models"
)

type Client struct {
	Settings        models.InfinitySettings
	HttpClient      *http.Client
	AzureBlobClient *azblob.Client
	IsMock          bool
}

func GetTLSConfigFromSettings(settings models.InfinitySettings) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: settings.InsecureSkipVerify,
		ServerName:         settings.ServerName,
	}
	if settings.TLSClientAuth {
		if settings.TLSClientCert == "" || settings.TLSClientKey == "" {
			return nil, errors.New("invalid Client cert or key")
		}
		cert, err := tls.X509KeyPair([]byte(settings.TLSClientCert), []byte(settings.TLSClientKey))
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if settings.TLSAuthWithCACert && settings.TLSCACert != "" {
		caPool := x509.NewCertPool()
		ok := caPool.AppendCertsFromPEM([]byte(settings.TLSCACert))
		if !ok {
			return nil, errors.New("invalid TLS CA certificate")
		}
		tlsConfig.RootCAs = caPool
	}
	return tlsConfig, nil
}

func getBaseHTTPClient(ctx context.Context, settings models.InfinitySettings) *http.Client {
	tlsConfig, err := GetTLSConfigFromSettings(settings)
	if err != nil {
		return nil
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	switch settings.ProxyType {
	case models.ProxyTypeNone:
		backend.Logger.Debug("proxy type is set to none. Not using the proxy")
	case models.ProxyTypeUrl:
		backend.Logger.Debug("proxy type is set to url. Using the proxy", "proxy_url", settings.ProxyUrl)
		u, err := url.Parse(settings.ProxyUrl)
		if err != nil {
			backend.Logger.Error("error parsing proxy url", "err", err.Error(), "proxy_url", settings.ProxyUrl)
			return nil
		}
		transport.Proxy = http.ProxyURL(u)
	default:
		transport.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{
		Transport: transport,
		Timeout:   time.Second * time.Duration(settings.TimeoutInSeconds),
	}
}

func NewClient(ctx context.Context, settings models.InfinitySettings) (client *Client, err error) {
	_, span := tracing.DefaultTracer().Start(ctx, "NewClient")
	defer span.End()
	if settings.AuthenticationMethod == "" {
		settings.AuthenticationMethod = models.AuthenticationMethodNone
		if settings.BasicAuthEnabled {
			settings.AuthenticationMethod = models.AuthenticationMethodBasic
		}
		if settings.ForwardOauthIdentity {
			settings.AuthenticationMethod = models.AuthenticationMethodForwardOauth
		}
	}
	httpClient := getBaseHTTPClient(ctx, settings)
	if httpClient == nil {
		span.RecordError(errors.New("invalid http client"))
		return nil, errors.New("invalid http client")
	}
	httpClient = ApplyDigestAuth(ctx, httpClient, settings)
	httpClient = ApplyOAuthClientCredentials(ctx, httpClient, settings)
	httpClient = ApplyOAuthJWT(ctx, httpClient, settings)
	httpClient = ApplyAWSAuth(ctx, httpClient, settings)
	client = &Client{
		Settings:   settings,
		HttpClient: httpClient,
	}
	if settings.AuthenticationMethod == models.AuthenticationMethodAzureBlob {
		cred, err := azblob.NewSharedKeyCredential(settings.AzureBlobAccountName, settings.AzureBlobAccountKey)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(500, err.Error())
			return nil, fmt.Errorf("invalid azure blob credentials. %s", err)
		}
		clientUrl := "https://%s.blob.core.windows.net/"
		if settings.AzureBlobAccountUrl != "" {
			clientUrl = settings.AzureBlobAccountUrl
		}
		if strings.Contains(clientUrl, "%s") {
			clientUrl = fmt.Sprintf(clientUrl, settings.AzureBlobAccountName)
		}
		azClient, err := azblob.NewClientWithSharedKeyCredential(clientUrl, cred, nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(500, err.Error())
			return nil, fmt.Errorf("invalid azure blob client. %s", err)
		}
		if azClient == nil {
			span.RecordError(errors.New("invalid/empty azure blob client"))
			span.SetStatus(500, "invalid/empty azure blob client")
			return nil, errors.New("invalid/empty azure blob client")
		}
		client.AzureBlobClient = azClient
	}
	if settings.IsMock {
		client.IsMock = true
	}
	return client, err
}

func replaceSect(input string, settings models.InfinitySettings, includeSect bool) string {
	for key, value := range settings.SecureQueryFields {
		if includeSect {
			input = strings.ReplaceAll(input, fmt.Sprintf("${__qs.%s}", key), value)
		}
		if !includeSect {
			input = strings.ReplaceAll(input, fmt.Sprintf("${__qs.%s}", key), dummyHeader)
		}
	}
	return input
}

var includeSect bool = true

func ApplyZCapAuth(settings models.InfinitySettings) (string, string, error) {
	var contentT []byte
	var dataT []byte
	var err error

	zcapInputTarget := dummyHeader

	if includeSect {
		zcapInputTarget = settings.ZCapJsonPath

	}
	var operation string = "download"
	var target string = zcapInputTarget

	content, data, err := mercury.Request(operation, target)
	//download - data
	//request - content

	if err != nil {
		backend.Logger.Error("Error:", err)
	}
	contentT = content
	dataT = data

	return string(contentT), string(dataT), err

}

func (client *Client) req(ctx context.Context, url string, body io.Reader, settings models.InfinitySettings, query models.Query, requestHeaders map[string]string) (obj any, statusCode int, duration time.Duration, err error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "client.req")
	defer span.End()

	req, _ := GetRequest(ctx, settings, body, query, requestHeaders, true)
	startTime := time.Now()
	if !CanAllowURL(req.URL.String(), settings.AllowedHosts) {
		backend.Logger.Error("url is not in the allowed list. make sure to match the base URL with the settings", "url", req.URL.String())
		return nil, http.StatusUnauthorized, 0, errors.New("requested URL is not allowed. To allow this URL, update the datasource config Security -> Allowed Hosts section")
	}
	backend.Logger.Debug("yesoreyeram-infinity-datasource plugin is now requesting URL", "url", req.URL.String())
	res, err := client.HttpClient.Do(req)
	duration = time.Since(startTime)

	// Use MercuryClient for zCap Authenticated Requests
	if settings.AuthenticationMethod == models.AuthenticationMethodZCAP {
		console, data, err := ApplyZCapAuth(settings)

		backend.Logger.Info("entered in ZCAP", console) //displays in powershell log when running
		res.StatusCode = 200
		//duration = time.Since(startTime)

		return data, res.StatusCode, duration, err
	}

	if res != nil {
		defer res.Body.Close()
	}
	if err != nil && res != nil {
		backend.Logger.Error("error getting response from server", "url", url, "method", req.Method, "error", err.Error(), "status code", res.StatusCode)
		return nil, res.StatusCode, duration, fmt.Errorf("error getting response from %s", url)
	}
	if err != nil && res == nil {
		backend.Logger.Error("error getting response from server. no response received", "url", url, "error", err.Error())
		return nil, http.StatusInternalServerError, duration, fmt.Errorf("error getting response from url %s. no response received. Error: %s", url, err.Error())
	}
	if err == nil && res == nil {
		backend.Logger.Error("invalid response from server and also no error", "url", url, "method", req.Method)
		return nil, http.StatusInternalServerError, duration, fmt.Errorf("invalid response received for the URL %s", url)
	}
	if res.StatusCode >= http.StatusBadRequest {
		return nil, res.StatusCode, duration, errors.New(res.Status)
	}
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		backend.Logger.Error("error reading response body", "url", url, "error", err.Error())
		return nil, res.StatusCode, duration, err
	}
	bodyBytes = removeBOMContent(bodyBytes)
	if CanParseAsJSON(query.Type, res.Header) {
		var out any
		err := json.Unmarshal(bodyBytes, &out)
		if err != nil {
			backend.Logger.Error("error un-marshaling JSON response", "url", url, "error", err.Error())
		}
		return out, res.StatusCode, duration, err
	}
	return string(bodyBytes), res.StatusCode, duration, err
}

// https://stackoverflow.com/questions/31398044/got-error-invalid-character-%C3%AF-looking-for-beginning-of-value-from-json-unmar
func removeBOMContent(input []byte) []byte {
	return bytes.TrimPrefix(input, []byte("\xef\xbb\xbf"))
}

func (client *Client) GetResults(ctx context.Context, query models.Query, requestHeaders map[string]string) (o any, statusCode int, duration time.Duration, err error) {
	if query.Source == "azure-blob" {
		if strings.TrimSpace(query.AzBlobContainerName) == "" || strings.TrimSpace(query.AzBlobName) == "" {
			return nil, http.StatusBadRequest, 0, errors.New("invalid/empty container name/blob name")
		}
		if client.AzureBlobClient == nil {
			return nil, http.StatusInternalServerError, 0, errors.New("invalid azure blob client")
		}
		blobDownloadResponse, err := client.AzureBlobClient.DownloadStream(ctx, strings.TrimSpace(query.AzBlobContainerName), strings.TrimSpace(query.AzBlobName), nil)
		if err != nil {
			return nil, http.StatusInternalServerError, 0, err
		}
		reader := blobDownloadResponse.Body
		bodyBytes, err := io.ReadAll(reader)
		if err != nil {
			return nil, http.StatusInternalServerError, 0, fmt.Errorf("error reading blob content. %w", err)
		}
		bodyBytes = removeBOMContent(bodyBytes)
		if CanParseAsJSON(query.Type, http.Header{}) {
			var out any
			err := json.Unmarshal(bodyBytes, &out)
			if err != nil {
				backend.Logger.Error("error un-marshaling blob content", "error", err.Error())
			}
			return out, http.StatusOK, duration, err
		}
		return string(bodyBytes), http.StatusOK, 0, nil
	}
	switch strings.ToUpper(query.URLOptions.Method) {
	case http.MethodPost:
		body := GetQueryBody(query)
		return client.req(ctx, query.URL, body, client.Settings, query, requestHeaders)
	default:
		return client.req(ctx, query.URL, nil, client.Settings, query, requestHeaders)
	}
}

func CanParseAsJSON(queryType models.QueryType, responseHeaders http.Header) bool {
	if queryType == models.QueryTypeJSON || queryType == models.QueryTypeGraphQL {
		return true
	}
	if queryType == models.QueryTypeUQL || queryType == models.QueryTypeGROQ {
		contentType := responseHeaders.Get(headerKeyContentType)
		if strings.Contains(strings.ToLower(contentType), contentTypeJSON) {
			return true
		}
	}
	return false
}

func CanAllowURL(url string, allowedHosts []string) bool {
	allow := false
	if len(allowedHosts) == 0 {
		return true
	}
	for _, host := range allowedHosts {
		if strings.HasPrefix(url, host) {
			return true
		}
	}
	return allow
}

func GetQueryBody(query models.Query) io.Reader {
	var body io.Reader
	if strings.EqualFold(query.URLOptions.Method, http.MethodPost) {
		switch query.URLOptions.BodyType {
		case "raw":
			body = strings.NewReader(query.URLOptions.Body)
		case "form-data":
			payload := &bytes.Buffer{}
			writer := multipart.NewWriter(payload)
			for _, f := range query.URLOptions.BodyForm {
				_ = writer.WriteField(f.Key, f.Value)
			}
			if err := writer.Close(); err != nil {
				backend.Logger.Error("error closing the query body reader")
				return nil
			}
			body = payload
		case "x-www-form-urlencoded":
			form := url.Values{}
			for _, f := range query.URLOptions.BodyForm {
				form.Set(f.Key, f.Value)
			}
			body = strings.NewReader(form.Encode())
		case "graphql":
			var variables map[string]interface{}
			if query.URLOptions.BodyGraphQLVariables != "" {
				err := json.Unmarshal([]byte(query.URLOptions.BodyGraphQLVariables), &variables)
				if err != nil {
					backend.Logger.Error("Error parsing graphql variable json", err)
				}
			}
			jsonData := map[string]interface{}{
				"query":     query.URLOptions.BodyGraphQLQuery,
				"variables": variables,
			}
			jsonValue, _ := json.Marshal(jsonData)
			body = strings.NewReader(string(jsonValue))
		default:
			body = strings.NewReader(query.URLOptions.Body)
		}
	}
	return body
}
