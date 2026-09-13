package mcpcontract

const (
	ToolAgentDockContext       = "agentdock_context"
<<<<<<< HEAD
=======
	ToolProjectList            = "project_list"
	ToolProjectOpen            = "project_open"
	ToolProjectContext         = "project_context"
>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)
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

<<<<<<< HEAD
=======
var nexusOnlyToolNames = []string{
	ToolProjectList,
	ToolProjectOpen,
	ToolProjectContext,
}

>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)
// ToolNames returns the canonical model-facing tools shared by AgentDock and NexusDock.
func ToolNames() []string { return append([]string(nil), toolNames...) }

func IsCanonicalTool(name string) bool {
	for _, candidate := range toolNames {
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
	readOnly := false
	destructive := true
	var idempotent *bool
	switch name {
<<<<<<< HEAD
	case ToolAgentDockContext, ToolRecallSearch, ToolRecallRead:
=======
	case ToolAgentDockContext, ToolProjectList, ToolRecallSearch, ToolRecallRead:
>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)
		readOnly = true
		destructive = false
	case ToolProjectOpen, ToolProjectContext:
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
