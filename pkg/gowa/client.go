package gowa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

type Config struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
	Timeout    time.Duration
}

type Client struct {
	cfg    Config
	c      *retryablehttp.Client
	base   *url.URL
	common http.Header
}

func New(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:3000"
	}
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	rc := retryablehttp.NewClient()
	rc.RetryWaitMin = 200 * time.Millisecond
	rc.RetryWaitMax = 2 * time.Second
	rc.RetryMax = 3
	if cfg.HTTPClient != nil {
		rc.HTTPClient = cfg.HTTPClient
	}
	rc.HTTPClient.Timeout = cfg.Timeout
	cl := &Client{
		cfg:  cfg,
		c:    rc,
		base: u,
		common: http.Header{
			"Accept":       []string{"application/json"},
			"Content-Type": []string{"application/json"},
		},
	}
	if cfg.Username != "" || cfg.Password != "" {
		basic := base64.StdEncoding.EncodeToString([]byte(cfg.Username + ":" + cfg.Password))
		cl.common.Set("Authorization", "Basic "+basic)
	}
	return cl, nil
}

func (c *Client) url(p string) string {
	return c.base.ResolveReference(&url.URL{Path: path.Join(c.base.Path, p)}).String()
}

func (c *Client) do(ctx context.Context, method, p string, body io.Reader, headers http.Header) (*http.Response, error) {
	req, err := retryablehttp.NewRequest(method, c.url(p), body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	for k, v := range c.common.Clone() {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
	for k, v := range headers {
		for _, vv := range v {
			req.Header.Set(k, vv)
		}
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
	}
	return resp, nil
}

func (c *Client) getJSON(ctx context.Context, p string, q url.Values, out any) error {
	if q != nil && len(q) > 0 {
		p = p + "?" + q.Encode()
	}
	resp, err := c.do(ctx, http.MethodGet, p, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) postJSON(ctx context.Context, p string, in any, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = strings.NewReader(string(b))
	}
	resp, err := c.do(ctx, http.MethodPost, p, body, http.Header{"Content-Type": []string{"application/json"}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) postFormFile(ctx context.Context, p string, fields map[string]string, fileField, filePath string, out any) error {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		defer mw.Close()
		for k, v := range fields {
			_ = mw.WriteField(k, v)
		}
		if filePath != "" {
			f, err := os.Open(filePath)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			defer f.Close()
			fw, err := mw.CreateFormFile(fileField, path.Base(filePath))
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			if _, err := io.Copy(fw, f); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()
	resp, err := c.do(ctx, http.MethodPost, p, pr, http.Header{"Content-Type": []string{mw.FormDataContentType()}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// Tipos de resposta mínimos conforme OpenAPI
type GenericResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Results interface{} `json:"results"`
}

type LoginResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Results struct {
		QRDuration int    `json:"qr_duration"`
		QRLink     string `json:"qr_link"`
	} `json:"results"`
}

type LoginWithCodeResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Results struct {
		PairCode string `json:"pair_code"`
	} `json:"results"`
}

type SendResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Results struct {
		MessageID string `json:"message_id"`
		Status    string `json:"status"`
	} `json:"results"`
}

type UserInfoResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Results struct {
		VerifiedName string `json:"verified_name"`
		Status       string `json:"status"`
		PictureID    string `json:"picture_id"`
	} `json:"results"`
}

type ChatListResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Results struct {
		Data []struct {
			JID             string `json:"jid"`
			Name            string `json:"name"`
			LastMessageTime string `json:"last_message_time"`
			EphemeralExpire int    `json:"ephemeral_expiration"`
		} `json:"data"`
	} `json:"results"`
}

type ChatMessagesResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Results struct {
		Data []struct {
			ID        string  `json:"id"`
			ChatJID   string  `json:"chat_jid"`
			SenderJID string  `json:"sender_jid"`
			Content   string  `json:"content"`
			Timestamp string  `json:"timestamp"`
			IsFromMe  bool    `json:"is_from_me"`
			MediaType *string `json:"media_type"`
		} `json:"data"`
	} `json:"results"`
}

// Métodos de alto nível inteligentes
func (c *Client) Login(ctx context.Context) (*LoginResponse, error) {
	var out LoginResponse
	if err := c.getJSON(ctx, "/app/login", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) LoginWithCode(ctx context.Context, phone string) (*LoginWithCodeResponse, error) {
	if strings.TrimSpace(phone) == "" {
		return nil, errors.New("phone is required")
	}
	q := url.Values{"phone": []string{phone}}
	var out LoginWithCodeResponse
	if err := c.getJSON(ctx, "/app/login-with-code", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Logout(ctx context.Context) error {
	var out GenericResponse
	if err := c.getJSON(ctx, "/app/logout", nil, &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) Reconnect(ctx context.Context) error {
	var out GenericResponse
	if err := c.getJSON(ctx, "/app/reconnect", nil, &out); err != nil {
		return err
	}
	return nil
}

func (c *Client) UserInfo(ctx context.Context, phoneJID string) (*UserInfoResponse, error) {
	if strings.TrimSpace(phoneJID) == "" {
		return nil, errors.New("phoneJID is required")
	}
	q := url.Values{"phone": []string{phoneJID}}
	var out UserInfoResponse
	if err := c.getJSON(ctx, "/user/info", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SendPresence(ctx context.Context, presenceType string, opts ...func(*map[string]any)) (*SendResponse, error) {
	if presenceType != "available" && presenceType != "unavailable" {
		return nil, errors.New("presenceType must be 'available' or 'unavailable'")
	}
	payload := map[string]any{"type": presenceType}
	for _, o := range opts {
		o(&payload)
	}
	var out SendResponse
	if err := c.postJSON(ctx, "/send/presence", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ListChatsParams struct {
	Limit    int
	Offset   int
	Search   string
	HasMedia *bool
}

func (c *Client) ListChats(ctx context.Context, p ListChatsParams) (*ChatListResponse, error) {
	q := url.Values{}
	if p.Limit > 0 {
		q.Set("limit", fmt.Sprint(p.Limit))
	}
	if p.Offset > 0 {
		q.Set("offset", fmt.Sprint(p.Offset))
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if p.HasMedia != nil {
		q.Set("has_media", fmt.Sprint(*p.HasMedia))
	}
	var out ChatListResponse
	if err := c.getJSON(ctx, "/chats", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type GetChatMessagesParams struct {
	Limit     int
	Offset    int
	StartTime string
	EndTime   string
	MediaOnly *bool
	IsFromMe  *bool
	Search    string
}

func (c *Client) GetChatMessages(ctx context.Context, chatJID string, p GetChatMessagesParams) (*ChatMessagesResponse, error) {
	if chatJID == "" {
		return nil, errors.New("chatJID is required")
	}
	q := url.Values{}
	if p.Limit > 0 {
		q.Set("limit", fmt.Sprint(p.Limit))
	}
	if p.Offset > 0 {
		q.Set("offset", fmt.Sprint(p.Offset))
	}
	if p.StartTime != "" {
		q.Set("start_time", p.StartTime)
	}
	if p.EndTime != "" {
		q.Set("end_time", p.EndTime)
	}
	if p.MediaOnly != nil {
		q.Set("media_only", fmt.Sprint(*p.MediaOnly))
	}
	if p.IsFromMe != nil {
		q.Set("is_from_me", fmt.Sprint(*p.IsFromMe))
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	var out ChatMessagesResponse
	path := "/chat/" + url.PathEscape(chatJID) + "/messages"
	if err := c.getJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SendMessage(ctx context.Context, phone, message string, opts ...func(*map[string]any)) (*SendResponse, error) {
	if strings.TrimSpace(phone) == "" || strings.TrimSpace(message) == "" {
		return nil, errors.New("phone and message are required")
	}
	payload := map[string]any{
		"phone":   phone,
		"message": message,
	}
	for _, o := range opts {
		o(&payload)
	}
	var out SendResponse
	if err := c.postJSON(ctx, "/send/message", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func WithReplyMessageID(id string) func(*map[string]any) {
	return func(m *map[string]any) { (*m)["reply_message_id"] = id }
}

func WithForwarded(forwarded bool) func(*map[string]any) {
	return func(m *map[string]any) { (*m)["is_forwarded"] = forwarded }
}

func WithDisappearingDuration(seconds int) func(*map[string]any) {
	return func(m *map[string]any) { (*m)["duration"] = seconds }
}

// Envio de imagem por arquivo local (ou use ImageURL)
func (c *Client) SendImageFile(ctx context.Context, phone, caption, filePath string, viewOnce, compress bool, opts ...func(*map[string]string)) (*SendResponse, error) {
	if phone == "" || filePath == "" {
		return nil, errors.New("phone and filePath are required")
	}
	fields := map[string]string{
		"phone":     phone,
		"caption":   caption,
		"view_once": fmt.Sprint(viewOnce),
		"compress":  fmt.Sprint(compress),
	}
	for _, o := range opts {
		o(&fields)
	}
	var out SendResponse
	if err := c.postFormFile(ctx, "/send/image", fields, "image", filePath, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func WithDurationStr(seconds int) func(*map[string]string) {
	return func(m *map[string]string) { (*m)["duration"] = fmt.Sprint(seconds) }
}

func (c *Client) SendImageURL(ctx context.Context, phone, caption, imageURL string, viewOnce, compress bool, opts ...func(*map[string]any)) (*SendResponse, error) {
	if phone == "" || imageURL == "" {
		return nil, errors.New("phone and imageURL are required")
	}
	payload := map[string]any{
		"phone":     phone,
		"caption":   caption,
		"view_once": viewOnce,
		"compress":  compress,
		"image_url": imageURL,
	}
	for _, o := range opts {
		o(&payload)
	}
	var out SendResponse
	if err := c.postJSON(ctx, "/send/image", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SendAudioParams struct {
	Phone       string
	AudioPath   string // arquivo local
	AudioURL    string // url
	IsForwarded bool
	Duration    int
}

func (c *Client) SendAudio(ctx context.Context, p SendAudioParams) (*SendResponse, error) {
	if p.Phone == "" || (p.AudioPath == "" && p.AudioURL == "") {
		return nil, errors.New("phone and audio required")
	}
	fields := map[string]string{
		"phone":        p.Phone,
		"is_forwarded": fmt.Sprint(p.IsForwarded),
	}
	if p.Duration > 0 {
		fields["duration"] = fmt.Sprint(p.Duration)
	}
	if p.AudioURL != "" {
		fields["audio_url"] = p.AudioURL
	}
	var out SendResponse
	fileField := "audio"
	filePath := p.AudioPath
	if filePath != "" {
		if err := c.postFormFile(ctx, "/send/audio", fields, fileField, filePath, &out); err != nil {
			return nil, err
		}
		return &out, nil
	}
	// se só url, manda como json
	payload := map[string]any{
		"phone":        p.Phone,
		"audio_url":    p.AudioURL,
		"is_forwarded": p.IsForwarded,
	}
	if p.Duration > 0 {
		payload["duration"] = p.Duration
	}
	if err := c.postJSON(ctx, "/send/audio", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SendFileParams struct {
	Phone       string
	Caption     string
	FilePath    string
	IsForwarded bool
	Duration    int
}

func (c *Client) SendFile(ctx context.Context, p SendFileParams) (*SendResponse, error) {
	if p.Phone == "" || p.FilePath == "" {
		return nil, errors.New("phone and filePath required")
	}
	fields := map[string]string{
		"phone":        p.Phone,
		"caption":      p.Caption,
		"is_forwarded": fmt.Sprint(p.IsForwarded),
	}
	if p.Duration > 0 {
		fields["duration"] = fmt.Sprint(p.Duration)
	}
	var out SendResponse
	if err := c.postFormFile(ctx, "/send/file", fields, "file", p.FilePath, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SendVideoParams struct {
	Phone       string
	Caption     string
	VideoPath   string
	VideoURL    string
	ViewOnce    bool
	Compress    bool
	IsForwarded bool
	Duration    int
}

func (c *Client) SendVideo(ctx context.Context, p SendVideoParams) (*SendResponse, error) {
	if p.Phone == "" || (p.VideoPath == "" && p.VideoURL == "") {
		return nil, errors.New("phone and video required")
	}
	fields := map[string]string{
		"phone":        p.Phone,
		"caption":      p.Caption,
		"view_once":    fmt.Sprint(p.ViewOnce),
		"compress":     fmt.Sprint(p.Compress),
		"is_forwarded": fmt.Sprint(p.IsForwarded),
	}
	if p.Duration > 0 {
		fields["duration"] = fmt.Sprint(p.Duration)
	}
	if p.VideoURL != "" {
		fields["video_url"] = p.VideoURL
	}
	var out SendResponse
	fileField := "video"
	filePath := p.VideoPath
	if filePath != "" {
		if err := c.postFormFile(ctx, "/send/video", fields, fileField, filePath, &out); err != nil {
			return nil, err
		}
		return &out, nil
	}
	// se só url, manda como json
	payload := map[string]any{
		"phone":        p.Phone,
		"caption":      p.Caption,
		"view_once":    p.ViewOnce,
		"compress":     p.Compress,
		"video_url":    p.VideoURL,
		"is_forwarded": p.IsForwarded,
	}
	if p.Duration > 0 {
		payload["duration"] = p.Duration
	}
	if err := c.postJSON(ctx, "/send/video", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SendContactParams struct {
	Phone        string
	ContactName  string
	ContactPhone string
	IsForwarded  bool
	Duration     int
}

func (c *Client) SendContact(ctx context.Context, p SendContactParams) (*SendResponse, error) {
	if p.Phone == "" || p.ContactName == "" || p.ContactPhone == "" {
		return nil, errors.New("phone, contactName, contactPhone required")
	}
	payload := map[string]any{
		"phone":         p.Phone,
		"contact_name":  p.ContactName,
		"contact_phone": p.ContactPhone,
		"is_forwarded":  p.IsForwarded,
	}
	if p.Duration > 0 {
		payload["duration"] = p.Duration
	}
	var out SendResponse
	if err := c.postJSON(ctx, "/send/contact", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SendLinkParams struct {
	Phone       string
	Link        string
	Caption     string
	IsForwarded bool
	Duration    int
}

func (c *Client) SendLink(ctx context.Context, p SendLinkParams) (*SendResponse, error) {
	if p.Phone == "" || p.Link == "" {
		return nil, errors.New("phone and link required")
	}
	payload := map[string]any{
		"phone":        p.Phone,
		"link":         p.Link,
		"caption":      p.Caption,
		"is_forwarded": p.IsForwarded,
	}
	if p.Duration > 0 {
		payload["duration"] = p.Duration
	}
	var out SendResponse
	if err := c.postJSON(ctx, "/send/link", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SendLocationParams struct {
	Phone       string
	Latitude    string
	Longitude   string
	IsForwarded bool
	Duration    int
}

func (c *Client) SendLocation(ctx context.Context, p SendLocationParams) (*SendResponse, error) {
	if p.Phone == "" || p.Latitude == "" || p.Longitude == "" {
		return nil, errors.New("phone, latitude, longitude required")
	}
	payload := map[string]any{
		"phone":        p.Phone,
		"latitude":     p.Latitude,
		"longitude":    p.Longitude,
		"is_forwarded": p.IsForwarded,
	}
	if p.Duration > 0 {
		payload["duration"] = p.Duration
	}
	var out SendResponse
	if err := c.postJSON(ctx, "/send/location", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SendPollParams struct {
	Phone     string
	Question  string
	Options   []string
	MaxAnswer int
	Duration  int
}

func (c *Client) SendPoll(ctx context.Context, p SendPollParams) (*SendResponse, error) {
	if p.Phone == "" || p.Question == "" || len(p.Options) == 0 || p.MaxAnswer == 0 {
		return nil, errors.New("phone, question, options, maxAnswer required")
	}
	payload := map[string]any{
		"phone":      p.Phone,
		"question":   p.Question,
		"options":    p.Options,
		"max_answer": p.MaxAnswer,
	}
	if p.Duration > 0 {
		payload["duration"] = p.Duration
	}
	var out SendResponse
	if err := c.postJSON(ctx, "/send/poll", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SendChatPresenceParams struct {
	Phone  string
	Action string // start ou stop
}

func (c *Client) SendChatPresence(ctx context.Context, p SendChatPresenceParams) (*SendResponse, error) {
	if p.Phone == "" || (p.Action != "start" && p.Action != "stop") {
		return nil, errors.New("phone and action=start|stop required")
	}
	payload := map[string]any{
		"phone":  p.Phone,
		"action": p.Action,
	}
	var out SendResponse
	if err := c.postJSON(ctx, "/send/chat-presence", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type MessageActionParams struct {
	MessageID string
	Phone     string
	Emoji     string // para react
	Message   string // para update
}

func (c *Client) RevokeMessage(ctx context.Context, p MessageActionParams) (*SendResponse, error) {
	if p.MessageID == "" || p.Phone == "" {
		return nil, errors.New("messageID and phone required")
	}
	payload := map[string]any{"phone": p.Phone}
	var out SendResponse
	path := "/message/" + url.PathEscape(p.MessageID) + "/revoke"
	if err := c.postJSON(ctx, path, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteMessage(ctx context.Context, p MessageActionParams) (*SendResponse, error) {
	if p.MessageID == "" || p.Phone == "" {
		return nil, errors.New("messageID and phone required")
	}
	payload := map[string]any{"phone": p.Phone}
	var out SendResponse
	path := "/message/" + url.PathEscape(p.MessageID) + "/delete"
	if err := c.postJSON(ctx, path, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ReactMessage(ctx context.Context, p MessageActionParams) (*SendResponse, error) {
	if p.MessageID == "" || p.Phone == "" || p.Emoji == "" {
		return nil, errors.New("messageID, phone, emoji required")
	}
	payload := map[string]any{"phone": p.Phone, "emoji": p.Emoji}
	var out SendResponse
	path := "/message/" + url.PathEscape(p.MessageID) + "/reaction"
	if err := c.postJSON(ctx, path, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateMessage(ctx context.Context, p MessageActionParams) (*SendResponse, error) {
	if p.MessageID == "" || p.Phone == "" || p.Message == "" {
		return nil, errors.New("messageID, phone, message required")
	}
	payload := map[string]any{"phone": p.Phone, "message": p.Message}
	var out SendResponse
	path := "/message/" + url.PathEscape(p.MessageID) + "/update"
	if err := c.postJSON(ctx, path, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ReadMessage(ctx context.Context, p MessageActionParams) (*SendResponse, error) {
	if p.MessageID == "" || p.Phone == "" {
		return nil, errors.New("messageID and phone required")
	}
	payload := map[string]any{"phone": p.Phone}
	var out SendResponse
	path := "/message/" + url.PathEscape(p.MessageID) + "/read"
	if err := c.postJSON(ctx, path, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) StarMessage(ctx context.Context, p MessageActionParams) (*GenericResponse, error) {
	if p.MessageID == "" || p.Phone == "" {
		return nil, errors.New("messageID and phone required")
	}
	payload := map[string]any{"phone": p.Phone}
	var out GenericResponse
	path := "/message/" + url.PathEscape(p.MessageID) + "/star"
	if err := c.postJSON(ctx, path, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UnstarMessage(ctx context.Context, p MessageActionParams) (*GenericResponse, error) {
	if p.MessageID == "" || p.Phone == "" {
		return nil, errors.New("messageID and phone required")
	}
	payload := map[string]any{"phone": p.Phone}
	var out GenericResponse
	path := "/message/" + url.PathEscape(p.MessageID) + "/unstar"
	if err := c.postJSON(ctx, path, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
