package actionguard

import "testing"

func BenchmarkPolicySetCheckThreePolicies(b *testing.B) {
	set := NewPolicySet().
		With(NewAllowList("read_file", "list_dir")).
		With(NewDenyList("shell_exec", "rm_rf"))
	call := NewToolCall("read_file", map[string]any{"path": "/tmp/a.txt"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Check(call)
	}
}
