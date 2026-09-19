package protocol

import (
	"errors"
	"fmt"
	"strings"
)

type FileCapability string

const (
	FileCapabilityNone      FileCapability = "none"
	FileCapabilityReadOnly  FileCapability = "read_only"
	FileCapabilityReadWrite FileCapability = "read_write"
)

type DeploymentApplyStatus string

const (
	DeploymentApplyDraft    DeploymentApplyStatus = "draft"
	DeploymentApplyPending  DeploymentApplyStatus = "pending"
	DeploymentApplyApplied  DeploymentApplyStatus = "applied"
	DeploymentApplyFailed   DeploymentApplyStatus = "failed"
	DeploymentApplyDisabled DeploymentApplyStatus = "disabled"
)

type WorkSessionStatus string

const (
	WorkSessionPreparing WorkSessionStatus = "preparing"
	WorkSessionReady     WorkSessionStatus = "ready"
	WorkSessionRunning   WorkSessionStatus = "running"
	WorkSessionCompleted WorkSessionStatus = "completed"
	WorkSessionPartial   WorkSessionStatus = "partial"
	WorkSessionFailed    WorkSessionStatus = "failed"
	WorkSessionCancelled WorkSessionStatus = "cancelled"
)

type TargetStatus string

const (
	TargetPreparing    TargetStatus = "preparing"
	TargetReady        TargetStatus = "ready"
	TargetRunning      TargetStatus = "running"
	TargetIdle         TargetStatus = "idle"
	TargetUnavailable  TargetStatus = "unavailable"
	TargetContextError TargetStatus = "context_error"
	TargetRevoked      TargetStatus = "revoked"
)

type ProjectContextDeliveryStatus string

const (
	ProjectContextReturned     ProjectContextDeliveryStatus = "returned"
	ProjectContextHostConsumed ProjectContextDeliveryStatus = "host_consumed"

	// ProjectContextAckMetaKey is reserved for explicit MCP Host acknowledgments.
	// It belongs in CallTool request _meta, never in model-controlled tool arguments.
	ProjectContextAckMetaKey = "io.nexusdock/project-context-ack"
	// ProjectContextDeliveryMetaKey carries the exact delivery identity back to the
	// MCP Host without exposing internal revisions in model-visible tool content.
	ProjectContextDeliveryMetaKey = "io.nexusdock/project-context-delivery"
)

type ProjectContextDelivery struct {
	Status          ProjectContextDeliveryStatus `json:"status"`
	ContextRevision string                       `json:"context_revision"`
}

type ProjectContextAcknowledgment struct {
	WorkSessionID   string `json:"work_session_id"`
	TargetID        string `json:"target_id,omitempty"`
	ContextRevision string `json:"context_revision"`
}

type DeploymentPermissions struct {
	// FullAccess grants all Project execution capabilities exposed by the target Node.
	// It is intentionally independent from Deployment.WorkingFolder: the folder is a
	// default cwd / Project Prompt anchor, not an OS access boundary.
	FullAccess bool           `json:"full_access"`
	Files      FileCapability `json:"files"`
	Shell      bool           `json:"shell"`
	Browser    bool           `json:"browser"`
	DynamicMCP bool           `json:"dynamic_mcp"`
	ACP        bool           `json:"acp"`
}

func (p DeploymentPermissions) Validate() error {
	switch p.Files {
	case FileCapabilityNone, FileCapabilityReadOnly, FileCapabilityReadWrite:
	default:
		return fmt.Errorf("invalid files capability %q", p.Files)
	}
	return nil
}

type Project struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	OrchestrationPolicy string `json:"orchestration_policy"`
	Revision            string `json:"revision"`
	Enabled             bool   `json:"enabled"`
}

type Deployment struct {
	ID              string                `json:"id"`
	ProjectID       string                `json:"project_id"`
	NodeID          string                `json:"node_id"`
	WorkingFolder   string                `json:"working_folder"`
	Role            string                `json:"role"`
	Purpose         string                `json:"purpose"`
	Permissions     DeploymentPermissions `json:"permissions"`
	DesiredRevision string                `json:"desired_revision"`
	AppliedRevision string                `json:"applied_revision"`
	Enabled         bool                  `json:"enabled"`
	ApplyStatus     DeploymentApplyStatus `json:"apply_status"`
	LastError       string                `json:"last_error,omitempty"`
}

