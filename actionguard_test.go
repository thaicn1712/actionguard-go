package actionguard

import (
	"context"
	"errors"
	"testing"
)

func TestPolicySetAllowListPermitsListedTool(t *testing.T) {
	set := NewPolicySet().With(NewAllowList("read_file"))
	decision := set.Check(NewToolCall("read_file", nil))
	if !decision.Allowed {
		t.Fatalf("expected allow, got deny: %s", decision.Reason)
	}
}

func TestPolicySetFailsClosedOnUnknownTool(t *testing.T) {
	set := NewPolicySet().With(NewAllowList("read_file"))
	decision := set.Check(NewToolCall("shell_exec", nil))
	if decision.Allowed {
		t.Fatal("expected deny for a tool no policy allowed")
	}
}

func TestPolicySetDenyOverridesAllow(t *testing.T) {
	set := NewPolicySet().
		With(NewAllowList("shell_exec")).
		With(NewDenyList("shell_exec"))
	decision := set.Check(NewToolCall("shell_exec", nil))
	if decision.Allowed {
		t.Fatal("expected deny to override allow")
	}
}

func TestPolicySetEmptySetFailsClosed(t *testing.T) {
	set := NewPolicySet()
	decision := set.Check(NewToolCall("anything", nil))
	if decision.Allowed {
		t.Fatal("expected an empty policy set to deny everything")
	}
}

func TestArgMatchesRegexAllowsMatchingArgument(t *testing.T) {
	policy, err := NewArgMatchesRegex("write_file", "path", `^/tmp/.*`)
	if err != nil {
		t.Fatal(err)
	}
	set := NewPolicySet().With(policy)

	decision := set.Check(NewToolCall("write_file", map[string]any{"path": "/tmp/scratch.txt"}))
	if !decision.Allowed {
		t.Fatalf("expected allow, got deny: %s", decision.Reason)
	}
}

func TestArgMatchesRegexDeniesNonMatchingArgument(t *testing.T) {
	policy, err := NewArgMatchesRegex("write_file", "path", `^/tmp/.*`)
	if err != nil {
		t.Fatal(err)
	}
	set := NewPolicySet().With(policy)

	decision := set.Check(NewToolCall("write_file", map[string]any{"path": "/etc/passwd"}))
	if decision.Allowed {
		t.Fatal("expected deny for a path outside /tmp")
	}
}

func TestArgMatchesRegexDeniesMissingArgument(t *testing.T) {
	policy, err := NewArgMatchesRegex("write_file", "path", `^/tmp/.*`)
	if err != nil {
		t.Fatal(err)
	}
	set := NewPolicySet().With(policy)

	decision := set.Check(NewToolCall("write_file", map[string]any{}))
	if decision.Allowed {
		t.Fatal("expected deny when the required argument is missing")
	}
}

func TestArgMatchesRegexAbstainsForOtherTools(t *testing.T) {
	policy, err := NewArgMatchesRegex("write_file", "path", `^/tmp/.*`)
	if err != nil {
		t.Fatal(err)
	}
	set := NewPolicySet().With(NewAllowList("read_file")).With(policy)

	decision := set.Check(NewToolCall("read_file", nil))
	if !decision.Allowed {
		t.Fatalf("expected the unrelated policy's abstain to not affect this call: %s", decision.Reason)
	}
}

type asyncApproval struct {
	deny bool
}

func (a asyncApproval) Vote(ctx context.Context, call ToolCall) (Vote, string, error) {
	if a.deny {
		return VoteDeny, "external service rejected it", nil
	}
	return VoteAllow, "", nil
}

func TestAsyncPolicySetCheck(t *testing.T) {
	set := NewAsyncPolicySet().With(asyncApproval{deny: false})
	decision, err := set.Check(context.Background(), NewToolCall("anything", nil))
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed {
		t.Fatal("expected allow")
	}

	set = NewAsyncPolicySet().With(asyncApproval{deny: true})
	decision, err = set.Check(context.Background(), NewToolCall("anything", nil))
	if err != nil {
		t.Fatal(err)
	}
	if decision.Allowed {
		t.Fatal("expected deny")
	}
}

type erroringPolicy struct{}

func (erroringPolicy) Vote(ctx context.Context, call ToolCall) (Vote, string, error) {
	return VoteAbstain, "", errors.New("service unavailable")
}

func TestAsyncPolicySetPropagatesPolicyError(t *testing.T) {
	set := NewAsyncPolicySet().With(erroringPolicy{})
	_, err := set.Check(context.Background(), NewToolCall("anything", nil))
	if err == nil {
		t.Fatal("expected an error to propagate")
	}
}
