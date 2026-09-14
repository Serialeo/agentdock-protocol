package protocol

import "strings"

// ComputerPermission is explicit in every deployment snapshot; it is independent of shell/browser access.
type ComputerPermission string

const (
	ComputerPermissionNone    ComputerPermission = "none"
	ComputerPermissionObserve ComputerPermission = "observe"
	ComputerPermissionControl ComputerPermission = "control"
	// ComputerCapability is advertised only by Nodes with the complete native
	// session/observe/act/stop/result contract, never by the P0 calibration app.
	ComputerCapability  = "bridge.computer.v1"
	ComputerWireVersion = 1
)

const (
	ToolComputerStatus  = "computer_status"
	ToolComputerSession = "computer_session"
	ToolComputerObserve = "computer_observe"
	ToolComputerAct     = "computer_act"
	ToolComputerStop    = "computer_stop"
)

// IsComputerTool recognizes reserved native tool names, including on older or
// third-party Nodes. Unknown computer_* names are not implicitly authorized.
func IsComputerTool(name string) bool {
	switch name {
	case ToolComputerStatus, ToolComputerSession, ToolComputerObserve, ToolComputerAct, ToolComputerStop:
		return true
	default:
		return false
	}
}

// AllowsComputerTool checks Deployment policy only. Runtime must additionally
// check native capability, execution owner, OS permissions and local stop state.
// A stopped owner's cleanup and persisted-result query never authorize live capture.
func (p DeploymentPermissions) AllowsComputerTool(name string) bool {
	if !IsComputerTool(name) {
		return false
	}
	if name == ToolComputerStatus || name == ToolComputerStop {
		return true
	}
	if p.Validate() != nil {
		return false
	}
	if p.FullAccess {
		return true
	}
	switch p.Computer {
	case ComputerPermissionControl:
		return true
	case ComputerPermissionObserve:
		return name == ToolComputerObserve
	default:
		return false
	}
}

// AllowsHistoricalComputerControl classifies only cleanup and persisted metadata.
// Ownership, operation lookup and response redaction remain mandatory at execution.
func AllowsHistoricalComputerControl(name string, args map[string]any) bool {
	if name == ToolComputerStop {
		return true
	}
	if name != ToolComputerStatus {
		return false
	}
	id, ok := args["operation_id"].(string)
	return ok && strings.TrimSpace(id) != ""
}

const (
	ErrorComputerUnsupported        = "COMPUTER_UNSUPPORTED"
	ErrorComputerContractMismatch   = "COMPUTER_CONTRACT_MISMATCH"
	ErrorComputerHelperOffline      = "COMPUTER_HELPER_OFFLINE"
	ErrorComputerPermissionDenied   = "COMPUTER_PERMISSION_DENIED"
	ErrorComputerDesktopUnavailable = "COMPUTER_DESKTOP_UNAVAILABLE"
	ErrorComputerDesktopBusy        = "COMPUTER_DESKTOP_BUSY"
	ErrorComputerObservationStale   = "COMPUTER_OBSERVATION_STALE"
	ErrorComputerSessionRevoked     = "COMPUTER_SESSION_REVOKED"
	ErrorComputerOperationConflict  = "COMPUTER_OPERATION_CONFLICT"
	ErrorComputerInputRejected      = "COMPUTER_INPUT_REJECTED"
	ErrorComputerExecutionUnknown   = "COMPUTER_EXECUTION_UNKNOWN"
	ErrorComputerEvidenceExpired    = "COMPUTER_EVIDENCE_EXPIRED"
)

// ComputerPoint always uses pixels in the referenced observation, never global OS coordinates.
type ComputerPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ComputerAction is a wire union validated by the shared input schema. Fields are
// optional here to avoid imposing a platform-specific backend on Go consumers.
type ComputerAction struct {
	Kind       string         `json:"kind"`
	Point      *ComputerPoint `json:"point,omitempty"`
	To         *ComputerPoint `json:"to,omitempty"`
	Button     string         `json:"button,omitempty"`
	Count      int            `json:"count,omitempty"`
	DeltaX     *int           `json:"delta_x,omitempty"`
	DeltaY     *int           `json:"delta_y,omitempty"`
	Keys       []string       `json:"keys,omitempty"`
	Text       string         `json:"text,omitempty"`
	DurationMS int            `json:"duration_ms,omitempty"`
}

type ComputerSession struct {
	SessionID    string `json:"session_id"`
	DesktopID    string `json:"desktop_id"`
	DesktopEpoch string `json:"desktop_epoch"`
	ExpiresAt    string `json:"expires_at"`
}

type ComputerObservation struct {
	ObservationID    string `json:"observation_id"`
	DesktopID        string `json:"desktop_id"`
	DisplayID        string `json:"display_id"`
	WindowID         string `json:"window_id,omitempty"`
	DesktopEpoch     string `json:"desktop_epoch"`
	GeometryRevision string `json:"geometry_revision"`
	InputSequence    uint64 `json:"input_sequence"`
	CapturedAt       string `json:"captured_at"`
	ExpiresAt        string `json:"expires_at"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	SHA256           string `json:"sha256"`
	// The server persists this transform; act never accepts one supplied by the model.
	ImageToNative [6]float64 `json:"image_to_native"`
}

type ComputerInputResult struct {
	Status          string `json:"status"` // not_submitted, submitted, accepted, rejected, partial, unknown
	RequestedEvents *int   `json:"requested_events,omitempty"`
	AcceptedEvents  *int   `json:"accepted_events,omitempty"` // CGEventPost cannot provide this count.
}

type ComputerActionResult struct {
	ErrorCode         string               `json:"error_code,omitempty"`
	ErrorMessage      string               `json:"error_message,omitempty"`
	OperationID       string               `json:"operation_id"`
	State             string               `json:"state"`
	Effects           string               `json:"effects"`
	Input             ComputerInputResult  `json:"input"`
	Observation       *ComputerObservation `json:"observation,omitempty"`
	ObservationStatus string               `json:"observation_status"`
	Verification      string               `json:"verification"` // not_checked unless a trusted predicate ran.
	ReplayedResult    bool                 `json:"replayed_result"`
}

// ComputerBackendStatus contains no window titles, screenshot bytes or other lease owners.
// Display IDs are the discovery inputs for computer_observe.
type ComputerBackendStatus struct {
	BackendState        string                    `json:"backend_state"`
	DesktopState        string                    `json:"desktop_state"`
	DesktopID           string                    `json:"desktop_id,omitempty"`
	Displays            []ComputerDisplay         `json:"displays,omitempty"`
	LocalControlEnabled bool                      `json:"local_control_enabled"`
	StopLatched         bool                      `json:"stop_latched"`
	Permissions         ComputerNativePermissions `json:"permissions"`
	Actions             []string                  `json:"actions"`
}
type ComputerNativePermissions struct {
	Observe string `json:"observe"`
	Control string `json:"control"`
}
type ComputerDisplay struct {
	DisplayID string `json:"display_id"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Rotation  int    `json:"rotation"`
}
