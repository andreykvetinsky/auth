//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/andreykvetinsky/auth/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestGetTokens(t *testing.T) {
	t.Run("Токены выдаются при валидном guid", func(t *testing.T) {
		body, err := makeRequest("http://localhost:8082/get-tokens?guid=1", "POST", []byte{})
		assert.NoError(t, err)

		authResponse := domain.AuthResponse{}
		err = json.Unmarshal(body, &authResponse)
		assert.NoError(t, err)

		assert.NotEqual(t, "", authResponse.AccessToken)
		assert.NotEqual(t, "", authResponse.RefreshToken)
	})

	t.Run("Получаем ошибку при невалидном guid", func(t *testing.T) {
		_, err := makeRequest("http://localhost:8082/get-tokens?", "POST", []byte{})
		assert.Error(t, err)
	})
}

func TestRefreshTokens(t *testing.T) {
	t.Run("Происходит замена токенов при валидных данных", func(t *testing.T) {
		// получаем токены
		body, err := makeRequest("http://localhost:8082/get-tokens?guid=1", "POST", []byte{})
		assert.NoError(t, err)

		authResponse := domain.AuthResponse{}
		err = json.Unmarshal(body, &authResponse)
		assert.NoError(t, err)

		assert.NotEqual(t, "", authResponse.AccessToken)
		assert.NotEqual(t, "", authResponse.RefreshToken)

		URL, err := buildURL("http://localhost:8082/refresh-tokens", map[string]string{
			"accessToken":  authResponse.AccessToken,
			"refreshToken": authResponse.RefreshToken,
		})
		assert.NoError(t, err)

		body, err = makeRequest(URL, "POST", []byte{})
		assert.NoError(t, err)

		authResponse = domain.AuthResponse{}
		err = json.Unmarshal(body, &authResponse)
		assert.NoError(t, err)

		assert.NotEqual(t, "", authResponse.AccessToken)
		assert.NotEqual(t, "", authResponse.RefreshToken)

	})

	t.Run("Получаем ошибку при подмене токенов", func(t *testing.T) {
		// получаем токены 1
		body, err := makeRequest("http://localhost:8082/get-tokens?guid=1", "POST", []byte{})
		assert.NoError(t, err)

		authResponse := domain.AuthResponse{}
		err = json.Unmarshal(body, &authResponse)
		assert.NoError(t, err)
		assert.NotEqual(t, "", authResponse.AccessToken)
		assert.NotEqual(t, "", authResponse.RefreshToken)

		// получаем токены 2
		body2, err := makeRequest("http://localhost:8082/get-tokens?guid=1", "POST", []byte{})
		assert.NoError(t, err)

		authResponse2 := domain.AuthResponse{}
		err = json.Unmarshal(body2, &authResponse2)
		assert.NoError(t, err)
		assert.NotEqual(t, "", authResponse2.AccessToken)
		assert.NotEqual(t, "", authResponse2.RefreshToken)

		// пытаемся сделать рефреш операцию с токенами разных сессий
		URL, err := buildURL("http://localhost:8082/refresh-tokens", map[string]string{
			"access_token":  authResponse.AccessToken,
			"refresh_token": authResponse2.RefreshToken,
		})
		assert.NoError(t, err)

		_, err = makeRequest(URL, "POST", []byte{})
		assert.Error(t, err)
	})
}

func buildURL(baseURL string, params map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse url error: %w", err)
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func makeRequest(
	url,
	method string,
	body []byte,
) (
	[]byte,
	error,
) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("new request error: %w", err)
	}

	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return nil, fmt.Errorf("status not okay:%d, body:%s", response.StatusCode, string(b))
	}

	return b, nil
}
