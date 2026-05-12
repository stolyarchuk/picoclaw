package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
)

func listCommand() Definition {
	return Definition{
		Name:        "list",
		Description: "List available options",
		SubCommands: []SubCommand{
			{
				Name:        "models",
				Description: "Configured models",
				Handler: func(_ context.Context, req Request, rt *Runtime) error {
					if rt == nil {
						return req.Reply(unavailableMsg)
					}
					return req.Reply(formatConfiguredModels(rt))
				},
			},
			{
				Name:        "channels",
				Description: "Enabled channels",
				Handler: func(_ context.Context, req Request, rt *Runtime) error {
					if rt == nil || rt.GetEnabledChannels == nil {
						return req.Reply(unavailableMsg)
					}
					enabled := rt.GetEnabledChannels()
					if len(enabled) == 0 {
						return req.Reply("No channels enabled")
					}
					return req.Reply(fmt.Sprintf("Enabled Channels:\n- %s", strings.Join(enabled, "\n- ")))
				},
			},
			{
				Name:        "agents",
				Description: "Registered agents",
				Handler:     agentsHandler(),
			},
			{
				Name:        "skills",
				Description: "Installed skills",
				Handler: func(_ context.Context, req Request, rt *Runtime) error {
					if rt == nil || rt.ListSkillNames == nil {
						return req.Reply(unavailableMsg)
					}
					names := rt.ListSkillNames()
					if len(names) == 0 {
						return req.Reply("No installed skills")
					}
					return req.Reply(fmt.Sprintf(
						"Installed Skills:\n- %s\n\nUse /use <skill> <message> to force one for a single request, or /use <skill> to apply it to your next message.",
						strings.Join(names, "\n- "),
					))
				},
			},
			{
				Name:        "mcp",
				Description: "Configured MCP servers",
				Handler:     listMCPServersHandler(),
			},
		},
	}
}

func formatConfiguredModels(rt *Runtime) string {
	if rt == nil || rt.Config == nil || len(rt.Config.ModelList) == 0 {
		return "No models configured in model_list"
	}

	currentModel := rt.Config.Agents.Defaults.GetModelName()
	if rt.GetModelInfo != nil {
		if name, _ := rt.GetModelInfo(); strings.TrimSpace(name) != "" {
			currentModel = name
		}
	}

	lines := []string{"Available Models:"}
	enabledCount := 0
	for _, model := range rt.Config.ModelList {
		if !isListableModel(model) {
			continue
		}
		enabledCount++

		marker := "-"
		if model.ModelName == currentModel {
			marker = ">"
		}

		provider := strings.TrimSpace(model.Provider)
		if provider == "" {
			provider = modelProviderName(model.Model)
		}
		if provider == "" {
			provider = "configured default"
		}

		lines = append(lines, fmt.Sprintf(
			"%s %s (%s, Provider: %s)",
			marker,
			model.ModelName,
			model.Model,
			provider,
		))
	}
	if enabledCount == 0 {
		return "No enabled models configured in model_list"
	}

	return strings.Join(lines, "\n")
}

func isListableModel(model *config.ModelConfig) bool {
	if model == nil {
		return false
	}
	if model.Enabled {
		return true
	}
	if len(model.APIKeys.Values()) > 0 {
		return true
	}
	protocol, _ := providers.ExtractProtocol(model)
	return providers.IsEmptyAPIKeyAllowedForProtocol(protocol)
}

func modelProviderName(model string) string {
	provider, _, ok := strings.Cut(strings.TrimSpace(model), "/")
	if !ok {
		return ""
	}
	return strings.TrimSpace(provider)
}
