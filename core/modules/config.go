package modules

import (
	"encoding/json"
	"fmt"
)

// Filename is the name of the module config file.
const Filename = "dagger.json"

// EngineVersionLatest is replaced by the current engine.Version during module init.
const EngineVersionLatest string = "latest"

// ModuleConfig is the config for a single module as loaded from a dagger.json file.
// Only contains fields that are set/edited by dagger utilities.
type ModuleConfig struct {
	// The name of the module.
	Name string `json:"name"`

	// The version of the engine this module was last updated with.
	EngineVersion string `json:"engineVersion"`

	// The SDK this module uses
	SDK string `json:"sdk,omitempty"`

	// Paths to explicitly include from the module, relative to the configuration file.
	Include []string `json:"include,omitempty"`

	// Paths to explicitly exclude from the module, relative to the configuration file.
	Exclude []string `json:"exclude,omitempty"`

	// The modules this module depends on.
	Dependencies []*ModuleConfigDependency `json:"dependencies,omitempty"`

	// The path, relative to this config file, to the subdir containing the module's implementation source code.
	Source string `json:"source,omitempty"`

	// Named views defined for this module, which are sets of directory filters that can be applied to
	// directory arguments provided to functions.
	Views []*ModuleConfigView `json:"views,omitempty"`

	// Codegen configuration for this module.
	Codegen *ModuleCodegenConfig `json:"codegen,omitempty"`
}

type ModuleConfigUserFields struct {
	// The self-describing json $schema
	Schema string `json:"$schema,omitempty"`
}

// ModuleConfigWithUserFields is the config for a single module as loaded from a dagger.json file.
// Includes additional fields that should only be set by the user.
type ModuleConfigWithUserFields struct {
	ModuleConfigUserFields
	ModuleConfig
}

func (modCfg *ModuleConfig) UnmarshalJSON(data []byte) error {
	if modCfg == nil {
		return fmt.Errorf("cannot unmarshal into nil %T", modCfg)
	}
	if len(data) == 0 {
		return nil
	}

	type alias ModuleConfig // lets us use the default json unmashaler
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("unmarshal module config: %w", err)
	}

	// Detect the case where SDK is set but Source isn't, which should only happen when loading an older config.
	// For those cases, the Source was implicitly ".", so set it to that.
	if tmp.SDK != "" && tmp.Source == "" {
		tmp.Source = "."
	}

	*modCfg = ModuleConfig(tmp)
	return nil
}

func (modCfg *ModuleConfigWithUserFields) UnmarshalJSON(data []byte) error {
	if modCfg == nil {
		return fmt.Errorf("cannot unmarshal into nil %T", modCfg)
	}
	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &modCfg.ModuleConfigUserFields); err != nil {
		return fmt.Errorf("unmarshal module config: %w", err)
	}
	if err := json.Unmarshal(data, &modCfg.ModuleConfig); err != nil {
		return fmt.Errorf("unmarshal module config: %w", err)
	}
	return nil
}

func (modCfg *ModuleConfig) DependencyByName(name string) (*ModuleConfigDependency, bool) {
	for _, dep := range modCfg.Dependencies {
		if dep.Name == name {
			return dep, true
		}
	}
	return nil, false
}

type ModuleConfigDependency struct {
	// The name to use for this dependency. By default, the same as the dependency module's name,
	// but can also be overridden to use a different name.
	Name string `json:"name"`

	// The source ref of the module dependency.
	Source string `json:"source"`

	// The pinned version of the module dependency.
	Pin string `json:"pin,omitempty"`
}

func (depCfg *ModuleConfigDependency) UnmarshalJSON(data []byte) error {
	if depCfg == nil {
		return fmt.Errorf("cannot unmarshal into nil ModuleConfigDependency")
	}
	if len(data) == 0 {
		depCfg.Source = ""
		return nil
	}

	// check if this is a legacy config, where deps were just a list of strings
	if data[0] == '"' {
		var depRefStr string
		if err := json.Unmarshal(data, &depRefStr); err != nil {
			return fmt.Errorf("unmarshal module config dependency: %w", err)
		}
		*depCfg = ModuleConfigDependency{Source: depRefStr}
		return nil
	}

	type alias ModuleConfigDependency // lets us use the default json unmashaler
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("unmarshal module config dependency: %w", err)
	}
	*depCfg = ModuleConfigDependency(tmp)
	return nil
}

type ModuleConfigView struct {
	Name     string   `json:"name"`
	Patterns []string `json:"patterns,omitempty"`
}

type ModuleCodegenConfig struct {
	// Whether to automatically generate a .gitignore file for this module.
	AutomaticGitignore *bool `json:"automaticGitignore,omitempty"`
}
