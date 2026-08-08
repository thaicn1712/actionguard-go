// shell_exec is both on an allow list and a deny list — deny always wins,
// and an unlisted tool call is denied by default (fail-closed).
package main

import (
	"fmt"

	"github.com/thaicn1712/actionguard-go"
)

func main() {
	policies := actionguard.NewPolicySet().
		With(actionguard.NewAllowList("read_file", "shell_exec")).
		With(actionguard.NewDenyList("shell_exec"))

	for _, name := range []string{"read_file", "shell_exec", "delete_database"} {
		decision := policies.Check(actionguard.NewToolCall(name, nil))
		fmt.Printf("%-16s allowed=%v reason=%q\n", name, decision.Allowed, decision.Reason)
	}
}
