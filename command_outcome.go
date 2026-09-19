package protocol

// Command outcomes are an optional extension of Bridge v4. Nodes advertise this
// token only when durable reads and acknowledgments are both supported.
const (
	CommandOutcomesCapability     = "bridge.command.outcomes.v1"
	OperationCommandOutcomesRead  = "command.outcomes.read"
	OperationCommandOutcomesAck   = "command.outcomes.ack"
	DefaultCommandOutcomesPerRead = 32
	MaxCommandOutcomesPerRead     = 128
	MaxCommandOutcomeBatchBytes   = 1 << 20
	MaxCommandOutcomeOutputBytes  = 32 << 10
)

type CommandOutcomeState string

const (
	CommandOutcomeStarting    CommandOutcomeState = "starting"
	CommandOutcomeRunning     CommandOutcomeState = "running"
	CommandOutcomeCompleted   CommandOutcomeState = "completed"
	CommandOutcomeFailed      CommandOutcomeState = "failed"
	CommandOutcomeInterrupted CommandOutcomeState = "interrupted"
	CommandOutcomeUnknown     CommandOutcomeState = "outcome_unknown"
)

func (s CommandOutcomeState) Terminal() bool {
	switch s {
	case CommandOutcomeCompleted, CommandOutcomeFailed, CommandOutcomeInterrupted, CommandOutcomeUnknown:
		return true
	default:
		return false
	}
}

// CommandOutcome is a replayable execution fact, not an instruction to run a
// command again. Terminal state and PendingReport must be persisted atomically.
// Timestamps use RFC3339Nano. Output fields are bounded durable snapshots, never
// the mutable read cursor of a live command session. ExecutionContext may be zero
// for node-local records, but only Project-bound outcomes cross the Bridge.
// A receiver obtains node identity from the authenticated Bridge connection.
type CommandOutcome struct {
	EventID            string              `json:"event_id"`
	CommandSessionID   string              `json:"command_session_id"`
	ExecutionContext   ExecutionContext    `json:"execution_context"`
	State              CommandOutcomeState `json:"state"`
	ExitCode           *int                `json:"exit_code,omitempty"`
	Output             string              `json:"output,omitempty"`
	OutputRef          string              `json:"output_ref,omitempty"`
	OutputTruncated    bool                `json:"output_truncated"`
	Stdout             string              `json:"stdout,omitempty"`
	Stderr             string              `json:"stderr,omitempty"`
	StdoutDroppedBytes int64               `json:"stdout_dropped_bytes,omitempty"`
	StderrDroppedBytes int64               `json:"stderr_dropped_bytes,omitempty"`
	Workdir            string              `json:"workdir,omitempty"`
	CommandError       string              `json:"command_error,omitempty"`
	TimedOut           bool                `json:"timed_out,omitempty"`
	StartedAt          string              `json:"started_at"`
	FinishedAt         string              `json:"finished_at,omitempty"`
	UpdatedAt          string              `json:"updated_at"`
	PendingReport      bool                `json:"pending_report"`
	ClientRequestID    string              `json:"client_request_id,omitempty"`
}

// CommandOutcomesReadRequest reads a bounded snapshot. Explicit session IDs take
// precedence over PendingOnly and include already acknowledged outcomes, so an
// await registered after completion can still observe the execution fact.
// With no explicit IDs, PendingOnly selects unacknowledged terminal outcomes.
// Limit zero selects the service default, never above MaxCommandOutcomesPerRead.
type CommandOutcomesReadRequest struct {
	PendingOnly       bool     `json:"pending_only,omitempty"`
	CommandSessionIDs []string `json:"command_session_ids,omitempty"`
	Limit             int      `json:"limit,omitempty"`
}

type CommandOutcomesReadResult struct {
	Outcomes []CommandOutcome `json:"outcomes"`
	HasMore  bool             `json:"has_more"`
}

// Acknowledge only after the Nexus receipt and corresponding Wake state commit
// together. Repeated acknowledgments must be idempotent and never delete the
// replayable outcome merely because its pending-report marker is cleared.
type CommandOutcomesAckRequest struct {
	EventIDs []string `json:"event_ids"`
}

type CommandOutcomesAckResult struct {
	AcknowledgedEventIDs []string `json:"acknowledged_event_ids"`
}