type PromptSource struct {
	Path    string `json:"path"`
	Scope   string `json:"scope"`
	SHA256  string `json:"sha256"`
	Bytes   int    `json:"bytes"`
	Content string `json:"content"`
}

type SourceProvenanceKind string

const (
	SourceProvenanceUnknown SourceProvenanceKind = "unknown"
	SourceProvenanceNone    SourceProvenanceKind = "none"
	SourceProvenanceGit     SourceProvenanceKind = "git"
)

// SourceProvenance identifies the exact source tree observed while preparing a
// Project Target. It deliberately contains no remote URL or credentials.
// RepositoryRoot is Node-local diagnostic context; Head/Branch/Dirty are the
// cross-node comparison fields.
type SourceProvenance struct {
	Kind           SourceProvenanceKind `json:"kind"`
	RepositoryRoot string               `json:"repository_root"`
	Head           string               `json:"head"`
	Branch         string               `json:"branch"`
	Detached       bool                 `json:"detached"`
	Unborn         bool                 `json:"unborn"`
	Dirty          bool                 `json:"dirty"`
}

func (p SourceProvenance) Validate() error {
	p.RepositoryRoot = strings.TrimSpace(p.RepositoryRoot)
	p.Head = strings.TrimSpace(p.Head)
	p.Branch = strings.TrimSpace(p.Branch)
	switch p.Kind {
	case SourceProvenanceUnknown:
		if p.RepositoryRoot != "" || p.Head != "" || p.Branch != "" || p.Detached || p.Unborn || p.Dirty {
			return errors.New("unknown source provenance must not contain Git state")
		}
		return nil
	case SourceProvenanceNone:
		if p.RepositoryRoot != "" || p.Head != "" || p.Branch != "" || p.Detached || p.Unborn || p.Dirty {
			return errors.New("non-Git source provenance must not contain Git state")
		}
		return nil
	case SourceProvenanceGit:
		if p.RepositoryRoot == "" {
			return errors.New("Git source provenance requires repository_root")
		}
		if p.Unborn {
			if p.Head != "" || p.Detached || p.Branch == "" {
				return errors.New("unborn Git source provenance requires a branch and no HEAD")
			}
			return nil
		}
		if p.Head == "" {
			return errors.New("Git source provenance requires HEAD")
		}
		if p.Detached {
			if p.Branch != "" {
				return errors.New("detached Git source provenance must not contain a branch")
			}
			return nil
		}
		if p.Branch == "" {
			return errors.New("attached Git source provenance requires branch")
		}
		return nil
	default:
		return fmt.Errorf("invalid source provenance kind %q", p.Kind)
	}
}

type ProjectPrompt struct {
	PromptRevision string         `json:"prompt_revision"`
	Complete       bool           `json:"complete"`
	Bytes          int            `json:"bytes"`
	Sources        []PromptSource `json:"sources"`
}

// WorkSession is the portable Project execution state. Authentication owner/client
// binding remains Nexus-owned state and is deliberately not serialized as model data.
type WorkSession struct {
	ID              string            `json:"work_session_id"`
	ProjectID       string            `json:"project_id"`
	ClientRequestID string            `json:"client_request_id"`
	Status          WorkSessionStatus `json:"status"`
	ContextRevision string            `json:"context_revision"`
	Targets         []WorkTarget      `json:"targets"`
}

type WorkTarget struct {
	ID                 string                `json:"target_id"`
	WorkSessionID      string                `json:"work_session_id"`
	ProjectID          string                `json:"project_id"`
	DeploymentID       string                `json:"deployment_id"`
	NodeID             string                `json:"node_id"`
	CWDRel             string                `json:"cwd_rel"`
	DeploymentRevision string                `json:"deployment_revision"`
	ContextRevision    string                `json:"context_revision"`
	Status             TargetStatus          `json:"status"`
	Permissions        DeploymentPermissions `json:"permissions"`
	Prompt             ProjectPrompt         `json:"prompt"`
	SourceProvenance   SourceProvenance      `json:"source_provenance"`
}

