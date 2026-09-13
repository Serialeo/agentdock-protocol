package mcpcontract

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	protocol "github.com/Serialeo/agentdock-protocol"
)

func TestContinuationVisibilityAndPresentationCannotLeakToStandalone(t *testing.T) {
	for _, name := range ContinuationToolNames() {
		if slices.Contains(ToolNames(), name) || !slices.Contains(NexusToolNames(), name) || !IsCanonicalTool(name) {
			t.Fatalf("continuation tool has wrong authority: %s", name)
		}
		meta, ok := ToolMeta(name)
		if !ok {
			t.Fatalf("missing explicit visibility for %s", name)
		}
		ui := meta["ui"].(map[string]any)
		visibility := ui["visibility"].([]string)
		want := "app"
		if name == ToolWorkContinuation || name == ToolPresentWorkContinuation || name == ToolConsumeWorkWake {
			want = "model"
		}
		if !reflect.DeepEqual(visibility, []string{want}) {
			t.Fatalf("%s exposure = %v, want only %s", name, visibility, want)
		}
		uri, presents := ui["resourceUri"]
		if presents != (name == ToolPresentWorkContinuation) || presents && uri != protocol.WorkContinuationUIResourceURI {
			t.Fatalf("only explicit presentation may create a controller: %s %#v", name, ui)
		}
		visibility[0] = "tampered"
		fresh, _ := ToolVisibility(name)
		if fresh[0] != want {
			t.Fatal("visibility contract shares mutable memory")
		}
	}
	if meta, known := ToolMeta("unknown"); known || meta != nil {
		t.Fatal("unknown tools must not implicitly acquire a visibility contract")
	}
}

func TestContinuationInputsRejectRoutingOverridesAndBoundAwaitSources(t *testing.T) {
	for _, name := range ContinuationToolNames() {
		schema, ok := InputSchema(name)
		if !ok || schema["additionalProperties"] != false {
			t.Fatalf("%s input is not strict", name)
		}
		props := schema["properties"].(map[string]any)
		for _, forbidden := range []string{"node_id", "working_folder", "permissions", "execution_context", "command", "shell"} {
			if _, exists := props[forbidden]; exists {
				t.Fatalf("%s allows model/App execution override %q", name, forbidden)
			}
		}
	}
	schema, _ := InputSchema(ToolWorkContinuation)
	props := schema["properties"].(map[string]any)
	sources := props["sources"].(map[string]any)
	if sources["minItems"] != 1 || sources["maxItems"] != protocol.MaxWorkContinuationSources || sources["items"].(map[string]any)["additionalProperties"] != false {
		t.Fatalf("await sources lost their bounded exact-identity contract: %#v", sources)
	}
	conditions := schema["allOf"].([]any)
	if len(conditions) != 3 {
		t.Fatal("enable consent, await sources, and settled checkpoint need separate action requirements")
	}
	consent := conditions[0].(map[string]any)["then"].(map[string]any)
	if consent["properties"].(map[string]any)["confirmed"].(map[string]any)["const"] != true {
		t.Fatal("enable must require explicit confirmation")
	}
}

func TestContinuationControllerSchemasRequireEveryBindingIdentity(t *testing.T) {
	for _, name := range ContinuationToolNames() {
		visibility, _ := ToolVisibility(name)
		if visibility[0] != "app" {
			continue
		}
		schema, _ := InputSchema(name)
		required := schema["required"].([]string)
		for _, field := range []string{"work_session_id", "endpoint_id", "controller_generation", "binding_id", "binding_secret"} {
			if !slices.Contains(required, field) {
				t.Fatalf("%s lost required binding identity %q", name, field)
			}
		}
		if name == ToolWorkWakePrepare || name == ToolWorkWakeFinish {
			for _, field := range []string{"wake_id", "attempt_id"} {
				if !slices.Contains(required, field) {
					t.Fatalf("%s lost required attempt identity %q", name, field)
				}
			}
		}
	}
	consume, _ := InputSchema(ToolConsumeWorkWake)
	props := consume["properties"].(map[string]any)
	if len(props) != reflect.TypeFor[protocol.ResumeEnvelope]().NumField() || len(consume["required"].([]string)) != len(props) {
		t.Fatalf("consume must require the complete exact envelope: %#v", consume)
	}
}

