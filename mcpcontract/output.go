package mcpcontract

// OutputSchema returns a fresh canonical output schema for tools whose result shape
// is shared by direct AgentDock and central NexusDock entrypoints.
func OutputSchema(name string) (map[string]any, bool) {
	if schema, ok := continuationOutputSchema(name); ok {
		return schema, true
	}
	props := map[string]any{}
	var required []string
	strict := false
	switch name {
	case ToolProjectList:
		props["projects"] = map[string]any{
			"type":        "array",
			"description": "Projects the authenticated client may enter.",
			"items":       projectSummarySchema(),
		}
		required = []string{"projects"}
		strict = true
	case ToolProjectOpen:
		props["work_session_id"] = stringProperty("Bound WorkSession id created or resolved idempotently for this request.")
		props["status"] = enumProperty("WorkSession state.", "preparing", "ready", "running", "completed", "partial", "failed", "cancelled")
		props["project"] = projectSchema()
		props["deployments"] = map[string]any{
			"type":        "array",
			"description": "Project Deployment topology including unavailable Deployments and explicit reasons.",
			"items":       deploymentViewSchema(),
		}
		props["targets"] = map[string]any{
			"type":        "array",
			"description": "WorkSession Targets prepared from authorized Deployments. Discovery never implies that all Targets were executed.",
			"items":       workTargetSchema(),
		}
		required = []string{"work_session_id", "status", "project", "deployments", "targets"}
		strict = true
	case ToolNodeOpen:
		props["work_session_id"] = stringProperty("Bound WorkSession id created or resolved idempotently for this request.")
		props["status"] = enumProperty("WorkSession state.", "preparing", "ready", "running", "completed", "partial", "failed", "cancelled")
		props["node"] = nodeOpenSchema()
		props["targets"] = map[string]any{
			"type":        "array",
			"minItems":    1,
			"maxItems":    1,
			"description": "The single WorkSession Target prepared for this Node.",
			"items":       workTargetSchema(),
		}
		required = []string{"work_session_id", "status", "node", "targets"}
		strict = true
	case ToolProjectContext:
		props["work_session_id"] = stringProperty("Bound WorkSession id.")
		props["project"] = projectSchema()
		props["node"] = nodeOpenSchema()
		props["deployment"] = deploymentViewSchema()
		props["target"] = workTargetSchema()
		required = []string{"work_session_id", "deployment", "target"}
		strict = true
	case ToolRecallSearch:
		props["results"] = map[string]any{
			"type": "array", "description": "Recall search results with source identity fields.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"frontmatter":    objectProperty("Recall frontmatter metadata."),
					"matched_fields": arrayStrings("Matched fields."),
					"matched_terms":  arrayStrings("Matched terms."),
					"path":           stringProperty("Recall-relative document path."),
					"snippet":        stringProperty("Matched content snippet."),
					"id":             stringProperty("Stable Recall document id."),
					"title":          stringProperty("Human-readable document title."),
					"url":            stringProperty("Absolute source URL."),
				},
				"required": []string{"id", "title", "url"}, "additionalProperties": true,
			},
		}
	case ToolRecallRead:
		props["recall"] = objectProperty("NexusDock Recall document. Raw Markdown is returned only when include_raw=true.")
	case ToolRecallWrite:
		props["recall_target"] = stringProperty("Recall target used.")
		props["recall_action"] = stringProperty("Recall action used.")
		props["recall"] = objectProperty("NexusDock Recall document returned when a write occurs.")
		props["card"] = objectProperty("Normalized card candidate or written card when target=card.")
		props["warnings"] = arrayObjects("Review warnings before writing.")
		props["capture_plan"] = objectProperty("Reviewable write plan for card captures.")
		props["similar_results"] = arrayObjects("Similar existing card search results.")
		props["path"] = stringProperty("NexusDock Recall-relative path.")
		props["changed"] = booleanProperty("Whether the proposed edit changes content.")
		props["dry_run"] = booleanProperty("Whether the operation only previewed changes.")
		props["written"] = booleanProperty("Whether the entry was written.")
		props["diff"] = stringProperty("Unified diff preview.")
		props["updates"] = arrayObjects("Fact update results.")
	case ToolRecallMaintain:
		props["entries"] = arrayObjects("NexusDock Recall entries for action=list.")
		props["terms"] = arrayStrings("Terms used for action=lint.")
		props["findings"] = arrayObjects("Lint findings.")
	case ToolPrivateNoteManage:
		props["results"] = arrayObjects("Metadata-only private-note search results; never plaintext snippets.")
		props["path"] = stringProperty("Plain note path for read/write/delete results.")
		props["encrypted_path"] = stringProperty("Age encrypted backup path.")
		props["content"] = stringProperty("Plaintext content returned only by explicit action=read.")
		props["truncated"] = booleanProperty("Whether returned content was truncated.")
		props["contains_secret"] = booleanProperty("Whether the note content is marked as containing secrets.")
		props["written"] = booleanProperty("Whether plaintext was written.")
		props["encrypted"] = booleanProperty("Whether encrypted backup was written.")
		props["deleted_plaintext"] = booleanProperty("Whether plaintext was deleted.")
		props["deleted_encrypted"] = booleanProperty("Whether encrypted backup was deleted.")
		props["notes"] = arrayObjects("Metadata-only private note summaries for status/list.")
		props["notes_count"] = integerProperty("Private note count for status checks.")
		props["encrypted_count"] = integerProperty("Encrypted backup count for maintenance actions.")
		props["recipient"] = stringProperty("Age public recipient generated or used.")
		props["identity_created"] = booleanProperty("Whether a new local age identity was created.")
		props["algorithm"] = stringProperty("Encryption algorithm.")
		props["missing_encrypted"] = arrayStrings("Missing encrypted backup paths.")
		props["encrypted_backup_ok"] = booleanProperty("Whether every private note has its required encrypted backup.")
		props["plaintext_git_ignored"] = booleanProperty("Whether private note plaintext is Git-ignored.")
		props["keys_git_ignored"] = booleanProperty("Whether private note keys are Git-ignored.")
	case ToolWorkflowTemplateManage:
		props["action"] = stringProperty("Completed workflow template action.")
		props["template"] = objectProperty("Full workflow template returned by get.")
		props["templates"] = arrayObjects("Compact summaries from list or full active templates from get_many.")
		props["composition_required"] = booleanProperty("Whether the returned templates must be combined by the model before task creation.")
		props["next_required_action"] = stringProperty("Required model action after get_many.")
		props["template_id"] = stringProperty("Workflow template id returned by publish or retire.")
		props["template_summary"] = objectProperty("Compact workflow template summary returned by publish, retire, and list items.")
		props["candidates"] = arrayObjects("Matched workflow template candidates with scores and reasons.")
		props["vector_search_enabled"] = booleanProperty("Whether optional embedding-backed template vector search is enabled for match.")
		props["vector_index_status"] = stringProperty("Template vector index status: disabled, ready, or degraded.")
		props["vector_index_items"] = integerProperty("Number of persisted template vectors for the current embedding model.")
		props["vector_index_available"] = booleanProperty("Whether workflow vector index content is available for export.")
		props["embedding_model"] = stringProperty("Embedding model configured for template vector search.")
		props["recommended"] = stringProperty("Template recommendation: use_template, consider_template, or plain_task.")
		props["recommendation_reason"] = stringProperty("Reason for recommendation.")
		props["best_candidate_score"] = integerProperty("Highest template match score.")
		props["score_thresholds"] = objectProperty("Template match score thresholds.")
	default:
		return nil, false
	}
	if strict {
		schema := strictObject(props, required...)
		if name == ToolProjectContext {
			schema["oneOf"] = []any{
				map[string]any{"required": []string{"project"}, "not": map[string]any{"required": []string{"node"}}},
				map[string]any{"required": []string{"node"}, "not": map[string]any{"required": []string{"project"}}},
			}
		}
		return schema, true
	}
	return map[string]any{"type": "object", "properties": props, "required": []string{}, "additionalProperties": true}, true
}

