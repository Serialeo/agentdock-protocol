package protocol

import (
	"encoding/json"
	"errors"
	"strings"
)

const (
	WorkContinuationProtocolVersion = 1
	WorkContinuationMetaKey         = "agentdock/work-continuation"
	MaxWorkContinuationSources      = 32
	MaxWorkContinuationRounds       = 100
	MaxWorkContinuationFailures     = 10
)

// ContinuationSource identifies an explicitly awaited command. Node and cwd are
// resolved from the authenticated WorkSession Target by Nexus, never from input.
type ContinuationSource struct {
	TargetID         string `json:"target_id"`
	CommandSessionID string `json:"command_session_id"`
}

type WorkContinuation struct {
	WorkSessionID string               `json:"work_session_id"`
	Phase         string               `json:"phase"`
	Enabled       bool                 `json:"enabled"`
	MaxRounds     int                  `json:"max_rounds"`
	RoundsUsed    int                  `json:"rounds_used"`
	MaxFailures   int                  `json:"max_failures"`
	FailureCount  int                  `json:"failure_count"`
	Sources       []ContinuationSource `json:"sources,omitempty"`
	Checkpoint    string               `json:"checkpoint,omitempty"`
	LastError     string               `json:"last_error,omitempty"`
	UpdatedAt     string               `json:"updated_at"`
}

type WorkContinuationEndpoint struct {
	WorkSessionID        string `json:"work_session_id"`
	EndpointID           string `json:"endpoint_id"`
	ControllerGeneration int64  `json:"controller_generation"`
	BindingID            string `json:"binding_id,omitempty"`
	LeaseExpiresAt       string `json:"lease_expires_at,omitempty"`
}

type WorkWakeState string

const (
	WorkWakePending          WorkWakeState = "pending"
	WorkWakeClaimed          WorkWakeState = "claimed"
	WorkWakePrepared         WorkWakeState = "prepared"
	WorkWakeDispatchAccepted WorkWakeState = "dispatch_accepted"
	WorkWakeDeliveryRejected WorkWakeState = "delivery_rejected"
	WorkWakeDeliveryUnknown  WorkWakeState = "delivery_unknown"
	WorkWakeConsumed         WorkWakeState = "consumed"
	WorkWakeSettled          WorkWakeState = "settled"
	WorkWakeNeedsAttention   WorkWakeState = "needs_attention"
)

// A prepared Wake's Sources are immutable. Later facts belong to a later Wake.
type WorkWake struct {
	WorkSessionID        string           `json:"work_session_id"`
	EndpointID           string           `json:"endpoint_id"`
	ControllerGeneration int64            `json:"controller_generation"`
	WakeID               string           `json:"wake_id"`
	State                WorkWakeState    `json:"state"`
	Sources              []CommandOutcome `json:"sources"`
	CreatedAt            string           `json:"created_at"`
	UpdatedAt            string           `json:"updated_at"`
}

type WorkWakeAttempt struct {
	AttemptID      string        `json:"attempt_id"`
	WakeID         string        `json:"wake_id"`
	State          WorkWakeState `json:"state"`
	LeaseExpiresAt string        `json:"lease_expires_at,omitempty"`
	PreparedAt     string        `json:"prepared_at,omitempty"`
	FinishedAt     string        `json:"finished_at,omitempty"`
	ConsumedAt     string        `json:"consumed_at,omitempty"`
	LastError      string        `json:"last_error,omitempty"`
}

// ResumeEnvelope contains only exact routing identity and a single-use consume
// credential. It must not acquire executable commands, cwd, node overrides, or
// arbitrary model instructions. ConsumeToken is returned only once by prepare;
// persistent attempt records retain its digest instead of the plaintext token.
type ResumeEnvelope struct {
	ProtocolVersion      int    `json:"protocol_version"`
	WorkSessionID        string `json:"work_session_id"`
	EndpointID           string `json:"endpoint_id"`
	ControllerGeneration int64  `json:"controller_generation"`
	WakeID               string `json:"wake_id"`
	AttemptID            string `json:"attempt_id"`
	ConsumeToken         string `json:"consume_token"`
}