func TestContinuationOutputNeverExposesBindingProofOrUnpreparedMessage(t *testing.T) {
	for _, name := range ContinuationToolNames() {
		schema, _ := OutputSchema(name)
		props := schema["properties"].(map[string]any)
		if _, exists := props["binding_secret"]; exists {
			t.Fatalf("%s exposes View proof to model-visible structuredContent", name)
		}
		_, message := props["automatic_message"]
		if message != (name == ToolWorkWakePrepare) {
			t.Fatalf("%s can reveal an unfenced or replayed automatic message", name)
		}
		first := props["state"].(map[string]any)
		first["additionalProperties"] = true
		fresh, _ := OutputSchema(name)
		if fresh["properties"].(map[string]any)["state"].(map[string]any)["additionalProperties"] != false {
			t.Fatal("output schema factories share nested mutable state")
		}
	}
}

func TestContinuationTypedResultsFitPublishedSchemas(t *testing.T) {
	zero := 0
	outcome := protocol.CommandOutcome{
		EventID: "event_1", CommandSessionID: "cmd_1", ExecutionContext: protocol.ExecutionContext{WorkSessionID: "ws_1", TargetID: "target_1", ProjectID: "project_1", DeploymentID: "dep_1", DeploymentRevision: "dr_1", ContextRevision: "cr_1"},
		State: protocol.CommandOutcomeCompleted, ExitCode: &zero, Stdout: "hello", Stderr: "warning", Output: "summary", OutputRef: "artifact_1", OutputTruncated: true,
		StdoutDroppedBytes: 3, StderrDroppedBytes: 4, Workdir: "/project", CommandError: "diagnostic", TimedOut: true,
		StartedAt: "start", FinishedAt: "finish", UpdatedAt: "update", PendingReport: true, ClientRequestID: "request_1", ArgumentsDigest: "digest",
	}
	result := protocol.WorkContinuationResult{
		WorkSessionID: "ws_1", EndpointID: "ep_1", ControllerGeneration: 1, BindingID: "binding_1", LeaseExpiresAt: "lease", WakeID: "wake_1", AttemptID: "attempt_1",
		State:    protocol.WorkContinuation{WorkSessionID: "ws_1", Phase: "awaiting", Enabled: true, MaxRounds: 3, RoundsUsed: 1, MaxFailures: 2, FailureCount: 1, Sources: []protocol.ContinuationSource{{TargetID: "target_1", CommandSessionID: "cmd_1"}}, Checkpoint: "checked", LastError: "diagnostic", UpdatedAt: "update"},
		Wake:     &protocol.WorkWake{WorkSessionID: "ws_1", EndpointID: "ep_1", ControllerGeneration: 1, WakeID: "wake_1", State: protocol.WorkWakePrepared, Sources: []protocol.CommandOutcome{outcome}, CreatedAt: "create", UpdatedAt: "update"},
		Attempt:  &protocol.WorkWakeAttempt{AttemptID: "attempt_1", WakeID: "wake_1", State: protocol.WorkWakePrepared, LeaseExpiresAt: "lease", PreparedAt: "prepare", FinishedAt: "finish", ConsumedAt: "consume", LastError: "diagnostic"},
		Outcomes: []protocol.CommandOutcome{outcome},
	}
	for _, name := range ContinuationToolNames() {
		current := result
		if name == ToolWorkWakePrepare {
			current.AutomaticMessage = "exact consume envelope"
		}
		data, err := json.Marshal(current)
		if err != nil {
			t.Fatal(err)
		}
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatal(err)
		}
		schema, _ := OutputSchema(name)
		assertDeclaredShape(t, name, schema, value)
	}
}

// Guard against cross-repository struct/schema drift, recursively checking the
// actual JSON payload including required fields and strict object boundaries.
func assertDeclaredShape(t *testing.T, path string, schema map[string]any, value any) {
	t.Helper()
	switch schema["type"] {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("%s is not an object: %#v", path, value)
		}
		properties := schema["properties"].(map[string]any)
		if required, ok := schema["required"].([]string); ok {
			for _, key := range required {
				if _, exists := object[key]; !exists {
					t.Fatalf("%s missing required field %s", path, key)
				}
			}
		}
		for key, item := range object {
			property, declared := properties[key].(map[string]any)
			if !declared {
				t.Fatalf("%s has undeclared field %s", path, key)
			}
			assertDeclaredShape(t, path+"."+key, property, item)
		}
	case "array":
		items, ok := value.([]any)
		if !ok {
			t.Fatalf("%s is not an array: %#v", path, value)
		}
		for _, item := range items {
			assertDeclaredShape(t, path+"[]", schema["items"].(map[string]any), item)
		}
	case "string":
		if _, ok := value.(string); !ok {
			t.Fatalf("%s is not a string", path)
		}
	case "integer":
		if number, ok := value.(float64); !ok || number != float64(int64(number)) {
			t.Fatalf("%s is not an integer", path)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			t.Fatalf("%s is not a boolean", path)
		}
	}
}
