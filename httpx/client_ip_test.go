package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestClientIPExtractor(t *testing.T) {
	tests := []struct {
		name       string
		trusted    string
		remoteAddr string
		want       string
	}{
		{"trusted proxy forwards client", "10.0.0.0/8", "10.1.2.3:4000", "203.0.113.7"},
		{"untrusted peer is used directly", "10.0.0.0/8", "198.51.100.9:4000", "198.51.100.9"},
		{"private peer outside list is untrusted", "10.0.0.0/8", "192.168.1.5:4000", "192.168.1.5"},
		{"loopback peer outside list is untrusted", "10.0.0.0/8", "127.0.0.1:4000", "127.0.0.1"},
		{"multiple ranges with spaces", " 192.168.0.0/16 , 10.0.0.0/8 ", "192.168.1.5:4000", "203.0.113.7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extract, err := ClientIPExtractor(tt.trusted)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set(echo.HeaderXForwardedFor, "203.0.113.7")
			require.Equal(t, tt.want, extract(req))
		})
	}
}

func TestClientIPExtractorEmpty(t *testing.T) {
	for _, trusted := range []string{"", " , "} {
		extract, err := ClientIPExtractor(trusted)
		require.NoError(t, err)
		require.Nil(t, extract)
	}
}

func TestClientIPExtractorInvalid(t *testing.T) {
	_, err := ClientIPExtractor("10.0.0.1")
	require.Error(t, err, "expected error for a bare IP without a prefix length")
}
