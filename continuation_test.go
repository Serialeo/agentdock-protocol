package protocol

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCommandOutcomesExtendBridgeV4WithoutRequiringCapability(t *testing.T) {
	if ConnectionProtocolVersion != "4" || CommandOutcomesCapability != "bridge.command.outcomes.v1" || OperationCommandOutcomesRead != "command.outcomes.read" || OperationCommandOutcomesAck != "command.outcomes.ack" {
		t.Fatal("command outcomes changed the frozen optional Bridge v4 contract")
	}
	var old Hello
	if err := json.Unmarshal([]byte(`{"protocol_version":"4","tools":[],"ui_resources":[]}`), &old); err != nil || len(old.BridgeCapabilities) != 0 {
		t.Fatalf("legacy v4 Hello must remain valid without outcome support: %#v, %v", old, err)
	}
	if OperationRequiresExecutionContext(OperationCommandOutcomesRead) || OperationRequiresExecutionContext(OperationCommandOutcomesAck) {
		t.Fatal("peer-authorized batched outcome reconciliation cannot require one tool invocation context")
	}
	if DefaultCommandOutcomesPerRead > MaxCommandOutcomesPerRead || MaxCommandOutcomeBatchBytes >= 8<<20 {
		t.Fatal("outcome bounds must fit below the Bridge message limit")
	}
}

func TestCommandOutcomeRoundTripRetainsExecutionIdentityAndZeroExit(t *testing.T) {
	zero := 0
	original := CommandOutcome{
		EventID: "event_1", CommandSessionID: "command_1",
		ExecutionContext: ExecutionContext{WorkSessionID: "ws_1", TargetID: "target_1", ProjectID: "project_1", DeploymentID: "deployment_1", DeploymentRevision: "dep_rev_2", ContextRevision: "ctx_rev_3"},
		State:            CommandOutcomeCompleted, ExitCode: &zero, Stdout: "durable output", Stderr: "warning", StdoutDroppedBytes: 10,
		StartedAt: "2026-09-13T10:00:00Z", FinishedAt: "2026-09-13T10:00:01Z", UpdatedAt: "2026-09-13T10:00:01Z", PendingReport: true,
		ClientRequestID: "request_1",
	}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded CommandOutcome
	if err := json.Unmarshal(encoded, &decoded); err != nil || !reflect.DeepEqual(original, decoded) {
		t.Fatalf("outcome wire round trip lost replay identity/result: %#v, %v", decoded, err)
	}
	if !strings.Contains(string(encoded), `"exit_code":0`) {
		t.Fatal("known successful exit must not be omitted")
	}
	original.ExitCode = nil
	original.State = CommandOutcomeUnknown
	encoded, err = json.Marshal(original)
	if err != nil || strings.Contains(string(encoded), `"exit_code"`) {
		t.Fatalf("unknown outcome must not masquerade as exit zero: %s, %v", encoded, err)
	}
}

func TestCommandOutcomeTerminalClassificationIncludesRecoveryUncertainty(t *testing.T) {
	for _, state := range []CommandOutcomeState{CommandOutcomeCompleted, CommandOutcomeFailed, CommandOutcomeInterrupted, CommandOutcomeUnknown} {
		if !state.Terminal() {
			t.Fatalf("replayable terminal outcome %q was excluded", state)
		}
	}
	for _, state := range []CommandOutcomeState{CommandOutcomeStarting, CommandOutcomeRunning, "", "future-state"} {
		if state.Terminal() {
			t.Fatalf("unrecognized or active state %q is not a completion fact", state)
		}
	}
}

func validResumeEnvelope() ResumeEnvelope {
	return ResumeEnvelope{ProtocolVersion: 1, WorkSessionID: "ws_1", EndpointID: "ep_1", ControllerGeneration: 7, WakeID: "wake_1", AttemptID: "attempt_1", ConsumeToken: "opaque-single-use-token"}
}

func TestAutomaticMessageContainsOnlyExactConsumeIdentity(t *testing.T) {
	envelope := validResumeEnvelope()
	// Quoted data remains within JSON; it cannot append another instruction.
	envelope.WorkSessionID = "ws_quoted\"\\\nvalue"
	message, err := envelope.AutomaticMessage()
	if err != nil {
		t.Fatal(err)
	}
	const prefix = "Call consume_work_wake with exactly this continuation envelope: "
	if !strings.HasPrefix(message, prefix) {
		t.Fatalf("message has an unexpected instruction: %q", message)
	}
	var recovered ResumeEnvelope
	if err := json.Unmarshal([]byte(strings.TrimPrefix(message, prefix)), &recovered); err != nil || recovered != envelope {
		t.Fatalf("automatic message changed the authoritative envelope: %#v, %v", recovered, err)
	}
	var fields map[string]any
	if err := json.Unmarshal([]byte(strings.TrimPrefix(message, prefix)), &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 7 {
		t.Fatalf("resume envelope must contain exactly seven identity fields: %#v", fields)
	}
	for _, forbidden := range []string{"command", "shell", "working_folder", "cwd", "node_id", "task", "instructions", "permissions"} {
		if _, exists := fields[forbidden]; exists {
			t.Fatalf("automatic message acquired execution payload %q", forbidden)
		}
	}
}

func TestResumeEnvelopeRejectsIncompleteOrStaleIdentity(t *testing.T) {
	mutations := []func(*ResumeEnvelope){
		func(e *ResumeEnvelope) { e.ProtocolVersion = 0 },
		func(e *ResumeEnvelope) { e.ProtocolVersion = 2 },
		func(e *ResumeEnvelope) { e.ControllerGeneration = 0 },
		func(e *ResumeEnvelope) { e.ControllerGeneration = -1 },
		func(e *ResumeEnvelope) { e.WorkSessionID = "" },
		func(e *ResumeEnvelope) { e.EndpointID = " " },
		func(e *ResumeEnvelope) { e.WakeID = "" },
		func(e *ResumeEnvelope) { e.AttemptID = "" },
		func(e *ResumeEnvelope) { e.ConsumeToken = "\t" },
	}
	for index, mutate := range mutations {
		envelope := validResumeEnvelope()
		mutate(&envelope)
		if message, err := envelope.AutomaticMessage(); err == nil || message != "" {
			t.Fatalf("invalid envelope %d produced a dispatchable message: %q, %v", index, message, err)
		}
	}
}
