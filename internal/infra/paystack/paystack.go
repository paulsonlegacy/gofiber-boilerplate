// Package paystack provides a thin client for the Paystack payment API.
// Initialize with NewPaystackClient and inject via the service layer.
package paystack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.paystack.co"

// Client is the Paystack API client. All methods on this struct require a valid secret key.
type Client struct {
	secretKey  string
	httpClient *http.Client
}

// NewPaystackClient creates a Paystack client using the given secret key.
// Inject this into the service layer at startup.
func NewPaystackClient(secretKey string) *Client {
	return &Client{
		secretKey: secretKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}


// ── Request / Response types ──────────────────────────────────────────────────

// InitializeRequest is the body for POST /transaction/initialize.
type InitializeRequest struct {
	Email     string            `json:"email"`
	Amount    int64             `json:"amount"`       // in kobo (or smallest currency unit)
	Reference string            `json:"reference"`    // idempotency key — your own unique ref
	Currency  string            `json:"currency"`     // NGN, USD, GHS, ZAR, KES
	CallbackURL string          `json:"callback_url,omitempty"`
	Metadata  map[string]any    `json:"metadata,omitempty"`
	Channels  []string          `json:"channels,omitempty"` // card, bank, ussd, qr, mobile_money, bank_transfer
}

// InitializeResponse is the data field of a successful initialize call.
type InitializeResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	AccessCode       string `json:"access_code"`
	Reference        string `json:"reference"`
}

// VerifyResponse is the data field of a successful verify call.
type VerifyResponse struct {
	Status    string `json:"status"`    // success | failed | abandoned
	Reference string `json:"reference"`
	Amount    int64  `json:"amount"`    // in kobo
	Currency  string `json:"currency"`
	GatewayResponse string `json:"gateway_response"`
	PaidAt    string `json:"paid_at"`
	Channel   string `json:"channel"`
	Customer  struct {
		Email string `json:"email"`
	} `json:"customer"`
}

type paystackResponse[T any] struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}


// ── API methods ───────────────────────────────────────────────────────────────

// InitializeTransaction creates a new payment session.
// Returns the authorization URL the user must visit to pay.
func (c *Client) InitializeTransaction(ctx context.Context, req InitializeRequest) (*InitializeResponse, error) {
	var resp paystackResponse[InitializeResponse]
	if err := c.post(ctx, "/transaction/initialize", req, &resp); err != nil {
		return nil, err
	}
	if !resp.Status {
		return nil, fmt.Errorf("paystack initialize: %s", resp.Message)
	}
	return &resp.Data, nil
}

// VerifyTransaction fetches the status of a transaction by reference.
// Call this from your webhook handler or after redirect to confirm payment.
func (c *Client) VerifyTransaction(ctx context.Context, reference string) (*VerifyResponse, error) {
	var resp paystackResponse[VerifyResponse]
	path := fmt.Sprintf("/transaction/verify/%s", reference)
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	if !resp.Status {
		return nil, fmt.Errorf("paystack verify: %s", resp.Message)
	}
	return &resp.Data, nil
}


// ── HTTP helpers ──────────────────────────────────────────────────────────────

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("paystack: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("paystack: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	return c.do(req, out)
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("paystack: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("paystack: http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("paystack: read body: %w", err)
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("paystack: server error %d: %s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, out)
}
