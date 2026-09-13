package protocol

import (
	"encoding/json"
	"testing"
)

func TestUIResourceContractsAreExplicit(t *testing.T) {
	cases := map[string]string{
		WorkContinuationUIResourceURI: WorkContinuationUIContract,
		ContextUIResourceURI:          ContextUIContract,
		TaskProgressUIResourceURI:     TaskProgressUIContract,
		FileChangeUIResourceURI:       FileChangeUIContract,
		RecallUIResourceURI:           RecallUIContract,
		WorkflowUIResourceURI:         WorkflowUIContract,
		DynamicMCPUIResourceURI:       DynamicMCPUIContract,
		ACPStatusUIResourceURI:        ACPStatusUIContract,
	}
	for uri, want := range cases {
		got, ok := UIResourceContract(uri)
		if !ok || got != want {
			t.Fatalf("UIResourceContract(%q) = %q, %v; want %q, true", uri, got, ok, want)
		}
	}
	if _, ok := UIResourceContract("ui://agentdock/unknown"); ok {
		t.Fatal("unknown AgentDock UI resource unexpectedly has a contract")
	}
}

func TestHelloRoundTripKeepsUIResourcesSeparateFromToolMeta(t *testing.T) {
	original := Message{
		Type:            MessageNodeHello,
		ProtocolVersion: ConnectionProtocolVersion,
		Hello: &Hello{
			DeviceID:           "device_abcdefgh",
			ProtocolVersion:    ConnectionProtocolVersion,
			Capabilities:       []string{"read_file"},
			BridgeCapabilities: []string{ArtifactReadCapability},
			Tools: []ToolDescriptor{{
				Name:        "workflow_template_manage",
				InputSchema: map[string]any{"type": "object"},
			}},
			UIResources: []UIResourceCapability{{
				URI: WorkflowUIResourceURI, Contract: WorkflowUIContract, MIMEType: MCPAppMIMEType,
			}},
		},
	}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Message
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ProtocolVersion != ConnectionProtocolVersion || decoded.Hello == nil {
		t.Fatalf("decoded message = %#v", decoded)
	}
	if len(decoded.Hello.UIResources) != 1 || decoded.Hello.UIResources[0].URI != WorkflowUIResourceURI {
		t.Fatalf("decoded ui_resources = %#v", decoded.Hello.UIResources)
	}
	if len(decoded.Hello.Capabilities) != 1 || decoded.Hello.Capabilities[0] != "read_file" {
		t.Fatalf("decoded capabilities = %#v", decoded.Hello.Capabilities)
	}
	if len(decoded.Hello.BridgeCapabilities) != 1 || decoded.Hello.BridgeCapabilities[0] != ArtifactReadCapability {
		t.Fatalf("decoded bridge_capabilities = %#v", decoded.Hello.BridgeCapabilities)
	}
	if decoded.Hello.Tools[0].Meta != nil {
		t.Fatalf("workflow tool unexpectedly carries static UI meta: %#v", decoded.Hello.Tools[0].Meta)
	}
}
