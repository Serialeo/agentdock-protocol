package mcpcontract

const (
	ToolAgentDockContext       = "agentdock_context"
	ToolProjectList            = "project_list"
	ToolProjectOpen            = "project_open"
	ToolProjectContext         = "project_context"
	ToolNodeOpen               = "node_open"
	ToolRecallSearch           = "recall_search"
	ToolRecallRead             = "recall_read"
	ToolRecallWrite            = "recall_write"
	ToolRecallMaintain         = "recall_maintain"
	ToolPrivateNoteManage      = "private_note_manage"
	ToolWorkflowTemplateManage = "workflow_template_manage"
)

var toolNames = []string{
	ToolAgentDockContext,
	ToolRecallSearch,
	ToolRecallRead,
	ToolRecallWrite,
	ToolRecallMaintain,
	ToolPrivateNoteManage,
	ToolWorkflowTemplateManage,
}

var nexusOnlyToolNames = []string{
	ToolProjectList,
	ToolProjectOpen,
	ToolProjectContext,
	ToolNodeOpen,
}

// ToolNames returns the canonical model-facing tools shared by AgentDock and NexusDock.
func ToolNames() []string { return append([]string(nil), toolNames...) }

// NexusToolNames returns all canonical NexusDock-facing tools, including the
// Project and continuation tools (model/app) that standalone AgentDock does not expose.
func NexusToolNames() []string {
	result := append([]string(nil), toolNames...)
	result = append(result, nexusOnlyToolNames...)
	return append(result, continuationToolNames...)
}

func IsCanonicalTool(name string) bool {
	if _, ok := ToolVisibility(name); ok {
		return true
	}
	for _, candidate := range toolNames {
		if candidate == name {
			return true
		}
	}
	for _, candidate := range nexusOnlyToolNames {
		if candidate == name {
			return true
		}
	}
	return false
}

type Annotations struct {
	ReadOnlyHint    bool
	DestructiveHint *bool
	IdempotentHint  *bool
	OpenWorldHint   *bool
}

func AnnotationContract(name string) (Annotations, bool) {
	if annotations, ok := continuationAnnotations(name); ok {
		return annotations, true
	}
	readOnly := false
	destructive := true
	var idempotent *bool
	switch name {
	case ToolAgentDockContext, ToolProjectList, ToolRecallSearch, ToolRecallRead:
		readOnly = true
		destructive = false
	case ToolProjectOpen, ToolProjectContext, ToolNodeOpen:
		destructive = false
		idempotent = boolPtr(true)
	case ToolRecallWrite, ToolRecallMaintain, ToolPrivateNoteManage, ToolWorkflowTemplateManage:
	default:
		return Annotations{}, false
	}
	openWorld := false
	return Annotations{
		ReadOnlyHint:    readOnly,
		DestructiveHint: boolPtr(destructive),
		IdempotentHint:  idempotent,
		OpenWorldHint:   boolPtr(openWorld),
	}, true
}

func boolPtr(value bool) *bool { return &value }
