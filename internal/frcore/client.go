package frcore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Client exposes the FR Core operations required by LCS.
type Client interface {
	UploadFace(ctx context.Context, req UploadRequest) (*UploadResponse, error)
	Recognize(ctx context.Context, req RecognizeRequest) (*RecognizeResponse, error)
}

// UploadRequest carries the data for registering a face encoding.
type UploadRequest struct {
	Label       string
	ExternalRef string
	ImageName   string
	Image       []byte
}

// UploadResponse mirrors the FR Core response payload.
type UploadResponse struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	ImagePath   string `json:"image_path"`
	ExternalRef string `json:"external_ref"`
}

// RecognizeRequest encapsulates a recognition attempt.
type RecognizeRequest struct {
	ImageName string
	Image     []byte
}

// RecognizeResponse captures the relevant match metadata.
type RecognizeResponse struct {
	Label      string   `json:"label"`
	Similarity float64  `json:"similarity"`
	Distance   *float64 `json:"distance"`
}

// Options configures the FR Core HTTP client.
type Options struct {
	BaseURL         string
	UploadAPIKey    string
	RecognizeAPIKey string
	TenantID        string
	Timeout         time.Duration
	HTTPClient      *http.Client
}

type apiClient struct {
	baseURL         *url.URL
	uploadAPIKey    string
	recognizeAPIKey string
	tenantID        string
	httpClient      *http.Client
}

// NewHTTPClient constructs a HTTP-backed FR Core client.
func NewHTTPClient(opts Options) (Client, error) {
	if opts.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	parsed, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	client := opts.HTTPClient
	if client == nil {
		if opts.Timeout == 0 {
			opts.Timeout = 10 * time.Second
		}
		client = &http.Client{Timeout: opts.Timeout}
	}

	return &apiClient{
		baseURL:         parsed,
		uploadAPIKey:    opts.UploadAPIKey,
		recognizeAPIKey: opts.RecognizeAPIKey,
		tenantID:        opts.TenantID,
		httpClient:      client,
	}, nil
}

func (c *apiClient) UploadFace(ctx context.Context, req UploadRequest) (*UploadResponse, error) {
	if len(req.Image) == 0 {
		return nil, fmt.Errorf("image payload is empty")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("label", req.Label); err != nil {
		return nil, fmt.Errorf("write label field: %w", err)
	}
	if req.ExternalRef != "" {
		if err := writer.WriteField("external_ref", req.ExternalRef); err != nil {
			return nil, fmt.Errorf("write external_ref field: %w", err)
		}
	}

	filename := req.ImageName
	if strings.TrimSpace(filename) == "" {
		filename = "selfie.jpg"
	}

	contentType := determineContentType(req.Image, filename)
	part, err := createFormFileWithContentType(writer, "image", filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(req.Image)); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	endpoint := c.resolvePath("upload")
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	c.applyAuthHeader(httpReq, c.uploadAPIKey)
	logRequest(httpReq, len(req.Image))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(resp.Body)
		logResponse(resp, payload)
		return nil, fmt.Errorf("frcore upload error: status=%d body=%s", resp.StatusCode, string(payload))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	logResponse(resp, bodyBytes)

	var apiResp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			ID          string `json:"id"`
			Label       string `json:"label"`
			ImagePath   string `json:"image_path"`
			ExternalRef string `json:"external_ref"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if strings.ToLower(apiResp.Status) != "success" {
		return nil, fmt.Errorf("frcore upload failed: %s", apiResp.Message)
	}

	return &UploadResponse{
		ID:          apiResp.Data.ID,
		Label:       apiResp.Data.Label,
		ImagePath:   apiResp.Data.ImagePath,
		ExternalRef: apiResp.Data.ExternalRef,
	}, nil
}

func (c *apiClient) Recognize(ctx context.Context, req RecognizeRequest) (*RecognizeResponse, error) {
	if len(req.Image) == 0 {
		return nil, fmt.Errorf("image payload is empty")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filename := req.ImageName
	if strings.TrimSpace(filename) == "" {
		filename = "selfie.jpg"
	}

	contentType := determineContentType(req.Image, filename)
	part, err := createFormFileWithContentType(writer, "image", filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(req.Image)); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	endpoint := c.resolvePath("recognize")
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	c.applyAuthHeader(httpReq, c.recognizeAPIKey)
	logRequest(httpReq, len(req.Image))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(resp.Body)
		logResponse(resp, payload)
		return nil, fmt.Errorf("frcore recognize error: status=%d body=%s", resp.StatusCode, string(payload))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	logResponse(resp, bodyBytes)

	var apiResp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Label      string   `json:"label"`
			Similarity float64  `json:"similarity"`
			Distance   *float64 `json:"distance"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if strings.ToLower(apiResp.Status) != "success" {
		return nil, fmt.Errorf("frcore recognize failed: %s", apiResp.Message)
	}

	return &RecognizeResponse{
		Label:      apiResp.Data.Label,
		Similarity: apiResp.Data.Similarity,
		Distance:   apiResp.Data.Distance,
	}, nil
}

func (c *apiClient) resolvePath(p string) string {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, p)
	return u.String()
}

func (c *apiClient) applyAuthHeader(req *http.Request, apiKey string) {
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	if c.tenantID != "" {
		req.Header.Set("X-Tenant-ID", c.tenantID)
	}
}

var _ Client = (*apiClient)(nil)

func logRequest(req *http.Request, payloadSize int) {
	headers := make(map[string]string)
	for k, v := range req.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	log.Printf("[frcore] request method=%s url=%s headers=%v payload_bytes=%d", req.Method, req.URL.String(), headers, payloadSize)
}

func logResponse(resp *http.Response, body []byte) {
	preview := string(body)
	const maxPreview = 1024
	if len(preview) > maxPreview {
		preview = preview[:maxPreview] + "..."
	}
	log.Printf("[frcore] response status=%d headers=%v body=%s", resp.StatusCode, resp.Header, preview)
}

func determineContentType(data []byte, filename string) string {
	if ext := strings.ToLower(filepath.Ext(filename)); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
	}

	if len(data) > 0 {
		max := len(data)
		if max > 512 {
			max = 512
		}
		return http.DetectContentType(data[:max])
	}

	return "application/octet-stream"
}

func createFormFileWithContentType(w *multipart.Writer, fieldname, filename, contentType string) (io.Writer, error) {
	head := make(textproto.MIMEHeader)
	head.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldname, filename))
	head.Set("Content-Type", contentType)
	return w.CreatePart(head)
}
