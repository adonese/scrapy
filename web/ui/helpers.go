package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/adonese/cost-of-living/internal/services/estimator"
)

// FormatAED renders AED currency without trailing decimals unless needed.
func FormatAED(v float64) string {
	if v == 0 {
		return "0"
	}
	return fmt.Sprintf("%s", commafy(v))
}

func commafy(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	n := len(s)
	if n <= 3 {
		return s
	}
	rem := n % 3
	if rem == 0 {
		rem = 3
	}
	var out strings.Builder
	out.WriteString(s[:rem])
	for i := rem; i < n; i += 3 {
		out.WriteString(",")
		out.WriteString(s[i : i+3])
	}
	return out.String()
}

func percentShare(part, total float64) string {
	if total <= 0 {
		return "0%"
	}
	return fmt.Sprintf("%d%%", int(math.Round((part/total)*100)))
}

func formatRange(low, high float64) string {
	if low == 0 && high == 0 {
		return "—"
	}
	return fmt.Sprintf("AED %s – %s", FormatAED(low), FormatAED(high))
}

func humanizeTime(ts time.Time) string {
	if ts.IsZero() {
		return "N/A"
	}
	diff := time.Since(ts)
	switch {
	case diff < time.Hour:
		return "just now"
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	default:
		return ts.Format("02 Jan 15:04")
	}
}

func formatConfidence(items []estimator.CategoryEstimate) string {
	if len(items) == 0 {
		return "—"
	}
	var sum float64
	for _, item := range items {
		sum += float64(item.Confidence)
	}
	avg := sum / float64(len(items))
	return fmt.Sprintf("%d%% avg", int(avg*100))
}

func badgeConfidence(conf float32) string {
	return fmt.Sprintf("%d%%", int(conf*100))
}

func TitleCase(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func PersonaHousingType(p estimator.PersonaInput) string {
	return string(p.HousingType)
}

func PersonaLifestyle(p estimator.PersonaInput) string {
	return string(p.Lifestyle)
}

func PersonaTransport(p estimator.PersonaInput) string {
	return string(p.TransportMode)
}
