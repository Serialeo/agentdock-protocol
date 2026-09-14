package mcpcontract

import protocol "github.com/Serialeo/agentdock-protocol"

// ComputerToolNames describes Node-owned tools. It does not add tools to the
// canonical Nexus catalog or advertise a backend that has not been implemented.
func ComputerToolNames() []string {
	return []string{protocol.ToolComputerStatus, protocol.ToolComputerSession, protocol.ToolComputerObserve, protocol.ToolComputerAct, protocol.ToolComputerStop}
}

func computerID(description string) map[string]any {
	return map[string]any{"type": "string", "minLength": 1, "description": description}
}
func computerPoint() map[string]any {
	coordinate := func() map[string]any { return map[string]any{"type": "number", "minimum": 0} }
	return strictObject(map[string]any{"x": coordinate(), "y": coordinate()}, "x", "y")
}

func ComputerInputSchema(name string) (map[string]any, bool) {
	props := map[string]any{}
	var required []string
	switch name {
	case protocol.ToolComputerStatus:
		props["operation_id"] = computerID("Omit for minimal live status; specify to read only this owner's saved operation metadata.")
	case protocol.ToolComputerStop:
		props["session_id"] = computerID("Session owned by this execution context. Does not acquire or re-enable control.")
		props["operation_id"] = computerID("Optional operation within the owned session to stop.")
		required = []string{"session_id"}
	case protocol.ToolComputerSession:
		props["action"] = enumProperty("Acquire, renew or release control explicitly.", "acquire", "renew", "release")
		props["desktop_id"] = computerID("Desktop reported by the native backend; required on acquire.")
		props["session_id"] = computerID("Existing owned session; required on renew/release.")
		required = []string{"action"}
	case protocol.ToolComputerObserve:
		props["desktop_id"] = computerID("Desktop to observe without acquiring control.")
		props["display_id"] = computerID("One display tile to capture; no mixed-DPI stitching.")
		props["window_id"] = computerID("Optional native window identity to bind to this observation.")
		props["max_size"] = map[string]any{"type": "integer", "minimum": 1, "description": "Preferred maximum image edge; default 1600. Backend may reduce it to fit transport limits and reports actual image dimensions."}
		props["after_operation_id"] = computerID("Require an observation after this owner's operation; never replay input.")
		required = []string{"desktop_id", "display_id"}
	case protocol.ToolComputerAct:
		props["session_id"] = computerID("Current owned input lease.")
		props["operation_id"] = computerID("Stable idempotency key. Re-read the same ID on uncertain outcomes; never automatically issue a new ID.")
		props["observation_id"] = computerID("Server-saved observation supplying coordinates and target identity.")
		props["timeout_ms"] = map[string]any{"type": "integer", "minimum": 1, "description": "Requested total deadline in milliseconds; default 15000. Backend applies its execution budget."}
		props["action"] = computerActionSchema()
		required = []string{"session_id", "operation_id", "observation_id", "action"}
	default:
		return nil, false
	}
	schema := strictObject(props, required...)
	if name == protocol.ToolComputerSession {
		// Branches constrain the root fields without imposing nested routing fields.
		schema["oneOf"] = []any{
			map[string]any{"properties": map[string]any{"action": map[string]any{"const": "acquire"}}, "required": []string{"desktop_id"}, "not": map[string]any{"required": []string{"session_id"}}},
			map[string]any{"properties": map[string]any{"action": enumProperty("", "renew", "release")}, "required": []string{"session_id"}, "not": map[string]any{"required": []string{"desktop_id"}}},
		}
	}
	return schema, true
}

func computerActionSchema() map[string]any {
	variant := func(kind string, props map[string]any, fields ...string) map[string]any {
		props["kind"] = map[string]any{"const": kind, "type": "string"}
		return strictObject(props, append([]string{"kind"}, fields...)...)
	}
	button := func() map[string]any { return enumProperty("Mouse button; default left.", "left", "right", "middle") }
	return map[string]any{"type": "object", "oneOf": []any{
		variant("move", map[string]any{"point": computerPoint()}, "point"),
		variant("click", map[string]any{"point": computerPoint(), "button": button(), "count": boundedIntegerProperty("Click count; default 1. Includes triple-click selection.", 1, 3)}, "point"),
		variant("scroll", map[string]any{"point": computerPoint(), "delta_x": boundedIntegerProperty("Horizontal logical wheel units.", -1200, 1200), "delta_y": boundedIntegerProperty("Vertical logical wheel units.", -1200, 1200)}, "point", "delta_x", "delta_y"),
		variant("key", map[string]any{"keys": map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"type": "string", "minLength": 1}}}, "keys"),
		variant("text", map[string]any{"text": map[string]any{"type": "string", "minLength": 1}}, "text"),
		variant("drag", map[string]any{"point": computerPoint(), "to": computerPoint(), "button": button(), "duration_ms": boundedIntegerProperty("Bounded paired mouse down/move/up.", 1, 3000)}, "point", "to", "duration_ms"),
	}}
}

