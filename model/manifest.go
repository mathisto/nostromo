package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pokanop/nostromo/keypath"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/version"
	"github.com/shivamMg/ppds/tree"
	"gopkg.in/yaml.v2"
)

type ManifestV0 struct {
	Version  string              `json:"version"`
	Config   *Config             `json:"config"`
	Commands map[string]*Command `json:"commands"`
}

// Manifest is the main container for nostromo based commands
type Manifest struct {
	Version  *version.Info       `json:"version"`
	Config   *Config             `json:"config"`
	Commands map[string]*Command `json:"commands"`
}

// NewManifest returns a newly initialized manifest
func NewManifest(version *version.Info) *Manifest {
	return &Manifest{
		Version:  version,
		Config:   NewConfig(),
		Commands: map[string]*Command{},
	}
}

// Link a newly loaded manifest
//
// This must be run after parsing a manifest to walk the command
// tree and build links.
func (m *Manifest) Link() {
	for _, cmd := range m.Commands {
		cmd.link(nil)
	}
}

// AddCommand tree up to key path
func (m *Manifest) AddCommand(keyPath, command, description string, code *Code, aliasOnly bool, mode string) (bool, error) {
	if len(keyPath) == 0 {
		return false, fmt.Errorf("invalid key path")
	}

	// Use config mode if not supplied on CLI
	if len(mode) == 0 {
		mode = m.Config.Mode.String()
	}

	// Only need to create one command for alias only mode
	if m.Config.AliasesOnly || aliasOnly {
		cmd := newCommand(command, keyPath, description, code, true, mode)
		m.Commands[cmd.Alias] = cmd
		return true, nil
	}

	// Build the root command first using the first key
	var isRoot bool
	key := keypath.Keys(keyPath)[0]
	cmd := m.Commands[key]
	if cmd == nil {
		// Create new command to build our the rest
		cmd = newCommand("", key, "", nil, false, mode)
		m.Commands[cmd.Alias] = cmd
		isRoot = true
	}

	// Modify or build the rest of the key path of commands
	cmd.build(keyPath, command, description, code, aliasOnly, mode)

	return isRoot, nil
}

// RemoveCommand tree at key path
func (m *Manifest) RemoveCommand(keyPath string) (bool, error) {
	cmd := m.Find(keyPath)
	if cmd == nil {
		return false, fmt.Errorf("command not found")
	}

	// Track if root command
	_, isRoot := m.Commands[keyPath]

	parent := cmd.parent
	if parent == nil {
		delete(m.Commands, keyPath)
		return isRoot, nil
	}

	parent.removeCommand(cmd)

	return isRoot, nil
}

// AddSubstitution with name and alias at key path
func (m *Manifest) AddSubstitution(keyPath, name, alias string) error {
	cmd := m.Find(keyPath)
	if cmd == nil {
		return fmt.Errorf("command not found")
	}

	s := &Substitution{name, alias}
	cmd.addSubstitution(s)

	return nil
}

// RemoveSubstitution at key path for given alias
func (m *Manifest) RemoveSubstitution(keyPath, alias string) error {
	cmd := m.Find(keyPath)
	if cmd == nil {
		return fmt.Errorf("command not found")
	}

	s := &Substitution{"", alias}
	cmd.removeSubstitution(s)

	return nil
}

// Find command at key path or nil if missing
func (m *Manifest) Find(keyPath string) *Command {
	for _, cmd := range m.Commands {
		if c := cmd.find(keyPath); c != nil {
			return c
		}
	}
	return nil
}

// AsJSON returns string representation of manifest
func (m *Manifest) AsJSON() string {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// AsYAML returns string representation of manifest
func (m *Manifest) AsYAML() string {
	b, err := yaml.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

// ExecutionString from input if possible or return error
func (m *Manifest) ExecutionString(args []string) (string, string, error) {
	for _, cmd := range m.Commands {
		keyPath := cmd.shortestKeyPath(keypath.KeyPath(args))
		if len(keyPath) > 0 {
			count := len(keypath.Keys(keyPath))

			log.Debug("key path:", keyPath)
			if len(args[count:]) > 0 {
				log.Debug("arguments:", args[count:])
			}

			c := cmd.find(keyPath)
			return c.Code.Language, c.executionString(args[count:]), nil
		}
	}

	log.Debug("arguments:", args)

	return "", "", fmt.Errorf("unable to execute command '%s'", strings.Join(args, " "))
}

// Keys as ordered list of fields for logging
func (m *Manifest) Keys() []string {
	return []string{"version", "commands"}
}

// Fields interface for logging
func (m *Manifest) Fields() map[string]interface{} {
	return map[string]interface{}{
		"version":  m.Version,
		"commands": joinedCommands(m.Commands),
	}
}

// Data method for Node interface to print tree
func (m *Manifest) Data() interface{} {
	return "manifest"
}

// Children method for Node interface to print tree
func (m *Manifest) Children() []tree.Node {
	nodes := make([]tree.Node, 0, len(m.Commands))
	for _, v := range m.Commands {
		nodes = append(nodes, v)
	}
	return nodes
}

// count of the total number of commands in this manifest
func (m *Manifest) count() int {
	count := 0
	for _, cmd := range m.Commands {
		count += cmd.count()
	}
	return count
}

func joinedCommands(cmdMap map[string]*Command) string {
	commands := []string{}
	for cmd := range cmdMap {
		commands = append(commands, cmd)
	}
	return strings.Join(commands, ", ")
}
