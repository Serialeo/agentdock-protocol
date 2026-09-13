package mcpcontract

// InputSchema returns a fresh canonical model-facing input schema.
func InputSchema(name string) (map[string]any, bool) {
	props := map[string]any{}
	var required []string
	switch name {
	case ToolAgentDockContext, ToolProjectList:
	case ToolProjectOpen:
		props["project_id"] = stringProperty("Stable NexusDock Project id to enter.")
		props["client_request_id"] = stringProperty("Client-generated idempotency key for this WorkSession creation request.")
		props["targets"] = map[string]any{
			"type":        "array",
			"minItems":    1,
			"description": "Optional explicit Deployment targets for this WorkSession. Omit to discover available Project Deployments; an explicit empty array is invalid.",
			"items": strictObject(map[string]any{
				"deployment_id": stringProperty("Stable Deployment id selected for this WorkSession."),
				"cwd_rel":       stringProperty("Optional starting subdirectory for this Target. Defaults to the configured Project Folder root, or the Node AgentDock default cwd when no Project Folder is configured."),
			}, "deployment_id"),
		}
		required = []string{"project_id", "client_request_id"}
	case ToolProjectContext:
		props["work_session_id"] = stringProperty("Bound WorkSession id returned by project_open.")
		props["target_id"] = stringProperty("Target id owned by the bound WorkSession.")
		props["cwd_rel"] = stringProperty("Optional Project-relative subdirectory to bind before refreshing the complete applicable Project Prompt.")
		required = []string{"work_session_id", "target_id"}
	case ToolRecallSearch:
		props["query"] = stringProperty("Text query to search in NexusDock Recall files and paths.")
		props["kind"] = enumProperty("Search kind. Defaults to all.", "all", "markdown", "card")
		props["max_results"] = integerProperty("Maximum results to return.")
		required = []string{"query"}
	case ToolRecallRead:
		props["path"] = stringProperty("NexusDock Recall-relative Markdown or card path.")
		props["include_raw"] = booleanProperty("Include raw Markdown as raw_content. Defaults to false to avoid duplicating body/content tokens.")
		required = []string{"path"}
	case ToolRecallWrite:
		props["target"] = enumProperty("Recall target selected by the model.", "card", "markdown")
		props["action"] = enumProperty("Recall action selected by the model.", "plan", "create", "replace", "append", "patch", "update_fact", "diff", "delete")
		props["confirmed"] = booleanProperty("Required for destructive or protected writes. Markdown inbox create may write without confirmation; card create and edit actions without confirmation return a review preview where supported.")
		props["path"] = stringProperty("NexusDock Recall-relative path when reading, updating, deleting, or writing a known entry.")
		props["content"] = stringProperty("Card or Markdown content, or proposed replacement content.")
		props["title"] = stringProperty("Short title for a card or Markdown entry.")
		props["summary"] = stringProperty("Short summary for a card.")
		props["overwrite"] = booleanProperty("Card only: replace an existing generated card path when explicitly supported. Markdown replace implies overwrite by its action semantics.")
		props["allow_warnings"] = booleanProperty("Card only: after reviewing warnings, allow writing a warned card. Do not use by default.")
		props["old"] = stringProperty("Patch only: literal text to replace.")
		props["new"] = stringProperty("Patch only: replacement text for old.")
		props["append"] = stringProperty("Append/patch only: text to append to the Recall document.")
		props["section"] = stringProperty("Patch/update_fact only: Markdown heading title whose section should be updated.")
		props["section_content"] = stringProperty("Patch only: new body for the selected Markdown section.")
		props["key"] = stringProperty("Update_fact only: fact key to update.")
		props["value"] = stringProperty("Update_fact only: new fact value.")
		props["facts"] = objectProperty("Update_fact only: multiple key/value facts to update; values are normalized to strings by the runtime.")
		props["append_if_missing"] = booleanProperty("Update_fact only: append missing keys to the selected section or document instead of failing.")
		props["max_bytes"] = integerProperty("Maximum diff/output bytes.")
		required = []string{"target", "action"}
	case ToolRecallMaintain:
		props["action"] = enumProperty("Maintenance action.", "list", "lint", "embedding_status", "reindex", "reindex_cards")
		props["prefix"] = stringProperty("Optional NexusDock Recall-relative prefix.")
		props["terms"] = arrayStrings("Terms or regex patterns for lint.")
		props["regex"] = booleanProperty("Treat terms as regex patterns for lint.")
		props["max_entries"] = integerProperty("Maximum entries to list or scan.")
		props["max_findings"] = integerProperty("Maximum lint findings to return.")
		props["max_results"] = integerProperty("Maximum results where supported.")
	case ToolPrivateNoteManage:
		props["action"] = enumProperty("NexusDock private note action. Do not use by default; use only for explicit private note access or clearly sensitive secrets, credentials, or personal information.", "search", "read", "write", "delete", "status", "maintain")
		props["query"] = stringProperty("Metadata-only query for action=search. Matches title, summary, tags, category, and path; never searches plaintext body.")
		props["max_results"] = boundedIntegerProperty("Maximum metadata search results to return. Defaults to 8 and is capped at 100.", 1, 100)
		props["path"] = stringProperty("Path under notes/ for action=read, action=write, or action=delete.")
		props["category"] = stringProperty("Optional category used with title when path is omitted. Defaults to services.")
		props["title"] = stringProperty("Title used for frontmatter or to derive the path when path is omitted.")
		props["summary"] = stringProperty("Optional human-maintained safe summary for metadata-only search.")
		props["tags"] = arrayStrings("Optional safe tags for metadata-only search.")
		props["content"] = stringProperty("Plaintext private note content for action=write.")
		props["confirmed"] = booleanProperty("Required for true action=write and action=delete mutations.")
		props["overwrite"] = booleanProperty("Replace an existing note for action=write.")
		props["max_bytes"] = boundedIntegerProperty("Maximum bytes to return for explicit action=read. Defaults to 256000 and is capped at 1048576.", 1, 1048576)
		props["status_action"] = enumProperty("Read-only status action when action=status.", "check", "list")
		props["maintenance_action"] = enumProperty("NexusDock encryption maintenance operation when action=maintain.", "init", "init-encryption", "sync-encrypted", "encrypt-all")
		required = []string{"action"}
	case ToolWorkflowTemplateManage:
		props["action"] = enumProperty("Workflow template action. publish accepts a complete template; get_many returns full active templates that the model must compose before task creation.", "publish", "retire", "list", "get", "get_many", "match", "vector_index")
		// The root tool contract is strict, but a complete template is an intentionally open domain payload.
		props["template"] = openObject("Complete workflow template for publish.")
		props["template_id"] = stringProperty("Workflow template id.")
		props["template_ids"] = map[string]any{"type": "array", "minItems": 2, "maxItems": 3, "items": map[string]any{"type": "string"}, "description": "Two or three active template ids for get_many. The returned templates must be pruned, deduplicated, ordered, and combined by the model."}
		props["template_version"] = stringProperty("Workflow template version for exact get or retire actions. Omit it for get to resolve the current active version.")
		props["template_status"] = enumProperty("Optional list status filter.", "active", "retired")
		props["allow_long_template"] = booleanProperty("Allow a workflow template to exceed default guardrails. Provide long_template_reason when true.")
		props["long_template_reason"] = stringProperty("Reason required when allow_long_template=true.")
		props["goal"] = stringProperty("Goal text for match.")
		props["device"] = stringProperty("Optional device hint for match.")
		props["type"] = stringProperty("Optional workflow type hint for match. This maps to template match.type.")
		required = []string{"action"}
	default:
		return nil, false
	}
	return strictObject(props, required...), true
}
