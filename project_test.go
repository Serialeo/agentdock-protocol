package protocol

import (
	"encoding/json"
	"testing"
)

func TestBridgeV4ToolCallRoundTripKeepsExecutionContextOutOfArguments(t *testing.T) {
	if ConnectionProtocolVersion != "4" {
		t.Fatalf("ConnectionProtocolVersion = %q, want 4", ConnectionProtocolVersion)
	}
	original := Message{
		Type:      MessageToolInvoke,
		RequestID: "req-1",
		Operation: OperationToolCall,
		ExecutionContext: &ExecutionContext{
			WorkSessionID:      "ws-1",
			TargetID:           "target-1",
			ProjectID:          "project-1",
			DeploymentID:       "deployment-1",
			DeploymentRevision: "rev-7",
			ContextRevision:    "sha256:context",
		},
		Arguments: json.RawMessage(`{"tool":"read_file","arguments":{"path":"README.md"}}`),
	}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Message
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ExecutionContext == nil || decoded.ExecutionContext.TargetID != "target-1" {
		t.Fatalf("execution context lost in round trip: %#v", decoded.ExecutionContext)
	}
	var toolRequest map[string]any
	if err := json.Unmarshal(decoded.Arguments, &toolRequest); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"node_id", "working_folder", "permissions", "target_id", "work_session_id"} {
		if _, exists := toolRequest[forbidden]; exists {
			t.Fatalf("routing field %q leaked into model-controlled tool arguments: %#v", forbidden, toolRequest)
		}
	}
}

func TestValidateExecutionContextRejectsMissingOrIncompleteContext(t *testing.T) {
	if got := ValidateExecutionContext(nil); got == nil || got.Code != ErrorExecutionContextRequired {
		t.Fatalf("nil context error = %#v", got)
	}
	complete := ExecutionContext{
		WorkSessionID:      "ws-1",
		TargetID:           "target-1",
		ProjectID:          "project-1",
		DeploymentID:       "deployment-1",
		DeploymentRevision: "rev-1",
		ContextRevision:    "sha256:context",
	}
	if got := ValidateExecutionContext(&complete); got != nil {
		t.Fatalf("complete context rejected: %#v", got)
	}
	complete.TargetID = "  "
	got := ValidateExecutionContext(&complete)
	if got == nil || got.Code != ErrorExecutionContextInvalid || got.Details["field"] != "target_id" {
		t.Fatalf("incomplete context error = %#v", got)
	}
}

func TestOnlyProjectToolCallsRequireExecutionContextAtBridgeContractLayer(t *testing.T) {
	if !OperationRequiresExecutionContext(OperationToolCall) {
		t.Fatal("tool.call must require Project execution context in Bridge v4")
	}
	for _, operation := range []string{OperationRuntimeRequest, OperationContextLocal, OperationResourceRead, OperationArtifactRead} {
		if OperationRequiresExecutionContext(operation) {
			t.Fatalf("management or transport operation %q unexpectedly requires a WorkSession Target", operation)
		}
	}
}

func TestDeploymentPermissionsValidateFileCapability(t *testing.T) {
	for _, files := range []FileCapability{FileCapabilityNone, FileCapabilityReadOnly, FileCapabilityReadWrite} {
		if err := (DeploymentPermissions{Computer: ComputerPermissionNone, Files: files}).Validate(); err != nil {
			t.Fatalf("valid files capability %q rejected: %v", files, err)
		}
	}
	if err := (DeploymentPermissions{Computer: ComputerPermissionNone, Files: FileCapability("sandbox")}).Validate(); err == nil {
		t.Fatal("unknown files capability unexpectedly accepted")
	}
}

