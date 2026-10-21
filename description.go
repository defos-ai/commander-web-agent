package web

import (
	"github.com/defos-ai/commander-web-agent/actions"
	actionmodel "github.com/defos-ai/commander/framework/domain/actions"
)

// Responsibilities returns the default capabilities without starting an agent.
func Responsibilities() []string {
	return []string{"Discover website pages and inspect links, SEO metadata and technology indicators"}
}

// ActionDescriptions returns release metadata without schemas or discovery state.
func ActionDescriptions() []actionmodel.Action {
	return []actionmodel.Action{
		{Name: actions.DiscoverPagesName, Description: actions.DiscoverPagesDescription},
		{Name: actions.DiscoverPagesNoBrowserName, Description: actions.DiscoverPagesNoBrowserDescription},
		{Name: actions.CheckLinksName, Description: actions.CheckLinksDescription},
		{Name: actions.CheckSEOName, Description: actions.CheckSEODescription},
		{Name: actions.DetectTechnologiesName, Description: actions.DetectTechnologiesDescription},
	}
}
