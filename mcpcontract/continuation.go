package mcpcontract

import protocol "github.com/Serialeo/agentdock-protocol"

const (
	ToolWorkContinuation          = "work_continuation"
	ToolPresentWorkContinuation   = "present_work_continuation"
	ToolConsumeWorkWake           = "consume_work_wake"
	ToolWorkContinuationBind      = "work_continuation_bind"
	ToolWorkContinuationState     = "work_continuation_state"
	ToolWorkContinuationHeartbeat = "work_continuation_heartbeat"
	ToolWorkContinuationPause     = "work_continuation_pause"
	ToolWorkWakeAcquire           = "work_wake_acquire"
	ToolWorkWakePrepare           = "work_wake_prepare"
	ToolWorkWakeFinish            = "work_wake_finish"
)

var continuationToolNames = []string{
	ToolWorkContinuation, ToolPresentWorkContinuation, ToolConsumeWorkWake,
	ToolWorkContinuationBind, ToolWorkContinuationState, ToolWorkContinuationHeartbeat,
	ToolWorkContinuationPause, ToolWorkWakeAcquire, ToolWorkWakePrepare, ToolWorkWakeFinish,
}

// ContinuationToolNames includes both model entrances and app-only coordination.
// These tools belong to Nexus; standalone AgentDock does not own a Wake authority.
func ContinuationToolNames() []string { return append([]string(nil), continuationToolNames...) }

// ToolVisibility is an execution exposure contract, not a rendering hint. Never
// remove it when stripping resourceUri or disabling Apps. App-only tools must be
// excluded entirely when their controller capability is disabled.
func ToolVisibility(name string) ([]string, bool) {
	switch name {
	case ToolWorkContinuation, ToolPresentWorkContinuation, ToolConsumeWorkWake:
		return []string{"model"}, true
	case ToolWorkContinuationBind, ToolWorkContinuationState, ToolWorkContinuationHeartbeat,
		ToolWorkContinuationPause, ToolWorkWakeAcquire, ToolWorkWakePrepare, ToolWorkWakeFinish:
		return []string{"app"}, true
	default:
		return nil, false
	}
}

// ToolMeta returns a fresh MCP metadata contract. Only explicit presentation
// creates a controller resource; neither regular work nor consume spawns a card.
func ToolMeta(name string) (map[string]any, bool) {
	visibility, ok := ToolVisibility(name)
	if !ok {
		return nil, false
	}
	ui := map[string]any{"visibility": visibility}
	if name == ToolPresentWorkContinuation {
		ui["resourceUri"] = protocol.WorkContinuationUIResourceURI
	}
	return map[string]any{"ui": ui}, true
}

func continuationAnnotations(name string) (Annotations, bool) {
	if _, ok := ToolVisibility(name); !ok {
		return Annotations{}, false
	}
	readOnly := name == ToolWorkContinuationState
	idempotent := readOnly || name == ToolWorkContinuationBind || name == ToolWorkContinuationHeartbeat || name == ToolWorkContinuationPause || name == ToolWorkWakeFinish
	return Annotations{
		ReadOnlyHint:    readOnly,
		DestructiveHint: boolPtr(false),
		IdempotentHint:  boolPtr(idempotent),
		OpenWorldHint:   boolPtr(false),
	}, true
}

func continuationID(description string) map[string]any {
	return map[string]any{"type": "string", "minLength": 1, "maxLength": 256, "description": description}
}

func continuationSourceSchema() map[string]any {
	return strictObject(map[string]any{
		"target_id":          continuationID("Existing authorized WorkSession Target; Nexus resolves node and execution context."),
		"command_session_id": continuationID("Exact command session to await, including one that has already completed."),
	}, "target_id", "command_session_id")
}

