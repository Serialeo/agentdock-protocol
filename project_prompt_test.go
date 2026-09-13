package protocol

import (
	"encoding/json"
	"testing"
)

func TestProjectPromptBridgeOperationLiteralAndRoundTrip(t *testing.T) {
	if OperationProjectPromptLoad != "project.prompt.load" {
		t.Fatalf("OperationProjectPromptLoad = %q", OperationProjectPromptLoad)
	}
	request := ProjectPromptLoadRequest{DeploymentID: "deployment-1", DeploymentRevision: "rev-2", CWDRel: "backend/auth"}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ProjectPromptLoadRequest
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != request {
		t.Fatalf("ProjectPromptLoadRequest round trip = %#v, want %#v", decoded, request)
	}

	result := ProjectPromptLoadResult{
		DeploymentID: "deployment-1",
		CWDRel:       "backend/auth",
		Prompt: ProjectPrompt{
			PromptRevision: "sha256:prompt",
			Complete:       true,
			Bytes:          11,
			Sources:        []PromptSource{{Path: "AGENTS.md", Scope: ".", SHA256: "sha256:root", Bytes: 11, Content: "root rules\n"}},
		},
	}
	encoded, err = json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decodedResult ProjectPromptLoadResult
	if err := json.Unmarshal(encoded, &decodedResult); err != nil {
		t.Fatal(err)
	}
	if decodedResult.DeploymentID != result.DeploymentID || decodedResult.CWDRel != result.CWDRel || decodedResult.Prompt.PromptRevision != result.Prompt.PromptRevision || len(decodedResult.Prompt.Sources) != 1 || decodedResult.Prompt.Sources[0].Content != "root rules\n" {
		t.Fatalf("ProjectPromptLoadResult round trip = %#v", decodedResult)
	}
}

func TestTargetPromptScopesSurviveBindAndRebindWireRoundTrip(t *testing.T) {
	scopes := []PromptScopeRevision{
		{Scope: ".", PromptRevision: "sha256:root"},
		{Scope: "backend", PromptRevision: "sha256:backend"},
	}
	bind := ProjectTargetBindRequest{
		WorkSessionID: "ws-1", TargetID: "target-1", ProjectID: "project-1", DeploymentID: "deployment-1",
		CWDRel: "backend", DeploymentRevision: "rev-1", ContextRevision: "ctx-1", PromptScopes: scopes,
	}
	encoded, err := json.Marshal(bind)
	if err != nil {
		t.Fatal(err)
	}
	var decodedBind ProjectTargetBindRequest
	if err := json.Unmarshal(encoded, &decodedBind); err != nil {
		t.Fatal(err)
	}
	if len(decodedBind.PromptScopes) != 2 || decodedBind.PromptScopes[1] != scopes[1] {
		t.Fatalf("bind prompt_scopes = %#v", decodedBind.PromptScopes)
	}

	rebind := ProjectTargetRebindRequest{WorkSessionID: "ws-1", TargetID: "target-1", CWDRel: "backend/auth", ContextRevision: "ctx-2", PromptScopes: scopes}
	encoded, err = json.Marshal(rebind)
	if err != nil {
		t.Fatal(err)
	}
	var decodedRebind ProjectTargetRebindRequest
	if err := json.Unmarshal(encoded, &decodedRebind); err != nil {
		t.Fatal(err)
	}
	if len(decodedRebind.PromptScopes) != 2 || decodedRebind.PromptScopes[0] != scopes[0] {
		t.Fatalf("rebind prompt_scopes = %#v", decodedRebind.PromptScopes)
	}
}
