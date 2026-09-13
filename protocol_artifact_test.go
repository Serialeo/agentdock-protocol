package protocol

import (
	"encoding/json"
	"testing"
)

func TestArtifactBridgeWireLiteralsStayStable(t *testing.T) {
<<<<<<< HEAD
	if ConnectionProtocolVersion != "2" {
		t.Fatalf("ConnectionProtocolVersion = %q, want 2", ConnectionProtocolVersion)
=======
	if ConnectionProtocolVersion != "4" {
		t.Fatalf("ConnectionProtocolVersion = %q, want 4", ConnectionProtocolVersion)
>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)
	}
	if OperationArtifactRead != "artifact.read" {
		t.Fatalf("OperationArtifactRead = %q", OperationArtifactRead)
	}
	if ArtifactReadCapability != "bridge.artifact.read.v1" {
		t.Fatalf("ArtifactReadCapability = %q", ArtifactReadCapability)
	}
	if MaxArtifactChunkBytes != 512<<10 {
		t.Fatalf("MaxArtifactChunkBytes = %d", MaxArtifactChunkBytes)
	}
}

func TestHelloAdditiveFieldsRemainWireCompatible(t *testing.T) {
	encoded := []byte(`{
		"type":"node.hello",
<<<<<<< HEAD
		"protocol_version":"2",
		"hello":{
			"device_id":"device_abcdefgh",
			"version":"0.8.0",
			"protocol_version":"2",
=======
		"protocol_version":"4",
		"hello":{
			"device_id":"device_abcdefgh",
			"version":"0.8.0",
			"protocol_version":"4",
>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)
			"os":"darwin",
			"arch":"arm64",
			"capabilities":["read_file"],
			"bridge_capabilities":["bridge.artifact.read.v1"],
			"tool_contract_hash":"",
			"tools":[],
			"ui_resources":[],
			"future_additive_field":{"supported":true}
		}
	}`)

	var message Message
	if err := json.Unmarshal(encoded, &message); err != nil {
<<<<<<< HEAD
		t.Fatalf("additive Hello field must be ignored by older-compatible decoding: %v", err)
=======
		t.Fatalf("unknown additive Hello field should remain forward-tolerant within Bridge v4: %v", err)
>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)
	}
	if message.Hello == nil || len(message.Hello.BridgeCapabilities) != 1 || message.Hello.BridgeCapabilities[0] != ArtifactReadCapability {
		t.Fatalf("decoded hello = %#v", message.Hello)
	}
}
