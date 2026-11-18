package terminalApi

type Parser interface {
	Parse(args []string) (Invocation, error)
}

type FlagType string

const (
	StringFlag FlagType = "string"
	BoolFlag   FlagType = "bool"
	ListFlag   FlagType = "list"
)

type Flag struct {
	Name          string
	Short         string
	Description   string
	Type          FlagType
	StringDefault string
	BoolDefault   bool
	ListDefault   []string
	Required      bool
	Persistent    bool
}

type Command struct {
	Name        string
	Description string
	Flags       []Flag
	Commands    []Command
}

type FlagValue struct {
	Type    FlagType
	String  string
	Bool    bool
	List    []string
	Changed bool
}

type Invocation struct {
	Command       []string
	Arguments     []string
	Flags         map[string]FlagValue
	HelpRequested bool
	Help          string
}