func continuationInputSchema(name string) (map[string]any, bool) {
	if _, ok := ToolVisibility(name); !ok {
		return nil, false
	}
	props := map[string]any{"work_session_id": continuationID("Explicit WorkSession owned by the authenticated client.")}
	required := []string{"work_session_id"}
	switch name {
	case ToolWorkContinuation:
		props["action"] = enumProperty("Enable only after explicit user consent; await hands off specific command results; settle records completed work, pause stops continuation, status inspects it; recover explicitly resolves inspected uncertain work or resets an exhausted budget.", "enable", "await", "settle", "pause", "status", "recover")
		props["confirmed"] = booleanProperty("Must be true for enable or recover, reflecting explicit user consent; recovery never silently repeats external work.")
		props["max_rounds"] = boundedIntegerProperty("Maximum automatic rounds before pausing.", 1, protocol.MaxWorkContinuationRounds)
		props["max_failures"] = boundedIntegerProperty("Failure circuit-breaker limit.", 1, protocol.MaxWorkContinuationFailures)
		props["sources"] = map[string]any{"type": "array", "minItems": 1, "maxItems": protocol.MaxWorkContinuationSources, "items": continuationSourceSchema()}
		props["wake_id"] = continuationID("Exact consumed Wake being settled, or unresolved Wake explicitly inspected by recover; required for recover when unresolved work exists.")
		props["checkpoint"] = map[string]any{"type": "string", "minLength": 1, "maxLength": 32768, "description": "Work completed and current recovery boundary. Required when recover resolves uncertain/unsettled work; no automatic re-execution is implied."}
		required = append(required, "action")
	case ToolPresentWorkContinuation:
		props["recover"] = booleanProperty("Explicitly replace an obsolete controller generation after inspecting uncertain or unsettled work. Never resends a prepared Wake.")
	case ToolConsumeWorkWake:
		props["protocol_version"] = map[string]any{"type": "integer", "const": protocol.WorkContinuationProtocolVersion}
		props["endpoint_id"] = continuationID("Exact endpoint from the server-generated resume envelope.")
		props["controller_generation"] = map[string]any{"type": "integer", "minimum": 1}
		props["wake_id"] = continuationID("Exact immutable Wake from the server-generated resume envelope.")
		props["attempt_id"] = continuationID("Exact prepared dispatch attempt from the resume envelope.")
		props["consume_token"] = map[string]any{"type": "string", "minLength": 1, "maxLength": 512, "description": "Single-use opaque credential from prepare; never synthesize, reuse, or substitute it."}
		required = append(required, "protocol_version", "endpoint_id", "controller_generation", "wake_id", "attempt_id", "consume_token")
	default:
		props["endpoint_id"] = continuationID("Endpoint created by explicit presentation.")
		props["controller_generation"] = map[string]any{"type": "integer", "minimum": 1}
		props["binding_id"] = continuationID("Random identifier for this controller View instance.")
		props["binding_secret"] = map[string]any{"type": "string", "minLength": 1, "maxLength": 512, "description": "Opaque presentation proof delivered only to the App in tool result metadata."}
		required = append(required, "endpoint_id", "controller_generation", "binding_id", "binding_secret")
		switch name {
		case ToolWorkContinuationBind:
			props["user_enabled"] = booleanProperty("Whether the user explicitly opted this visible View into dispatch. Does not independently enable the WorkSession.")
		case ToolWorkWakeAcquire:
			props["wake_id"] = continuationID("Optional exact pending Wake to claim; otherwise the server chooses eligible pending work.")
		case ToolWorkWakePrepare, ToolWorkWakeFinish:
			props["wake_id"] = continuationID("Exact Wake acquired by this binding.")
			props["attempt_id"] = continuationID("Exact acquired attempt; prepare establishes a durable no-resend fence.")
			required = append(required, "wake_id", "attempt_id")
			if name == ToolWorkWakeFinish {
				props["delivery_status"] = enumProperty("Host acceptance, explicit rejection including isError=true, or unknown delivery after timeout/disconnect. Late finish must not regress consumed state.", "dispatch_accepted", "delivery_rejected", "delivery_unknown")
				required = append(required, "delivery_status")
			}
		}
	}
	schema := strictObject(props, required...)
	if name == ToolWorkContinuation {
		schema["allOf"] = []any{
			map[string]any{"if": map[string]any{"properties": map[string]any{"action": map[string]any{"enum": []string{"enable", "recover"}}}}, "then": map[string]any{"required": []string{"confirmed"}, "properties": map[string]any{"confirmed": map[string]any{"const": true}}}},
			map[string]any{"if": map[string]any{"properties": map[string]any{"action": map[string]any{"const": "await"}}}, "then": map[string]any{"required": []string{"sources"}}},
			map[string]any{"if": map[string]any{"properties": map[string]any{"action": map[string]any{"const": "settle"}}}, "then": map[string]any{"required": []string{"wake_id", "checkpoint"}}},
		}
	}
	return schema, true
}

func continuationStateSchema() map[string]any {
	return strictObject(map[string]any{
		"work_session_id": continuationID("Bound WorkSession."),
		"phase":           stringProperty("Authoritative continuation phase; uncertain or consumed-unsettled work requires inspection."),
		"enabled":         booleanProperty("Whether the user enabled automatic continuation."),
		"max_rounds":      integerProperty("Automatic round budget."),
		"rounds_used":     integerProperty("Automatic rounds consumed."),
		"max_failures":    integerProperty("Failure circuit-breaker threshold."),
		"failure_count":   integerProperty("Failures counted toward the circuit breaker."),
		"sources":         map[string]any{"type": "array", "items": continuationSourceSchema()},
		"checkpoint":      stringProperty("Last explicitly settled checkpoint."),
		"last_error":      stringProperty("Reason continuation requires attention."),
		"updated_at":      stringProperty("RFC3339Nano state update time."),
	}, "work_session_id", "phase", "enabled", "max_rounds", "rounds_used", "max_failures", "failure_count", "updated_at")
}

