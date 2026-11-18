package terminalApi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`)
var validShort = regexp.MustCompile(`^[a-zA-Z0-9]$`)

type commandParser struct {
	spec Command
}

type explicitBool bool

func (value *explicitBool) Set(text string) error {
	parsed, err := strconv.ParseBool(text)
	if err != nil {
		return fmt.Errorf("invalid boolean %q: %w", text, err)
	}
	*value = explicitBool(parsed)
	return nil
}

func (value *explicitBool) String() string { return strconv.FormatBool(bool(*value)) }
func (value *explicitBool) Type() string   { return "bool" }

func NewParser(root Command) (Parser, error) {
	copy := cloneCommand(root)
	if err := validateCommand(copy, nil, nil); err != nil {
		return nil, err
	}
	return &commandParser{spec: copy}, nil
}

func cloneCommand(command Command) Command {
	command.Flags = append([]Flag(nil), command.Flags...)
	for i := range command.Flags {
		command.Flags[i].ListDefault = append([]string(nil), command.Flags[i].ListDefault...)
	}
	children := make([]Command, len(command.Commands))
	for i := range command.Commands {
		children[i] = cloneCommand(command.Commands[i])
	}
	command.Commands = children
	return command
}

func validateCommand(command Command, inheritedNames, inheritedShorts map[string]bool) error {
	if !validName.MatchString(command.Name) || command.Name == "help" {
		return fmt.Errorf("invalid command name %q", command.Name)
	}
	names, shorts := make(map[string]bool), make(map[string]bool)
	persistentNames, persistentShorts := make(map[string]bool), make(map[string]bool)
	for name := range inheritedNames {
		names[name], persistentNames[name] = true, true
	}
	for short := range inheritedShorts {
		shorts[short], persistentShorts[short] = true, true
	}
	for _, flag := range command.Flags {
		if !validName.MatchString(flag.Name) || flag.Name == "help" || names[flag.Name] {
			return fmt.Errorf("invalid or duplicate flag %q in %s", flag.Name, command.Name)
		}
		if flag.Short != "" && (!validShort.MatchString(flag.Short) || flag.Short == "h" || shorts[flag.Short]) {
			return fmt.Errorf("invalid or duplicate shorthand %q in %s", flag.Short, command.Name)
		}
		switch flag.Type {
		case StringFlag, BoolFlag, ListFlag:
		default:
			return fmt.Errorf("invalid type %q for flag %s", flag.Type, flag.Name)
		}
		names[flag.Name] = true
		if flag.Short != "" {
			shorts[flag.Short] = true
		}
		if flag.Persistent {
			persistentNames[flag.Name] = true
			if flag.Short != "" {
				persistentShorts[flag.Short] = true
			}
		}
	}
	children := make(map[string]bool)
	for _, child := range command.Commands {
		if children[child.Name] {
			return fmt.Errorf("duplicate command %q in %s", child.Name, command.Name)
		}
		children[child.Name] = true
		if err := validateCommand(child, persistentNames, persistentShorts); err != nil {
			return err
		}
	}
	return nil
}

func (p *commandParser) Parse(args []string) (Invocation, error) {
	var result Invocation
	root, err := buildCommand(p.spec, &result)
	if err != nil {
		return Invocation{}, err
	}
	root.SilenceErrors, root.SilenceUsage = true, true
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		result.Command = strings.Fields(cmd.CommandPath())
		result.HelpRequested = true
		result.Help = cmd.UsageString()
	})
	root.SetArgs(append([]string{}, args...))
	if _, err := root.ExecuteC(); err != nil {
		return Invocation{}, err
	}
	return result, nil
}

func buildCommand(spec Command, result *Invocation) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   spec.Name,
		Short: spec.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.HasAvailableSubCommands() && len(args) > 0 {
				return fmt.Errorf("unknown command %q for %s", args[0], cmd.CommandPath())
			}
			result.Command = strings.Fields(cmd.CommandPath())
			result.Arguments = append([]string(nil), args...)
			result.Flags = make(map[string]FlagValue)
			var valueErr error
			collect := func(flag *pflag.Flag) {
				if valueErr != nil || flag.Name == "help" {
					return
				}
				value := FlagValue{Changed: flag.Changed}
				switch flag.Value.Type() {
				case "string":
					value.Type = StringFlag
					value.String, valueErr = cmd.Flags().GetString(flag.Name)
				case "bool":
					value.Type = BoolFlag
					value.Bool, valueErr = cmd.Flags().GetBool(flag.Name)
				case "stringSlice":
					value.Type = ListFlag
					value.List, valueErr = cmd.Flags().GetStringSlice(flag.Name)
				default:
					valueErr = fmt.Errorf("unsupported flag type %q", flag.Value.Type())
				}
				result.Flags[flag.Name] = value
			}
			cmd.Flags().VisitAll(collect)
			cmd.InheritedFlags().VisitAll(collect)
			return valueErr
		},
	}
	for _, flag := range spec.Flags {
		flags := cmd.Flags()
		if flag.Persistent {
			flags = cmd.PersistentFlags()
		}
		switch flag.Type {
		case StringFlag:
			flags.StringP(flag.Name, flag.Short, flag.StringDefault, flag.Description)
		case BoolFlag:
			value := explicitBool(flag.BoolDefault)
			flags.VarP(&value, flag.Name, flag.Short, flag.Description)
		case ListFlag:
			flags.StringSliceP(flag.Name, flag.Short, flag.ListDefault, flag.Description)
		}
		if flag.Required {
			if err := cobra.MarkFlagRequired(flags, flag.Name); err != nil {
				return nil, fmt.Errorf("mark flag %s required: %w", flag.Name, err)
			}
		}
	}
	for _, child := range spec.Commands {
		childCommand, err := buildCommand(child, result)
		if err != nil {
			return nil, err
		}
		cmd.AddCommand(childCommand)
	}
	return cmd, nil
}