func projectSummarySchema() map[string]any {
	return strictObject(map[string]any{
		"id":                         stringProperty("Stable Project id."),
		"name":                       stringProperty("Project display name."),
		"deployment_count":           integerProperty("Configured Deployment count."),
		"available_deployment_count": integerProperty("Deployments currently eligible to become WorkSession Targets."),
	}, "id", "name", "deployment_count", "available_deployment_count")
}

func projectSchema() map[string]any {
	return strictObject(map[string]any{
		"id":                   stringProperty("Stable Project id."),
		"name":                 stringProperty("Project display name."),
		"orchestration_policy": stringProperty("User-authored multi-node collaboration guidance. It never expands hard permissions."),
	}, "id", "name", "orchestration_policy")
}

func nodeOpenSchema() map[string]any {
	return strictObject(map[string]any{
		"node_id": stringProperty("Stable AgentDock Node id."),
		"name":    stringProperty("AgentDock Node display name."),
		"online":  booleanProperty("Whether the AgentDock Node is currently online."),
	}, "node_id", "name", "online")
}

func deploymentPermissionsSchema() map[string]any {
	return strictObject(map[string]any{
		"full_access": booleanProperty("Whether all Project execution capabilities exposed by this Node are allowed. Independent from working_folder."),
		"files":       enumProperty("Built-in file capability when full_access is false.", "none", "read_only", "read_write"),
		"shell":       booleanProperty("Whether command execution is allowed when full_access is false. This is not an OS sandbox."),
		"browser":     booleanProperty("Whether browser capabilities are allowed when full_access is false."),
		"dynamic_mcp": booleanProperty("Whether configured dynamic MCP calls are allowed when full_access is false."),
		"acp":         booleanProperty("Whether ACP agent execution is allowed when full_access is false."),
	}, "full_access", "files", "shell", "browser", "dynamic_mcp", "acp")
}