func commandOutcomeSchema() map[string]any {
	executionProps := map[string]any{}
	for _, key := range []string{"work_session_id", "target_id", "project_id", "deployment_id"} {
		executionProps[key] = stringProperty("Execution-time identity resolved by Nexus from the authorized Target.")
	}
	props := map[string]any{
		"event_id":             continuationID("Stable execution event; deduplicated with authenticated node identity."),
		"command_session_id":   continuationID("Durable command session identity."),
		"execution_context":    strictObject(executionProps, "work_session_id", "target_id", "project_id", "deployment_id"),
		"state":                enumProperty("Execution fact; interrupted or outcome_unknown never instructs automatic rerun.", "starting", "running", "completed", "failed", "interrupted", "outcome_unknown"),
		"exit_code":            integerProperty("Known process exit code; omitted for unknown or nonterminal execution."),
		"output_truncated":     booleanProperty("Whether output snapshots were truncated."),
		"stdout_dropped_bytes": integerProperty("Stdout bytes omitted from the retained snapshot."),
		"stderr_dropped_bytes": integerProperty("Stderr bytes omitted from the retained snapshot."),
		"timed_out":            booleanProperty("Whether the execution timeout elapsed."),
		"pending_report":       booleanProperty("Whether the node still needs a durable Nexus acknowledgment."),
	}
	for _, key := range []string{"output", "output_ref", "stdout", "stderr", "workdir", "command_error", "started_at", "finished_at", "updated_at", "client_request_id"} {
		props[key] = stringProperty("Durable execution result metadata; output is replayable, not a live session cursor.")
	}
	return strictObject(props, "event_id", "command_session_id", "execution_context", "state", "output_truncated", "started_at", "updated_at", "pending_report")
}

func workWakeStateSchema() map[string]any {
	return enumProperty("Durable Wake or attempt state.", "pending", "claimed", "prepared", "dispatch_accepted", "delivery_rejected", "delivery_unknown", "consumed", "settled", "needs_attention")
}

func workWakeSchema() map[string]any {
	return strictObject(map[string]any{
		"work_session_id":       continuationID("Bound WorkSession."),
		"endpoint_id":           stringProperty("Dispatch endpoint, empty while pending before explicit presentation."),
		"controller_generation": map[string]any{"type": "integer", "minimum": 0, "description": "Zero until explicit presentation binds the pending Wake."},
		"wake_id":               continuationID("Immutable Wake identity."),
		"state":                 workWakeStateSchema(),
		"sources":               map[string]any{"type": "array", "items": commandOutcomeSchema()},
		"created_at":            stringProperty("RFC3339Nano creation time."),
		"updated_at":            stringProperty("RFC3339Nano update time."),
	}, "work_session_id", "endpoint_id", "controller_generation", "wake_id", "state", "sources", "created_at", "updated_at")
}

func workWakeAttemptSchema() map[string]any {
	props := map[string]any{
		"attempt_id": continuationID("Dispatch attempt identity."),
		"wake_id":    continuationID("Parent Wake."),
		"state":      workWakeStateSchema(),
	}
	for _, key := range []string{"lease_expires_at", "prepared_at", "finished_at", "consumed_at", "last_error"} {
		props[key] = stringProperty("Durable attempt metadata; plaintext consume credentials are never included.")
	}
	return strictObject(props, "attempt_id", "wake_id", "state")
}

func continuationOutputSchema(name string) (map[string]any, bool) {
	if _, ok := ToolVisibility(name); !ok {
		return nil, false
	}
	props := map[string]any{
		"work_session_id":       continuationID("Authorized WorkSession."),
		"endpoint_id":           continuationID("Explicit presentation endpoint."),
		"controller_generation": map[string]any{"type": "integer", "minimum": 1},
		"binding_id":            continuationID("Current View binding."),
		"lease_expires_at":      stringProperty("RFC3339Nano binding lease expiry."),
		"state":                 continuationStateSchema(),
		"wake_id":               continuationID("Current pending or claimed Wake, when available."),
		"attempt_id":            continuationID("Current dispatch attempt, when available."),
		"wake":                  workWakeSchema(),
		"attempt":               workWakeAttemptSchema(),
		"outcomes":              map[string]any{"type": "array", "items": commandOutcomeSchema()},
	}
	if name == ToolWorkWakePrepare {
		props["automatic_message"] = stringProperty("Server-generated exact consume envelope. Send verbatim once through ui/message; never replay after uncertain delivery.")
	}
	return strictObject(props, "work_session_id", "state"), true
}
