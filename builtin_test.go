package protocol

import (
	"encoding/json"
	"testing"
)

func TestBuiltinSnapshotKeepsReleaseChoiceAndReadinessSeparate(t *testing.T) {
	original := Message{Type: MessageNodeUpdated, ProtocolVersion: ConnectionProtocolVersion, Hello: &Hello{Builtins: []BuiltinCapability{
		{ID: "acp", Provided: false, Enabled: true, Ready: false, Available: false, Reason: "excluded by distribution", Tools: []string{"acp_session"}},
		{ID: "browser", Provided: true, Enabled: true, Ready: false, Available: false, Reason: "CDP unavailable", Tools: []string{"browser_session"}},
	}}}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Type != MessageNodeUpdated || len(decoded.Hello.Builtins) != 2 {
		t.Fatal("snapshot lost")
	}
	excluded := decoded.Hello.Builtins[0]
	if excluded.Provided || !excluded.Enabled || excluded.Available || excluded.Reason == "" {
		t.Fatalf("independent states collapsed: %+v", excluded)
	}
	var omitted BuiltinUpdate
	if err := json.Unmarshal([]byte(`{"id":"browser"}`), &omitted); err != nil {
		t.Fatal(err)
	}
	if omitted.Enabled != nil {
		t.Fatal("omitted choice became a disable request")
	}
}