// ComputerOutputSchema is the common metadata contract. Image bytes travel in
// standard MCP image content, not public artifact URLs or filesystem paths.
func ComputerOutputSchema(name string) (map[string]any, bool) {
	if !protocol.IsComputerTool(name) {
		return nil, false
	}
	switch name {
	case protocol.ToolComputerObserve:
		return computerObservationSchema(), true
	case protocol.ToolComputerSession:
		return strictObject(map[string]any{
			"state":   enumProperty("Lease result.", "acquired", "renewed", "released"),
			"session": computerSessionSchema(),
		}, "state", "session"), true
	case protocol.ToolComputerStop:
		return strictObject(map[string]any{
			"session_id": computerID("Owned session."), "operation_id": computerID("Optional stopped operation."),
			"state": enumProperty("Cleanup result; not evidence that already-submitted input was undone.", "stopped", "already_stopped", "outcome_unknown"),
		}, "session_id", "state"), true
	case protocol.ToolComputerAct:
		return computerActionResultSchema(), true
	default:
		return map[string]any{"type": "object", "oneOf": []any{
			strictObject(map[string]any{
				"backend_state":         enumProperty("Native backend lifecycle.", "offline", "starting", "ready", "degraded", "unsupported"),
				"desktop_state":         enumProperty("Interactive desktop state.", "interactive", "locked", "disconnected", "unsupported", "unknown"),
				"local_control_enabled": booleanProperty("Actual native local opt-in."), "stop_latched": booleanProperty("Only the local user may clear this latch."),
				"permissions": strictObject(map[string]any{"observe": enumProperty("Actual helper capture permission.", "granted", "denied", "not_determined", "unavailable"), "control": enumProperty("Actual helper input permission.", "granted", "denied", "not_determined", "unavailable")}, "observe", "control"),
				"desktop_id":  computerID("Current desktop identity, when known."),
				"displays": map[string]any{"type": "array", "items": strictObject(map[string]any{
					"display_id": computerID("Display identity accepted by observe."),
					"width":      boundedIntegerProperty("Native pixel width.", 1, 32768),
					"height":     boundedIntegerProperty("Native pixel height.", 1, 32768),
					"rotation":   map[string]any{"type": "integer", "enum": []int{0, 90, 180, 270}},
				}, "display_id", "width", "height", "rotation")},
				"actions": arrayStrings("Implemented native action kinds; empty on unsupported backends."),
			}, "backend_state", "desktop_state", "local_control_enabled", "stop_latched", "permissions", "actions"),
			computerActionResultSchema(),
		}}, true
	}
}
func computerSessionSchema() map[string]any {
	return strictObject(map[string]any{"session_id": computerID("Lease ID."), "desktop_id": computerID("Desktop ID."), "desktop_epoch": computerID("Helper desktop generation."), "expires_at": stringProperty("Absolute expiry, RFC3339Nano.")}, "session_id", "desktop_id", "desktop_epoch", "expires_at")
}
func computerObservationSchema() map[string]any {
	return strictObject(map[string]any{
		"observation_id": computerID("Immutable observation ID."), "desktop_id": computerID("Desktop ID."), "display_id": computerID("Display ID."), "window_id": computerID("Optional bound target window."),
		"desktop_epoch": computerID("Desktop generation."), "geometry_revision": computerID("Display geometry revision."), "input_sequence": map[string]any{"type": "integer", "minimum": 0},
		"captured_at": stringProperty("Actual frame time, RFC3339Nano."), "expires_at": stringProperty("Observation expiry, RFC3339Nano."),
		"width": map[string]any{"type": "integer", "minimum": 1, "description": "Actual returned image width."}, "height": map[string]any{"type": "integer", "minimum": 1, "description": "Actual returned image height."}, "sha256": map[string]any{"type": "string", "pattern": `^[A-Fa-f0-9]{64}$`},
		"image_to_native": map[string]any{"type": "array", "minItems": 6, "maxItems": 6, "items": map[string]any{"type": "number"}},
	}, "observation_id", "desktop_id", "display_id", "desktop_epoch", "geometry_revision", "input_sequence", "captured_at", "expires_at", "width", "height", "sha256", "image_to_native")
}
func computerActionResultSchema() map[string]any {
	return strictObject(map[string]any{
		"operation_id": computerID("Saved operation ID."), "state": enumProperty("Durable execution state.", "prepared", "executing", "finished", "cancelled", "interrupted", "outcome_unknown"),
		"error_code":    stringProperty("Native failure code, when input was refused or incomplete."),
		"error_message": stringProperty("Native diagnostic, retained when rereading the operation."),
		"effects":       enumProperty("Known possible effects.", "none", "partial", "unknown", "applied"),
		"input":         strictObject(map[string]any{"status": enumProperty("API submission is not business verification.", "not_submitted", "submitted", "accepted", "rejected", "partial", "unknown"), "requested_events": map[string]any{"type": "integer", "minimum": 0}, "accepted_events": map[string]any{"type": "integer", "minimum": 0}}, "status"),
		"observation":   computerObservationSchema(), "observation_status": enumProperty("Post-input capture outcome.", "not_requested", "captured", "unavailable"),
		"verification": enumProperty("No business predicate is implemented in this contract version.", "not_checked"), "replayed_result": booleanProperty("Stored result was returned; input was not replayed."),
	}, "operation_id", "state", "effects", "input", "observation_status", "verification", "replayed_result")
}
