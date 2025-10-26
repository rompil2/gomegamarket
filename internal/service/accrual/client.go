package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rompil2/gomegamarket/internal/models"
)

var (
	ErrRateLimitExceeded         = errors.New("accrual service rate limit exceeded")
	ErrAccrualServiceInternal    = errors.New("accrual service internal error")
	ErrAccrualServiceUnavailable = errors.New("accrual service unavailable")
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*models.OrderAccrual, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrual models.OrderAccrual
		if err := decodeJSON(resp.Body, &accrual); err != nil {
			return nil, err
		}
		return &accrual, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		return nil, ErrRateLimitExceeded
	case http.StatusInternalServerError:
		return nil, ErrAccrualServiceInternal
	case http.StatusServiceUnavailable:
		return nil, ErrAccrualServiceUnavailable
	default:
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func decodeJSON(readCloser io.ReadCloser, orderAccrual *models.OrderAccrual) error {
	jsonDecoder := json.NewDecoder(readCloser)
	if err := jsonDecoder.Decode(orderAccrual); err != nil {
		if err == io.EOF {
			return nil
		} else {
			return err
		}
	}
	return nil
}