func (e ResumeEnvelope) Validate() error {
	if e.ProtocolVersion != WorkContinuationProtocolVersion || e.ControllerGeneration < 1 {
		return errors.New("invalid continuation protocol version or generation")
	}
	for _, value := range []string{e.WorkSessionID, e.EndpointID, e.WakeID, e.AttemptID, e.ConsumeToken} {
		if strings.TrimSpace(value) == "" {
			return errors.New("continuation envelope requires every exact identity and consume token")
		}
	}
	return nil
}

// AutomaticMessage builds the only server-generated automatic continuation
// message. The controller sends it verbatim; all actual work is obtained from
// consume_work_wake after server-side authorization and token consumption.
func (e ResumeEnvelope) AutomaticMessage() (string, error) {
	if err := e.Validate(); err != nil {
		return "", err
	}
	body, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return "Call consume_work_wake with exactly this continuation envelope: " + string(body), nil
}

type WorkContinuationInput struct {
	Action        string               `json:"action"`
	WorkSessionID string               `json:"work_session_id"`
	Confirmed     bool                 `json:"confirmed,omitempty"`
	MaxRounds     int                  `json:"max_rounds,omitempty"`
	MaxFailures   int                  `json:"max_failures,omitempty"`
	Sources       []ContinuationSource `json:"sources,omitempty"`
	WakeID        string               `json:"wake_id,omitempty"`
	Checkpoint    string               `json:"checkpoint,omitempty"`
}

type PresentWorkContinuationInput struct {
	WorkSessionID string `json:"work_session_id"`
	Recover       bool   `json:"recover,omitempty"`
}

// BindingSecret is presentation proof, not a substitute for authentication.
// Each operation must still authorize owner, WorkSession, Target, Endpoint,
// generation, active binding, and the operation's Wake/attempt where relevant.
type ContinuationControllerInput struct {
	WorkSessionID        string `json:"work_session_id"`
	EndpointID           string `json:"endpoint_id"`
	ControllerGeneration int64  `json:"controller_generation"`
	BindingID            string `json:"binding_id"`
	BindingSecret        string `json:"binding_secret"`
	UserEnabled          bool   `json:"user_enabled,omitempty"`
	WakeID               string `json:"wake_id,omitempty"`
	AttemptID            string `json:"attempt_id,omitempty"`
	DeliveryStatus       string `json:"delivery_status,omitempty"`
}

type ConsumeWorkWakeInput = ResumeEnvelope

// WorkContinuationResult is the common controller/model response envelope.
// BindingSecret is set only for explicit presentation and must be moved into
// CallToolResult._meta[WorkContinuationMetaKey] by the MCP adapter, excluded from
// model-visible structuredContent. AutomaticMessage is only returned by prepare.
type WorkContinuationResult struct {
	WorkSessionID        string           `json:"work_session_id"`
	EndpointID           string           `json:"endpoint_id,omitempty"`
	ControllerGeneration int64            `json:"controller_generation,omitempty"`
	BindingID            string           `json:"binding_id,omitempty"`
	BindingSecret        string           `json:"binding_secret,omitempty"`
	LeaseExpiresAt       string           `json:"lease_expires_at,omitempty"`
	State                WorkContinuation `json:"state"`
	WakeID               string           `json:"wake_id,omitempty"`
	AttemptID            string           `json:"attempt_id,omitempty"`
	AutomaticMessage     string           `json:"automatic_message,omitempty"`
	Wake                 *WorkWake        `json:"wake,omitempty"`
	Attempt              *WorkWakeAttempt `json:"attempt,omitempty"`
	Outcomes             []CommandOutcome `json:"outcomes,omitempty"`
}