func TestSourceProvenanceValidationAndWireRoundTrip(t *testing.T) {
	cases := []struct {
		name       string
		provenance SourceProvenance
		wantErr    bool
	}{
		{name: "non_git", provenance: SourceProvenance{Kind: SourceProvenanceNone}},
		{name: "clean_branch", provenance: SourceProvenance{Kind: SourceProvenanceGit, RepositoryRoot: "/work/repo", Head: "0123456789abcdef", Branch: "main"}},
		{name: "dirty_detached", provenance: SourceProvenance{Kind: SourceProvenanceGit, RepositoryRoot: "/work/repo", Head: "fedcba9876543210", Detached: true, Dirty: true}},
		{name: "unborn", provenance: SourceProvenance{Kind: SourceProvenanceGit, RepositoryRoot: "/work/repo", Branch: "main", Unborn: true, Dirty: true}},
		{name: "none_with_git_state", provenance: SourceProvenance{Kind: SourceProvenanceNone, Head: "abc"}, wantErr: true},
		{name: "git_missing_root", provenance: SourceProvenance{Kind: SourceProvenanceGit, Head: "abc", Branch: "main"}, wantErr: true},
		{name: "detached_with_branch", provenance: SourceProvenance{Kind: SourceProvenanceGit, RepositoryRoot: "/work/repo", Head: "abc", Branch: "main", Detached: true}, wantErr: true},
		{name: "unborn_with_head", provenance: SourceProvenance{Kind: SourceProvenanceGit, RepositoryRoot: "/work/repo", Head: "abc", Branch: "main", Unborn: true}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.provenance.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("invalid source provenance unexpectedly accepted")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("valid source provenance rejected: %v", err)
			}
			if tc.wantErr {
				return
			}
			encoded, err := json.Marshal(tc.provenance)
			if err != nil {
				t.Fatal(err)
			}
			var decoded SourceProvenance
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded != tc.provenance {
				t.Fatalf("source provenance round trip = %#v, want %#v", decoded, tc.provenance)
			}
		})
	}
}

func TestProjectTargetBindWireCarriesSourceProvenance(t *testing.T) {
	provenance := SourceProvenance{Kind: SourceProvenanceGit, RepositoryRoot: "/work/repo", Head: "abc123", Branch: "main", Dirty: true}
	request := ProjectTargetBindRequest{
		WorkSessionID: "ws-1", TargetID: "target-1", ProjectID: "project-1", DeploymentID: "deployment-1",
		CWDRel: ".", DeploymentRevision: "rev-1", ContextRevision: "sha256:ctx",
		PromptScopes: []PromptScopeRevision{{Scope: ".", PromptRevision: "sha256:prompt"}}, SourceProvenance: provenance,
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ProjectTargetBindRequest
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.SourceProvenance != provenance {
		t.Fatalf("bind source provenance = %#v, want %#v", decoded.SourceProvenance, provenance)
	}
}

func TestProjectContextDeliveryAndHostAckRoundTrip(t *testing.T) {
	if ProjectContextAckMetaKey != "io.nexusdock/project-context-ack" {
		t.Fatalf("ProjectContextAckMetaKey = %q", ProjectContextAckMetaKey)
	}
	delivery := ProjectContextDelivery{Status: ProjectContextReturned, ContextRevision: "sha256:context"}
	encoded, err := json.Marshal(delivery)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ProjectContextDelivery
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != delivery {
		t.Fatalf("delivery round trip = %#v, want %#v", decoded, delivery)
	}
	ack := ProjectContextAcknowledgment{WorkSessionID: "ws-1", TargetID: "target-1", ContextRevision: "sha256:context"}
	encoded, err = json.Marshal(ack)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" || !json.Valid(encoded) {
		t.Fatalf("ack did not encode as JSON: %q", encoded)
	}
}

func TestProjectPromptBudgetsAreOrderedAndBoundedBelowBridgeFrame(t *testing.T) {
	if MaxProjectPromptFileBytes != 64<<10 {
		t.Fatalf("file prompt budget = %d", MaxProjectPromptFileBytes)
	}
	if MaxProjectPromptTargetBytes != 256<<10 {
		t.Fatalf("target prompt budget = %d", MaxProjectPromptTargetBytes)
	}
	if MaxProjectPromptContextBytes != 1<<20 {
		t.Fatalf("context prompt budget = %d", MaxProjectPromptContextBytes)
	}
	if MaxProjectContextDeliveryBytes != 4<<20 {
		t.Fatalf("Project Context delivery budget = %d", MaxProjectContextDeliveryBytes)
	}
	if !(MaxProjectPromptFileBytes < MaxProjectPromptTargetBytes && MaxProjectPromptTargetBytes < MaxProjectPromptContextBytes && MaxProjectPromptContextBytes < MaxProjectContextDeliveryBytes) {
		t.Fatalf("Project Context budgets are not strictly ordered: file=%d target=%d context=%d delivery=%d", MaxProjectPromptFileBytes, MaxProjectPromptTargetBytes, MaxProjectPromptContextBytes, MaxProjectContextDeliveryBytes)
	}
}
