package birdsparty

import (
	"strings"

	"github.com/JILI-GAMES/b_backend_games8/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games8/pkg/common/settings"
	"github.com/gofiber/fiber/v2"
)

// RouteGroup holds the dependencies for the handlers
type RouteGroup struct {
	RNGProd      *rng.Client
	SettingsProd *settings.Client
	RNGTest      *rng.Client
	SettingsTest *settings.Client
}

// NewRouteGroup creates a new RouteGroup
func NewRouteGroup(rngProd *rng.Client, settingsProd *settings.Client, rngTest *rng.Client, settingsTest *settings.Client) *RouteGroup {
	return &RouteGroup{
		RNGProd:      rngProd,
		SettingsProd: settingsProd,
		RNGTest:      rngTest,
		SettingsTest: settingsTest,
	}
}

// Helper to select the correct clients per request
func (rg *RouteGroup) getClientsForRequest(c *fiber.Ctx) (*rng.Client, *settings.Client) {
	origin := c.Get("Origin")
	if len(origin) > 0 && (strings.Contains(strings.ToLower(origin), "test")) {
		return rg.RNGTest, rg.SettingsTest
	}
	return rg.RNGProd, rg.SettingsProd
}

// Register registers the routes with the Fiber app
func (rg *RouteGroup) Register(app *fiber.App) {
	app.Post("/spin/birdsparty", rg.SpinHandler)
	app.Post("/process-stage-cleared/birdsparty", rg.ProcessStageClearedHandler)
	app.Post("/cascade/birdsparty", rg.CascadeHandler)
}