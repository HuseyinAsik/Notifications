package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"go.uber.org/zap"
)

type httpClient struct {
	Client *http.Client
	Logger *logging.LogWrapper
}

type HTTPClient interface {
	DoJSON(ctx context.Context, method, url string, req any, headers map[string]string, response any) (int, error)
}

func NewHTTPClient(client *http.Client, logger *logging.LogWrapper) HTTPClient {
	return &httpClient{
		Client: client,
		Logger: logger,
	}
}

func (h *httpClient) DoJSON(ctx context.Context, method, url string, reqBody any, headers map[string]string, response any) (int, error) {
	var body io.Reader
	if reqBody != nil {
		byteReqBody, marshalErr := json.Marshal(reqBody)
		if marshalErr != nil {
			h.Logger.Error(ctx, "DoJson marshal err:", zap.Error(marshalErr))
			return 0, marshalErr
		}
		body = bytes.NewBuffer(byteReqBody)
	}

	req, buildErr := http.NewRequestWithContext(ctx, method, url, body)
	if buildErr != nil {
		h.Logger.Error(ctx, "DoJson request build err", zap.Error(buildErr))
		return 0, buildErr
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, doErr := h.Client.Do(req)
	if doErr != nil {
		h.Logger.Error(ctx, "DoJson http do error:", zap.Error(doErr))
		return 0, doErr
	}

	// Body always closed here
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	status := resp.StatusCode

	if response == nil {
		return status, nil
	}

	respBytes, readBodyErr := io.ReadAll(resp.Body)
	if readBodyErr != nil {
		h.Logger.Error(ctx, "DoJson read body error:", zap.Error(readBodyErr))
		return status, readBodyErr
	}

	if len(respBytes) == 0 {
		return status, nil
	}

	if unmarshalErr := json.Unmarshal(respBytes, response); unmarshalErr != nil {
		h.Logger.Error(ctx, "DoJson unmarshal error:", zap.Error(unmarshalErr))
		return status, unmarshalErr
	}

	return status, nil

}