func deploymentViewSchema() map[string]any {
	return strictObject(map[string]any{
		"id":                  stringProperty("Stable Deployment id."),
		"working_folder":      stringProperty("Optional Node-local Project Folder. Empty means use the Node AgentDock default cwd and do not auto-discover Project AGENTS.md."),
		"role":                stringProperty("User-authored short role label; not an authorization role."),
		"purpose":             stringProperty("User-authored description of when this environment is useful."),
		"permissions":         deploymentPermissionsSchema(),
		"availability_status": stringProperty("Preparation status or explicit reason this Deployment cannot become a Target."),
		"last_error":          stringProperty("Safe latest apply or preparation error; empty when none."),
	}, "id", "working_folder", "role", "purpose", "permissions", "availability_status")
}

func projectPromptSchema() map[string]any {
	source := strictObject(map[string]any{
		"path":    stringProperty("AGENTS.md path relative to the Deployment working folder."),
		"scope":   stringProperty("Project-relative directory scope where this source applies."),
		"content": stringProperty("Complete source body. Successful complete=true results never silently truncate it."),
	}, "path", "scope", "content")
	return strictObject(map[string]any{
		"complete": booleanProperty("Whether every applicable source was loaded and returned completely."),
		"sources":  map[string]any{"type": "array", "items": source},
	}, "complete", "sources")
}

func sourceProvenanceSchema() map[string]any {
	return strictObject(map[string]any{
		"kind":            enumProperty("Observed source-control kind. unknown means the Target was not inspected yet; none means the working folder is not inside a Git worktree.", "unknown", "none", "git"),
		"repository_root": stringProperty("Node-local Git repository root path; empty for non-Git Targets."),
		"head":            stringProperty("Observed Git HEAD commit; empty only for non-Git or unborn repositories."),
		"branch":          stringProperty("Observed branch name; empty for detached HEAD or non-Git Targets."),
		"detached":        booleanProperty("Whether Git HEAD is detached."),
		"unborn":          booleanProperty("Whether the Git repository has no HEAD commit yet."),
		"dirty":           booleanProperty("Whether tracked or untracked working-tree changes were observed."),
	}, "kind", "repository_root", "head", "branch", "detached", "unborn", "dirty")
}

func workTargetSchema() map[string]any {
	return strictObject(map[string]any{
		"target_id":     stringProperty("Stable Target id within this WorkSession."),
		"deployment_id": stringProperty("Bound Deployment id."),
		"cwd_rel":       stringProperty("Current Project-relative Target working directory."),
		"status":        enumProperty("Target state.", "preparing", "ready", "running", "idle", "unavailable", "context_error", "revoked"),
		"prompt":        projectPromptSchema(),
	}, "target_id", "deployment_id", "cwd_rel", "status", "prompt")
}

func LocalAgentDockContextOutputSchema() map[string]any {
	props := localContextProperties(true)
	props["runtime"] = agentDockRuntimeSchema()
	return strictObject(props, "runtime", "skills", "common_skills", "dynamic_mcp", "workflow_templates")
}

func FleetAgentDockContextOutputSchema() map[string]any {
	item := contextItemSchema(false)
	warning := map[string]any{
		"type": "object", "properties": map[string]any{
			"source": stringProperty("Context section identifier."), "message": stringProperty("Safe warning message."),
		}, "required": []string{"source", "message"}, "additionalProperties": false,
	}
	local := strictObject(localContextProperties(false), "skills", "common_skills", "dynamic_mcp")
	shared := strictObject(map[string]any{
		"workflow_templates": map[string]any{"type": "array", "items": item},
		"recall": strictObject(map[string]any{
			"enabled": booleanProperty("Whether NexusDock Recall is available."),
			"items":   map[string]any{"type": "array", "items": item},
		}, "enabled", "items"),
		"warnings": map[string]any{"type": "array", "items": warning},
	}, "workflow_templates", "recall")
	return strictObject(map[string]any{
		"nodes": map[string]any{
			"type": "array", "description": "Enabled AgentDock nodes and their node-local context.",
			"items": strictObject(map[string]any{
				"node_id":           map[string]any{"type": "string"},
				"name":              map[string]any{"type": "string"},
				"online":            map[string]any{"type": "boolean"},
				"version":           map[string]any{"type": "string"},
				"os":                map[string]any{"type": "string"},
				"arch":              map[string]any{"type": "string"},
				"capabilities":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"capability_status": stringProperty("Node capability context status such as ready, offline, unsupported, timeout, malformed, or unavailable."),
				"context":           local,
				"error":             map[string]any{"type": "string"},
			}, "node_id", "name", "online", "capabilities", "capability_status"),
		},
		"shared": shared,
	}, "nodes", "shared")
}

