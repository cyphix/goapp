package httpx

import (
	"fmt"
	"net"
	"strings"

	"github.com/labstack/echo/v4"
)

// ClientIPExtractor builds the goauth ClientIP extractor from a comma-separated
// list of trusted reverse-proxy CIDRs. X-Forwarded-For is honored only when the
// request comes from one of those ranges. An empty list returns nil, keeping
// goauth's default of the direct peer address.
func ClientIPExtractor(trustedProxies string) (echo.IPExtractor, error) {
	var opts []echo.TrustOption
	for _, raw := range strings.Split(trustedProxies, ",") {
		cidr := strings.TrimSpace(raw)
		if cidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("parse trusted proxy %q: %w", cidr, err)
		}
		opts = append(opts, echo.TrustIPRange(ipNet))
	}
	if len(opts) == 0 {
		return nil, nil
	}
	opts = append(opts,
		echo.TrustLoopback(false),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	)
	return echo.ExtractIPFromXFFHeader(opts...), nil
}
