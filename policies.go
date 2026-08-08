package actionguard

import (
	"fmt"
	"regexp"
)

// AllowList votes Allow for calls whose name is in Names, Abstain otherwise.
type AllowList struct {
	Names map[string]struct{}
}

// NewAllowList builds an AllowList from a slice of tool names.
func NewAllowList(names ...string) AllowList {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return AllowList{Names: set}
}

func (a AllowList) Vote(call ToolCall) (Vote, string) {
	if _, ok := a.Names[call.Name]; ok {
		return VoteAllow, ""
	}
	return VoteAbstain, ""
}

// DenyList votes Deny for calls whose name is in Names, Abstain otherwise.
type DenyList struct {
	Names map[string]struct{}
}

// NewDenyList builds a DenyList from a slice of tool names.
func NewDenyList(names ...string) DenyList {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return DenyList{Names: set}
}

func (d DenyList) Vote(call ToolCall) (Vote, string) {
	if _, ok := d.Names[call.Name]; ok {
		return VoteDeny, fmt.Sprintf("tool %q is on the deny list", call.Name)
	}
	return VoteAbstain, ""
}

// ArgMatchesRegex votes Deny when call.Name matches ToolName and the string
// argument named ArgName either doesn't exist or doesn't match Pattern.
// Abstains for any other tool name, so it can sit alongside policies for
// other tools in the same PolicySet.
type ArgMatchesRegex struct {
	ToolName string
	ArgName  string
	Pattern  *regexp.Regexp
}

// NewArgMatchesRegex compiles pattern and builds an ArgMatchesRegex policy.
func NewArgMatchesRegex(toolName, argName, pattern string) (ArgMatchesRegex, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ArgMatchesRegex{}, err
	}
	return ArgMatchesRegex{ToolName: toolName, ArgName: argName, Pattern: re}, nil
}

func (a ArgMatchesRegex) Vote(call ToolCall) (Vote, string) {
	if call.Name != a.ToolName {
		return VoteAbstain, ""
	}
	value, ok := call.Arguments[a.ArgName].(string)
	if !ok || !a.Pattern.MatchString(value) {
		return VoteDeny, fmt.Sprintf(
			"argument %q on tool %q does not match required pattern %q",
			a.ArgName, a.ToolName, a.Pattern.String(),
		)
	}
	return VoteAllow, ""
}