// ToolCallRequest is the canonical Bridge payload inside Message.Arguments for tool.call.
// Trusted Project routing is carried only by Message.ExecutionContext.
type ToolCallRequest struct {
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments"`
}

type ProjectDeploymentRemoveRequest struct {
	DeploymentID string `json:"deployment_id"`
}

type ProjectPromptLoadRequest struct {
	DeploymentID       string `json:"deployment_id"`
	DeploymentRevision string `json:"deployment_revision"`
	CWDRel             string `json:"cwd_rel"`
}

type ProjectPromptLoadResult struct {
	DeploymentID     string           `json:"deployment_id"`
	CWDRel           string           `json:"cwd_rel"`
	Prompt           ProjectPrompt    `json:"prompt"`
	SourceProvenance SourceProvenance `json:"source_provenance"`
}

type ProjectPromptWriteRequest struct {
	DeploymentID       string `json:"deployment_id"`
	DeploymentRevision string `json:"deployment_revision"`
	Scope              string `json:"scope"`
	Content            string `json:"content"`
	ExpectedSHA256     string `json:"expected_sha256,omitempty"`
	Create             bool   `json:"create"`
}

type ProjectPromptWriteResult struct {
	DeploymentID string        `json:"deployment_id"`
	Scope        string        `json:"scope"`
	Source       PromptSource  `json:"source"`
	Prompt       ProjectPrompt `json:"prompt"`
	Created      bool          `json:"created"`
}

type PromptScopeRevision struct {
	Scope          string `json:"scope"`
	PromptRevision string `json:"prompt_revision"`
}

type ProjectTargetBindRequest struct {
	WorkSessionID      string                `json:"work_session_id"`
	TargetID           string                `json:"target_id"`
	ProjectID          string                `json:"project_id"`
	DeploymentID       string                `json:"deployment_id"`
	CWDRel             string                `json:"cwd_rel"`
	DeploymentRevision string                `json:"deployment_revision"`
	ContextRevision    string                `json:"context_revision"`
	PromptScopes       []PromptScopeRevision `json:"prompt_scopes"`
	SourceProvenance   SourceProvenance      `json:"source_provenance"`
}

type ProjectTargetRebindRequest struct {
	WorkSessionID    string                `json:"work_session_id"`
	TargetID         string                `json:"target_id"`
	CWDRel           string                `json:"cwd_rel"`
	ContextRevision  string                `json:"context_revision"`
	PromptScopes     []PromptScopeRevision `json:"prompt_scopes"`
	SourceProvenance SourceProvenance      `json:"source_provenance"`
}

type ProjectTargetRevokeRequest struct {
	TargetID string `json:"target_id"`
}

type ProjectSessionRevokeRequest struct {
	WorkSessionID string `json:"work_session_id"`
}

// ExecutionContext is generated by NexusDock from a bound WorkSession Target.
// Tool arguments must not contain node_id, working_folder, permissions, or other routing overrides.
type ExecutionContext struct {
	WorkSessionID      string `json:"work_session_id"`
	TargetID           string `json:"target_id"`
	ProjectID          string `json:"project_id"`
	DeploymentID       string `json:"deployment_id"`
	DeploymentRevision string `json:"deployment_revision"`
	ContextRevision    string `json:"context_revision"`
}

func OperationRequiresExecutionContext(operation string) bool {
	return operation == OperationToolCall
}

func ValidateExecutionContext(context *ExecutionContext) *RemoteError {
	if context == nil {
		return &RemoteError{
			Code:     ErrorExecutionContextRequired,
			Message:  "Project execution context is required",
			Category: "authorization",
		}
	}
	fields := []struct {
		name  string
		value string
	}{
		{name: "work_session_id", value: context.WorkSessionID},
		{name: "target_id", value: context.TargetID},
		{name: "project_id", value: context.ProjectID},
		{name: "deployment_id", value: context.DeploymentID},
		{name: "deployment_revision", value: context.DeploymentRevision},
		{name: "context_revision", value: context.ContextRevision},
	}
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			return &RemoteError{
				Code:     ErrorExecutionContextInvalid,
				Message:  "Project execution context is incomplete",
				Category: "authorization",
				Details:  map[string]any{"field": field.name},
			}
		}
	}
	return nil
}
