package model

import (
	"math"
	"reflect"
	"testing"

	"github.com/shivamMg/ppds/tree"

	"github.com/pokanop/nostromo/keypath"
)

var depthKeys = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name        string
		cmdName     string
		alias       string
		aliasOnly   bool
		description string
		code        *Code
		expected    *Command
	}{
		{"empty alias", "cmd", "", false, "", nil, &Command{nil, "cmd", "cmd", "cmd", false, "", map[string]*Command{}, map[string]*Substitution{}, &Code{}, ConcatenateMode, false}},
		{"empty name", "", "alias", false, "", nil, &Command{nil, "alias", "", "alias", false, "", map[string]*Command{}, map[string]*Substitution{}, &Code{}, ConcatenateMode, false}},
		{"valid alias", "cmd", "cmd-alias", false, "description", nil, &Command{nil, "cmd-alias", "cmd", "cmd-alias", false, "description", map[string]*Command{}, map[string]*Substitution{}, &Code{}, ConcatenateMode, false}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := newCommand(test.cmdName, test.alias, test.description, test.code, test.aliasOnly, ConcatenateMode.String())
			if !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestAddCommand(t *testing.T) {
	tests := []struct {
		name    string
		command *Command
		add     *Command
	}{
		{"nil command", fakeCommand(1), nil},
		{"invalid command", fakeCommand(1), fakeCommand(2).Commands["two"]},
		{"valid command", fakeCommand(1), fakeCommand(1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.addCommand(test.add)
			if test.add != nil && test.command.Commands[test.add.Alias] == nil {
				t.Errorf("expected command to be added but was not")
			}
		})
	}
}

func TestRemoveCommand(t *testing.T) {
	tests := []struct {
		name    string
		command *Command
		remove  *Command
	}{
		{"nil command", fakeCommand(1), nil},
		{"invalid command", fakeCommand(1), fakeCommand(2).Commands["two"]},
		{"valid command", fakeCommand(1), fakeCommand(1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.removeCommand(test.remove)
			if test.remove != nil && test.command.Commands[test.remove.Alias] != nil {
				t.Errorf("expected command to be removed but was not")
			}
		})
	}
}

func TestAddSubstitution(t *testing.T) {
	tests := []struct {
		name    string
		command *Command
		add     *Substitution
	}{
		{"nil sub", fakeCommand(1), nil},
		{"invalid sub", fakeCommand(1), fakeCommand(2).Subs["one"]},
		{"valid sub", fakeCommand(1), &Substitution{"two", ""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.addSubstitution(test.add)
			if test.add != nil && test.command.Subs[test.add.Alias] == nil {
				t.Errorf("expected sub to be added but was not")
			}
		})
	}
}

func TestRemoveSubstitution(t *testing.T) {
	tests := []struct {
		name    string
		command *Command
		remove  *Substitution
	}{
		{"nil sub", fakeCommand(1), nil},
		{"invalid sub", fakeCommand(1), fakeCommand(2).Subs["one"]},
		{"valid sub", fakeCommand(1), &Substitution{"two", ""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.removeSubstitution(test.remove)
			if test.remove != nil && test.command.Subs[test.remove.Alias] != nil {
				t.Errorf("expected sub to be removed but was not")
			}
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		keyPath  string
		expected *Command
	}{
		{"empty key path", fakeCommand(1), "", nil},
		{"wrong key path", fakeCommand(1), "wrong", nil},
		{"valid key path first level", fakeCommand(1), "one-alias", fakeCommand(1)},
		{"valid key path nth level", fakeCommand(4), "one-alias.two-alias.three-alias", fakeCommand(4).Commands["two-alias"].Commands["three-alias"]},
		{"valid key path last level", fakeCommand(7), "one-alias.two-alias.three-alias.four-alias.five-alias.six-alias.seven-alias", fakeCommand(7).Commands["two-alias"].Commands["three-alias"].Commands["four-alias"].Commands["five-alias"].Commands["six-alias"].Commands["seven-alias"]},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.command.find(test.keyPath)
			if !reflect.DeepEqual(test.expected, actual) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestShortestKeyPath(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		keyPath  string
		expected string
	}{
		{"empty key path", fakeCommand(1), "", ""},
		{"missing key path", fakeCommand(1), "missing", ""},
		{"missing long key path", fakeCommand(1), "this.is.missing", ""},
		{"valid key path first level", fakeCommand(1), "one-alias", "one-alias"},
		{"valid key path nth level", fakeCommand(3), "one-alias.two-alias", "one-alias.two-alias"},
		{"valid key path last level", fakeCommand(4), "one-alias.two-alias.three-alias.four-alias", "one-alias.two-alias.three-alias.four-alias"},
		{"valid key path shortened", fakeCommand(2), "one-alias.two-alias.three-alias.four-alias", "one-alias.two-alias"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.command.shortestKeyPath(test.keyPath); test.expected != actual {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestExecutionString(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		args     []string
		expected string
	}{
		{"one level nil args", fakeCommand(1), nil, "one"},
		{"one level empty args", fakeCommand(1), []string{}, "one"},
		{"one level no dot arg", fakeCommand(1), []string{"arg"}, "one arg"},
		{"one level dot arg", fakeCommand(1), []string{"arg.1"}, "one arg.1"},
		{"n level no dot args", fakeCommand(3).Commands["two-alias"].Commands["three-alias"], []string{"arg1", "arg2"}, "one two three arg1 arg2"},
		{"n level dot args", fakeCommand(4).Commands["two-alias"], []string{"arg.1", "arg2", "arg.3"}, "one two arg.1 arg2 arg.3"},
		{"n level dot sub args", fakeCommand(4).Commands["two-alias"], []string{"arg.1", "one-sub", "two-sub"}, "one two arg.1 one two"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.command.executionString(test.args); test.expected != actual {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestReverseWalk(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		fn       func(*Command, *bool)
		expected *Command
	}{
		{"nil fn", fakeCommand(1), nil, fakeCommand(1)},
		{"stop fn", fakeCommand(1), func(cmd *Command, stop *bool) { *stop = true }, fakeCommand(1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.reverseWalk(test.fn)
			if !reflect.DeepEqual(test.command, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, test.command)
			}
		})
	}
}

func TestForwardWalk(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		fn       func(*Command, *bool)
		expected *Command
	}{
		{"nil fn", fakeCommand(1), nil, fakeCommand(1)},
		{"stop fn", fakeCommand(1), func(cmd *Command, stop *bool) { *stop = true }, fakeCommand(1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.forwardWalk(test.fn)
			if !reflect.DeepEqual(test.command, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, test.command)
			}
		})
	}
}

func TestWalk(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		fn       func(*Command, *bool)
		expected *Command
	}{
		{"nil fn", fakeCommand(1), nil, fakeCommand(1)},
		{"stop fn", fakeCommand(1), func(cmd *Command, stop *bool) { *stop = true }, fakeCommand(1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.Walk(test.fn)
			if !reflect.DeepEqual(test.command, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, test.command)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name       string
		command    *Command
		keyPath    string
		commandStr string
		aliasOnly  bool
		expected   *Command
	}{
		{"empty key path and command", fakeCommand(1), "", "", false, fakeCommand(1)},
		{"empty key path", fakeCommand(1), "", "command", false, fakeCommand(1)},
		{"empty command", fakeCommand(1), "key path", "", false, fakeCommand(1)},
		{"single no change", fakeCommand(1), "one-alias", "one", false, fakeCommand(1)},
		{"single change", fakeCommand(1), "one-alias", "diff", false, fakeBuiltCommand(1, 1, "one-alias", "diff")},
		{"multi no change", fakeCommand(3), "one-alias.two-alias.three-alias", "three", false, fakeCommand(3)},
		{"multi mid change", fakeCommand(3), "one-alias.two-alias", "diff", false, fakeBuiltCommand(3, 2, "one-alias.two-alias", "diff")},
		{"multi last change", fakeCommand(2), "one-alias.two-alias", "diff", false, fakeBuiltCommand(2, 2, "one-alias.two-alias", "diff")},
		{"multi add no command", fakeCommand(2), "one-alias.two-alias.three-alias.four-alias", "", false, fakeBuiltCommand(2, 4, "one-alias.two-alias.three-alias.four-alias", "")},
		{"multi add command", fakeCommand(1), "one-alias.two-alias.three-alias.four-alias", "four", false, fakeBuiltCommand(1, 4, "one-alias.two-alias.three-alias.four-alias", "four")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.command.build(test.keyPath, test.commandStr, "", &Code{}, test.aliasOnly, ConcatenateMode.String())
			if !reflect.DeepEqual(test.expected, test.command) {
				t.Errorf("expected: %s, actual: %s", test.expected, test.command)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		expected string
	}{
		{"single command", fakeCommand(1), "[one-alias] one -> one-alias"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.command.String(); actual != test.expected {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		expected []string
	}{
		{"keys", fakeCommand(1), []string{"keypath", "alias", "command", "description", "commands", "substitutions", "code", "mode", "aliasOnly", "disabled"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.command.Keys(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		expected map[string]interface{}
	}{
		{
			"keys",
			fakeCommand(1),
			map[string]interface{}{
				"alias":         "one-alias",
				"command":       "one",
				"description":   "",
				"commands":      "",
				"substitutions": "one-sub",
				"code":          false,
				"keypath":       "one-alias",
				"mode":          "concatenate",
				"aliasOnly":     false,
				"disabled":      false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.command.Fields(); !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %s, actual: %s", test.expected, actual)
			}
		})
	}
}

func TestCommand_Children(t *testing.T) {
	commands := map[string]*Command{
		"foo": {},
		"bar": {},
	}
	type fields struct {
		Commands map[string]*Command
	}
	tests := []struct {
		name   string
		fields fields
		want   []tree.Node
	}{
		{"children", fields{commands}, []tree.Node{commands["foo"], commands["bar"]}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{
				Commands: tt.fields.Commands,
			}
			if got := c.Children(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Children() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandData(t *testing.T) {
	type fields struct {
		Alias string
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{"data", fields{"foo"}, "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{
				Alias: tt.fields.Alias,
			}
			if got := c.Data(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Data() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fakeCommand(depth int) *Command {
	return fakeCommandWithPrefix(depth, "")
}

func fakeCommandWithPrefix(depth int, prefix string) *Command {
	var firstCmd *Command
	var lastCmd *Command
	var cmd *Command
	for i := 0; i < depth; i++ {
		name := depthKeys[i+1]
		cmd = newCommand(prefix+name, prefix+name+"-alias", "", nil, false, ConcatenateMode.String())
		cmd.addSubstitution(&Substitution{prefix + name, prefix + name + "-sub"})
		if lastCmd != nil {
			lastCmd.addCommand(cmd)
		} else {
			firstCmd = cmd
		}
		lastCmd = cmd
	}
	return firstCmd
}

func fakeBuiltCommand(startDepth, endDepth int, keyPath, command string) *Command {
	first := fakeCommand(int(math.Max(float64(startDepth), float64(endDepth))))
	cmd := first
	keys := keypath.Keys(keyPath)
	for i := 1; i < endDepth; i++ {
		cmd = cmd.Commands[keys[i]]
		if i >= startDepth {
			cmd.Name = ""
			cmd.Subs = map[string]*Substitution{}
		}
	}
	if len(command) > 0 {
		cmd.Name = command
	}
	return first
}

func TestCommandEffectiveCommand(t *testing.T) {
	type fields struct {
		Name string
		Code *Code
		Mode Mode
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"use code", fields{"", &Code{"js", "code"}, ConcatenateMode}, "code"},
		{"concatenate", fields{"command", nil, ConcatenateMode}, "command"},
		{"independent", fields{"command", nil, IndependentMode}, "command;"},
		{"exclusive", fields{"command", nil, ExclusiveMode}, "command;"},
		{"empty command", fields{"", nil, ExclusiveMode}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{
				Name: tt.fields.Name,
				Code: tt.fields.Code,
				Mode: tt.fields.Mode,
			}
			if got := c.effectiveCommand(); got != tt.want {
				t.Errorf("effectiveCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
