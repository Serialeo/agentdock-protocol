package protocol

<<<<<<< HEAD
const ConnectionProtocolVersion = "2"
=======
// ConnectionProtocolVersion is intentionally incompatible with the legacy Bridge v3 wire.
// Generation 4 requires Project-scoped execution context for model-facing OS tool calls.
const ConnectionProtocolVersion = "4"
>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)

const (
	MessageNodeHello     = "node.hello"
	MessageNodeReady     = "node.ready"
	MessageNodeHeartbeat = "node.heartbeat"
	MessageToolInvoke    = "tool.invoke"
	MessageToolResult    = "tool.result"
	MessageToolError     = "tool.error"
	MessageToolCancel    = "tool.cancel"
)

const (
	OperationRuntimeRequest          = "runtime.request"
	OperationContextLocal            = "context.local"
	OperationToolCall                = "tool.call"
	OperationResourceRead            = "resource.read"
	OperationArtifactRead            = "artifact.read"
	OperationProjectDeploymentApply  = "project.deployment.apply"
	OperationProjectDeploymentRemove = "project.deployment.remove"
	OperationProjectPromptLoad       = "project.prompt.load"
	OperationProjectPromptWrite      = "project.prompt.write"
	OperationProjectTargetBind       = "project.target.bind"
	OperationProjectTargetRebind     = "project.target.rebind"
	OperationProjectTargetRevoke     = "project.target.revoke"
	OperationProjectSessionRevoke    = "project.session.revoke"
)

const (
	// ArtifactReadCapability is the frozen Bridge capability token for private Artifact reads.
	ArtifactReadCapability = "bridge.artifact.read.v1"
	// MaxArtifactChunkBytes bounds one raw Artifact chunk before JSON/base64 encoding on the Bridge.
	MaxArtifactChunkBytes = 512 << 10
)

const (
	// MaxProjectPromptFileBytes bounds one AGENTS.md source before composition.
	MaxProjectPromptFileBytes = 64 << 10
	// MaxProjectPromptTargetBytes bounds the complete applicable AGENTS.md chain for one Target.
	MaxProjectPromptTargetBytes = 256 << 10
	// MaxProjectPromptContextBytes bounds complete Project Prompt source text returned for one WorkSession preparation.
	MaxProjectPromptContextBytes = 1 << 20
	// MaxProjectContextDeliveryBytes bounds the final model-visible Project Context JSON envelope.
	// It remains comfortably below the 8 MiB Bridge message ceiling while allowing metadata around the 1 MiB source-body budget.
	MaxProjectContextDeliveryBytes = 4 << 20
)

const (
	ErrorExecutionContextRequired = "EXECUTION_CONTEXT_REQUIRED"
	ErrorExecutionContextInvalid  = "EXECUTION_CONTEXT_INVALID"
	ErrorProjectNotFound          = "PROJECT_NOT_FOUND"
	ErrorDeploymentNotReady       = "DEPLOYMENT_NOT_READY"
	ErrorSessionTargetDenied      = "SESSION_TARGET_DENIED"
	ErrorCapabilityDenied         = "CAPABILITY_DENIED"
	ErrorContextRefreshRequired   = "CONTEXT_REFRESH_REQUIRED"
	ErrorProjectPromptReadFailed  = "PROJECT_PROMPT_READ_FAILED"
	ErrorProjectPromptTooLarge    = "PROJECT_PROMPT_TOO_LARGE"
	ErrorPromptScopeEscape        = "PROMPT_SCOPE_ESCAPE"
	ErrorRevisionConflict         = "REVISION_CONFLICT"
	ErrorExecutionOutcomeUnknown  = "EXECUTION_OUTCOME_UNKNOWN"
)

const AgentDockUIResourcePrefix = "ui://agentdock/"

const (
	ContextUIResourceURI      = "ui://agentdock/context"
	TaskProgressUIResourceURI = "ui://agentdock/task-progress"
	FileChangeUIResourceURI   = "ui://agentdock/file-change"
	RecallUIResourceURI       = "ui://agentdock/recall"
	WorkflowUIResourceURI     = "ui://agentdock/workflow"
	DynamicMCPUIResourceURI   = "ui://agentdock/dynamic-mcp"
	ACPStatusUIResourceURI    = "ui://agentdock/acp-status"
)

const (
	ContextUIContract      = "agentdock.context.fleet.v1"
	TaskProgressUIContract = "agentdock.task-progress.v1"
	FileChangeUIContract   = "agentdock.file-change.v1"
	RecallUIContract       = "agentdock.recall.v1"
	WorkflowUIContract     = "agentdock.workflow.v1"
	DynamicMCPUIContract   = "agentdock.dynamic-mcp.v1"
	ACPStatusUIContract    = "agentdock.acp-status.v1"
)

const MCPAppMIMEType = "text/html;profile=mcp-app"

// UIResourceContract returns the renderer contract bound to one AgentDock MCP App URI.
// URIs identify resources and remain stable; renderer compatibility evolves through the contract string.
func UIResourceContract(uri string) (string, bool) {
	switch uri {
	case ContextUIResourceURI:
		return ContextUIContract, true
	case TaskProgressUIResourceURI:
		return TaskProgressUIContract, true
	case FileChangeUIResourceURI:
		return FileChangeUIContract, true
	case RecallUIResourceURI:
		return RecallUIContract, true
	case WorkflowUIResourceURI:
		return WorkflowUIContract, true
	case DynamicMCPUIResourceURI:
		return DynamicMCPUIContract, true
	case ACPStatusUIResourceURI:
		return ACPStatusUIContract, true
	default:
		return "", false
	}
}
