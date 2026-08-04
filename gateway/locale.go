package gateway

import "github.com/jeandeaual/go-locale"

func (g *gatewayImpl) mustGetLocale() string {
	getLocale, err := locale.GetLocale()
	if err != nil || getLocale == "" {
		// Fallback for headless/server environments (e.g. Replit) where no
		// locale is configured. Discord accepts "en-US" as a safe default.
		return "en-US"
	}
	return getLocale
}
