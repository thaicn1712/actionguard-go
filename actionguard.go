// Package actionguard provides policy-as-code (deny-overrides, fail-closed)
// for tool calls an agent is about to make, in Go.
package actionguard

import "context"

// ToolCall is a tool an agent is about to invoke, with the arguments it
// intends to pass.
type ToolCall struct {
	Name      string
	Arguments map[string]any
}

// NewToolCall builds a ToolCall.
func NewToolCall(name string, arguments map[string]any) ToolCall {
	return ToolCall{Name: name, Arguments: arguments}
}

// Vote is one Policy's opinion on a ToolCall.
type Vote int

const (
	// VoteAbstain means the policy has no opinion on this call.
	VoteAbstain Vote = iota
	// VoteAllow permits the call, unless a later policy denies it.
	VoteAllow
	// VoteDeny blocks the call outright — deny always overrides allow.
	VoteDeny
)

// Decision is the outcome of checking a ToolCall against a PolicySet.
type Decision struct {
	Allowed bool
	// Reason explains a denial. Empty when Allowed is true.
	Reason string
}

// Policy votes on a ToolCall synchronously.
type Policy interface {
	// Vote returns the policy's vote, and a reason (used only for VoteDeny).
	Vote(call ToolCall) (Vote, string)
}

// AsyncPolicy votes on a ToolCall against something that needs network
// access — an external authorization service, an LLM-as-judge call.
type AsyncPolicy interface {
	Vote(ctx context.Context, call ToolCall) (Vote, string, error)
}

// PolicySet checks a ToolCall against a sequence of Policies. Deny-overrides:
// any single VoteDeny wins immediately, regardless of order. Fail-closed: if
// no policy affirmatively allows the call, it's denied by default — a tool
// call is never allowed just because nothing objected to it.
type PolicySet struct {
	policies []Policy
}

// NewPolicySet creates an empty PolicySet.
func NewPolicySet() *PolicySet {
	return &PolicySet{}
}

// With appends a Policy and returns the PolicySet for chaining.
func (p *PolicySet) With(policy Policy) *PolicySet {
	p.policies = append(p.policies, policy)
	return p
}

// Check evaluates every policy, deny-overrides, fail-closed.
func (p *PolicySet) Check(call ToolCall) Decision {
	sawAllow := false
	for _, policy := range p.policies {
		switch vote, reason := policy.Vote(call); vote {
		case VoteDeny:
			return Decision{Allowed: false, Reason: reason}
		case VoteAllow:
			sawAllow = true
		}
	}
	if sawAllow {
		return Decision{Allowed: true}
	}
	return Decision{Allowed: false, Reason: "no policy allowed this tool call (fail-closed)"}
}

// AsyncPolicySet is PolicySet for AsyncPolicy — same deny-overrides,
// fail-closed evaluation, with network calls allowed along the way.
type AsyncPolicySet struct {
	policies []AsyncPolicy
}

// NewAsyncPolicySet creates an empty AsyncPolicySet.
func NewAsyncPolicySet() *AsyncPolicySet {
	return &AsyncPolicySet{}
}

// With appends an AsyncPolicy and returns the AsyncPolicySet for chaining.
func (p *AsyncPolicySet) With(policy AsyncPolicy) *AsyncPolicySet {
	p.policies = append(p.policies, policy)
	return p
}

// Check evaluates every policy in order, deny-overrides, fail-closed.
// Returns an error only if a policy itself errors.
func (p *AsyncPolicySet) Check(ctx context.Context, call ToolCall) (Decision, error) {
	sawAllow := false
	for _, policy := range p.policies {
		vote, reason, err := policy.Vote(ctx, call)
		if err != nil {
			return Decision{}, err
		}
		switch vote {
		case VoteDeny:
			return Decision{Allowed: false, Reason: reason}, nil
		case VoteAllow:
			sawAllow = true
		}
	}
	if sawAllow {
		return Decision{Allowed: true}, nil
	}
	return Decision{Allowed: false, Reason: "no policy allowed this tool call (fail-closed)"}, nil
}