func localContextProperties(includeShared bool) map[string]any {
	skill := map[string]any{
		"type": "object", "properties": map[string]any{
			"name": stringProperty("Skill name."), "description": stringProperty("Short capability description."),
			"file": stringProperty("skill:// URI for the active SKILL.md."), "bundled": booleanProperty("Whether the Skill is bundled by AgentDock."),
		}, "required": []string{"name", "description", "file"}, "additionalProperties": false,
	}
	commonSkill := map[string]any{
		"type": "object", "properties": map[string]any{
			"name": stringProperty("Common Skill name."), "description": stringProperty("Short capability description."),
			"file": stringProperty("Host path to the common SKILL.md."),
		}, "required": []string{"name", "description", "file"}, "additionalProperties": false,
	}
	commonSkills := strictObject(map[string]any{
		"root":      stringProperty("Common Agent Skills root path."),
		"total":     integerProperty("Total valid common Skills discovered before truncation."),
		"effective": integerProperty("Common Skills remaining after installed AgentDock Skills shadow same-name entries."),
		"shadowed":  integerProperty("Common Skills hidden because an installed AgentDock Skill takes precedence."),
		"truncated": booleanProperty("Whether the common Skill index was truncated."),
		"items":     map[string]any{"type": "array", "items": commonSkill},
	}, "root", "total", "effective", "shadowed", "truncated", "items")
	commonSkills["description"] = "Lower-priority common Agent Skill capability index; installed AgentDock Skills take precedence on conflicts."
	dynamicItem := dynamicMCPItemSchema()
	indexItem := contextItemSchema(false)
	warning := map[string]any{
		"type": "object", "properties": map[string]any{
			"source": stringProperty("Context section identifier."), "message": stringProperty("Safe warning message."),
		}, "required": []string{"source", "message"}, "additionalProperties": false,
	}
	props := map[string]any{
		"skills":        map[string]any{"type": "array", "description": "Installed document Skill capability index.", "items": skill},
		"common_skills": commonSkills,
		"dynamic_mcp":   map[string]any{"type": "array", "description": "Enabled dynamic MCP server capability index.", "items": dynamicItem},
		"acp": strictObject(map[string]any{
			"enabled": booleanProperty("Whether ACP is enabled."), "agent": stringProperty("Configured ACP agent name."), "description": stringProperty("Short ACP usage orientation."),
		}, "enabled", "agent", "description"),
		"warnings": map[string]any{"type": "array", "description": "Best-effort context sections that could not be loaded.", "items": warning},
	}
	if includeShared {
		props["workflow_templates"] = map[string]any{"type": "array", "description": "Active NexusDock Workflow template index; empty when Nexus is unavailable.", "items": indexItem}
		props["recall"] = strictObject(map[string]any{
			"enabled": booleanProperty("Whether NexusDock Recall context is configured."),
			"items":   map[string]any{"type": "array", "items": indexItem},
		}, "enabled", "items")
	}
	return props
}

func agentDockRuntimeSchema() map[string]any {
	return strictObject(map[string]any{
		"version":               stringProperty("AgentDock runtime version."),
		"os":                    stringProperty("Host operating system."),
		"arch":                  stringProperty("Host architecture."),
		"agentdock_home":        stringProperty("AgentDock state and configuration directory."),
		"agentdock_default_dir": stringProperty("AgentDock default working directory."),
		"default_cwd":           stringProperty("Default cwd relative to the AgentDock working directory when applicable."),
		"path_model":            stringProperty("Path resolution model used by host tools."),
	}, "version", "os", "arch", "agentdock_home", "agentdock_default_dir", "default_cwd", "path_model")
}

func contextItemSchema(requireDescription bool) map[string]any {
	required := []string{"name"}
	if requireDescription {
		required = append(required, "description")
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":        stringProperty("Capability name."),
			"description": stringProperty("Short capability description."),
		},
		"required":             required,
		"additionalProperties": false,
	}
}

func dynamicMCPItemSchema() map[string]any {
	schema := contextItemSchema(true)
	properties := schema["properties"].(map[string]any)
	properties["status"] = map[string]any{
		"type": "string", "description": "Runtime health state: idle, ready, or error.",
		"enum": []string{"idle", "ready", "error"},
	}
	properties["tool_count"] = integerProperty("Discovered tool count from the latest successful refresh.")
	properties["last_error_code"] = stringProperty("Safe machine-readable code for the latest refresh error.")
	return schema
}
