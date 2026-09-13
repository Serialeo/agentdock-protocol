package mcpcontract

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCanonicalToolContractsAreCompleteAndFresh(t *testing.T) {
	if got := len(ToolNames()); got != 7 {
<<<<<<< HEAD
		t.Fatalf("tool count = %d, want 7", got)
	}
	for _, name := range ToolNames() {
=======
		t.Fatalf("shared tool count = %d, want 7", got)
	}
	if got := len(NexusToolNames()); got != 10 {
		t.Fatalf("Nexus tool count = %d, want 10", got)
	}
	projectTools := map[string]bool{
		ToolProjectList:    true,
		ToolProjectOpen:    true,
		ToolProjectContext: true,
	}
	for name := range projectTools {
		if !IsCanonicalTool(name) {
			t.Fatalf("Project tool %q is not reserved as canonical", name)
		}
	}
	for _, name := range ToolNames() {
		if projectTools[name] {
			t.Fatalf("Nexus-only Project tool %q leaked into shared AgentDock tool names", name)
		}
	}
	if IsCanonicalTool("agentdock_guidance") {
		t.Fatal("legacy agentdock_guidance remains canonical in the v4 contract")
	}
	for _, name := range NexusToolNames() {
>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)
		input, ok := InputSchema(name)
		if !ok {
			t.Fatalf("missing input schema for %s", name)
		}
		if input["additionalProperties"] != false {
			t.Fatalf("%s root input must be strict: %#v", name, input["additionalProperties"])
		}
		if _, ok := AnnotationContract(name); !ok {
			t.Fatalf("missing annotations for %s", name)
		}
		if name == ToolAgentDockContext {
			continue
		}
		if _, ok := OutputSchema(name); !ok {
			t.Fatalf("missing output schema for %s", name)
		}
	}

	first, _ := InputSchema(ToolRecallSearch)
	second, _ := InputSchema(ToolRecallSearch)
	first["additionalProperties"] = true
	if reflect.DeepEqual(first, second) {
		t.Fatal("schema factories share mutable state")
	}
}

func TestInputSchemasNeverSerializeNullRequired(t *testing.T) {
	for _, name := range ToolNames() {
		schema, ok := InputSchema(name)
		if !ok {
			t.Fatalf("missing input schema for %s", name)
		}
		encoded, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("marshal %s input schema: %v", name, err)
		}
		if string(encoded) == "" || json.Valid(encoded) == false {
			t.Fatalf("%s input schema is not valid JSON: %s", name, encoded)
		}
		if required, exists := schema["required"]; exists && required == nil {
			t.Fatalf("%s input schema contains required:null: %s", name, encoded)
		}
	}

	for _, name := range []string{ToolAgentDockContext, ToolProjectList, ToolRecallMaintain} {
		schema, _ := InputSchema(name)
		if _, exists := schema["required"]; exists {
			t.Fatalf("%s should omit empty required", name)
		}
	}
}

func TestWorkflowRootIsStrictButTemplatePayloadIsOpen(t *testing.T) {
	schema, _ := InputSchema(ToolWorkflowTemplateManage)
	props := schema["properties"].(map[string]any)
	template := props["template"].(map[string]any)
	if schema["additionalProperties"] != false || template["additionalProperties"] != true {
		t.Fatalf("unexpected workflow strictness root=%#v template=%#v", schema["additionalProperties"], template["additionalProperties"])
	}
}

func TestProjectToolInputsDoNotAcceptTrustedRoutingOverrides(t *testing.T) {
	for _, name := range []string{ToolProjectOpen, ToolProjectContext} {
		schema, ok := InputSchema(name)
		if !ok {
			t.Fatalf("missing Project input schema for %s", name)
		}
		props := schema["properties"].(map[string]any)
		for _, forbidden := range []string{"node_id", "working_folder", "permissions", "deployment_revision", "context_revision"} {
			if _, exists := props[forbidden]; exists {
				t.Fatalf("%s exposes trusted routing override %q", name, forbidden)
			}
		}
	}

	open, _ := InputSchema(ToolProjectOpen)
	targets := open["properties"].(map[string]any)["targets"].(map[string]any)
	if targets["minItems"] != 1 {
		t.Fatalf("project_open targets minItems = %#v, want 1", targets["minItems"])
	}
	target := targets["items"].(map[string]any)
	if target["additionalProperties"] != false {
		t.Fatalf("project_open target selection must be strict: %#v", target)
	}
	targetProps := target["properties"].(map[string]any)
	for _, allowed := range []string{"deployment_id", "cwd_rel"} {
		if _, exists := targetProps[allowed]; !exists {
			t.Fatalf("project_open target selection missing %q", allowed)
		}
	}
}

func TestProjectOutputCarriesCompletePromptSources(t *testing.T) {
	for _, name := range []string{ToolProjectList, ToolProjectOpen, ToolProjectContext} {
		candidate, ok := OutputSchema(name)
		if !ok {
			t.Fatalf("missing %s output schema", name)
		}
		if candidate["additionalProperties"] != false {
			t.Fatalf("%s output must be strict: %#v", name, candidate)
		}
		if required, ok := candidate["required"].([]string); !ok || len(required) == 0 {
			t.Fatalf("%s output must declare required fields: %#v", name, candidate["required"])
		}
	}

	openSchema, _ := OutputSchema(ToolProjectOpen)
	openProps := openSchema["properties"].(map[string]any)
	for _, field := range []string{"context_revision", "delivery"} {
		if _, exists := openProps[field]; !exists {
			t.Fatalf("project_open output missing %q", field)
		}
	}
	openDelivery := openProps["delivery"].(map[string]any)
	if openDelivery["additionalProperties"] != false {
		t.Fatalf("project_open delivery must be strict: %#v", openDelivery)
	}

	schema, ok := OutputSchema(ToolProjectContext)
	if !ok {
		t.Fatal("missing project_context output schema")
	}
	contextProps := schema["properties"].(map[string]any)
	for _, field := range []string{"project", "deployment", "target", "delivery"} {
		if _, exists := contextProps[field]; !exists {
			t.Fatalf("project_context output missing %q", field)
		}
	}
	delivery := contextProps["delivery"].(map[string]any)
	deliveryProps := delivery["properties"].(map[string]any)
	for _, field := range []string{"status", "context_revision"} {
		if _, exists := deliveryProps[field]; !exists {
			t.Fatalf("Project delivery schema missing %q", field)
		}
	}
	target := contextProps["target"].(map[string]any)
	prompt := target["properties"].(map[string]any)["prompt"].(map[string]any)
	promptProps := prompt["properties"].(map[string]any)
	for _, field := range []string{"prompt_revision", "complete", "bytes", "sources"} {
		if _, exists := promptProps[field]; !exists {
			t.Fatalf("project prompt schema missing %q", field)
		}
	}
	source := promptProps["sources"].(map[string]any)["items"].(map[string]any)
	sourceProps := source["properties"].(map[string]any)
	for _, field := range []string{"path", "scope", "sha256", "bytes", "content"} {
		if _, exists := sourceProps[field]; !exists {
			t.Fatalf("project prompt source schema missing %q", field)
		}
	}
	provenance := target["properties"].(map[string]any)["source_provenance"].(map[string]any)
	provenanceProps := provenance["properties"].(map[string]any)
	for _, field := range []string{"kind", "repository_root", "head", "branch", "detached", "unborn", "dirty"} {
		if _, exists := provenanceProps[field]; !exists {
			t.Fatalf("source provenance schema missing %q", field)
		}
	}
}

func TestRecallFactsAcceptRuntimeCoercibleValues(t *testing.T) {
	schema, _ := InputSchema(ToolRecallWrite)
	facts := schema["properties"].(map[string]any)["facts"].(map[string]any)
	if facts["additionalProperties"] != true {
		t.Fatalf("facts must match runtime fmt.Sprint coercion: %#v", facts)
	}
}

func TestPrivateNoteMissingEncryptedIsStringArray(t *testing.T) {
	schema, _ := OutputSchema(ToolPrivateNoteManage)
	missing := schema["properties"].(map[string]any)["missing_encrypted"].(map[string]any)
	items := missing["items"].(map[string]any)
	if missing["type"] != "array" || items["type"] != "string" {
		t.Fatalf("missing_encrypted schema = %#v", missing)
	}
}

func TestContextHasExplicitLocalAndFleetProfiles(t *testing.T) {
	local := LocalAgentDockContextOutputSchema()
	fleet := FleetAgentDockContextOutputSchema()
	localProperties := local["properties"].(map[string]any)
	if _, ok := localProperties["skills"]; !ok {
		t.Fatal("local context is missing skills")
	}
	commonSkills, ok := localProperties["common_skills"].(map[string]any)
	if !ok {
		t.Fatal("local context is missing common_skills")
	}
	commonProperties := commonSkills["properties"].(map[string]any)
	for _, name := range []string{"root", "total", "truncated", "items"} {
		if _, ok := commonProperties[name]; !ok {
			t.Fatalf("common_skills is missing %s", name)
		}
	}
<<<<<<< HEAD
	for _, required := range local["required"].([]string) {
		if required == "common_skills" {
			t.Fatal("common_skills must remain optional for rolling compatibility with older AgentDock nodes")
		}
	}
	nodes := fleet["properties"].(map[string]any)["nodes"].(map[string]any)
	node := nodes["items"].(map[string]any)
	nodeContext := node["properties"].(map[string]any)["context"].(map[string]any)
=======
	requireStringMember(t, local["required"].([]string), "common_skills", "local context required")
	nodes := fleet["properties"].(map[string]any)["nodes"].(map[string]any)
	node := nodes["items"].(map[string]any)
	nodeProperties := node["properties"].(map[string]any)
	if _, ok := nodeProperties["capability_status"]; !ok {
		t.Fatal("fleet node is missing capability_status")
	}
	if _, ok := nodeProperties["guidance"]; ok {
		t.Fatal("legacy Node Guidance remains in the v4 fleet context")
	}
	shared := fleet["properties"].(map[string]any)["shared"].(map[string]any)
	if _, ok := shared["properties"].(map[string]any)["global_instructions"]; ok {
		t.Fatal("legacy Global Instructions metadata remains in the v4 fleet context")
	}
	nodeContext := nodeProperties["context"].(map[string]any)
>>>>>>> 0332bc6 (feat(project): add full access and optional project folder semantics)
	if _, ok := nodeContext["properties"].(map[string]any)["common_skills"]; !ok {
		t.Fatal("fleet node context is missing common_skills")
	}
	runtimeSchema, ok := localProperties["runtime"].(map[string]any)
	if !ok {
		t.Fatal("local context is missing runtime")
	}
	runtimeProperties := runtimeSchema["properties"].(map[string]any)
	for _, name := range []string{"version", "os", "arch", "agentdock_home", "agentdock_default_dir", "default_cwd", "path_model"} {
		if _, ok := runtimeProperties[name]; !ok {
			t.Fatalf("local runtime is missing %s", name)
		}
	}
	if _, ok := fleet["properties"].(map[string]any)["nodes"]; !ok {
		t.Fatal("fleet context is missing nodes")
	}
	dynamicMCP := localProperties["dynamic_mcp"].(map[string]any)
	dynamicItem := dynamicMCP["items"].(map[string]any)
	dynamicProperties := dynamicItem["properties"].(map[string]any)
	for _, name := range []string{"name", "description", "status", "tool_count", "last_error_code"} {
		if _, ok := dynamicProperties[name]; !ok {
			t.Fatalf("dynamic MCP context item is missing %s", name)
		}
	}
	if dynamicItem["additionalProperties"] != false {
		t.Fatalf("dynamic MCP context item must remain strict: %#v", dynamicItem)
	}
	if reflect.DeepEqual(local, fleet) {
		t.Fatal("local and fleet context profiles unexpectedly match")
	}
}

func TestRecallWriteBehaviorVectorsCoverSafetyBoundary(t *testing.T) {
	cases := RecallWriteBehaviorCases()
	seen := map[string]RecallWriteBehaviorCase{}
	for _, c := range cases {
		seen[c.Name] = c
		if c.DryRun && c.Expected == RecallWriteMutation {
			t.Fatalf("dry run case mutates: %#v", c)
		}
		wantsOverwrite := c.Target == "markdown" && (c.Action == "replace" || c.Action == "append" || c.Action == "patch" || c.Action == "update_fact")
		if c.OverwriteSemantic != wantsOverwrite {
			t.Fatalf("overwrite semantic mismatch for %q: got=%t want=%t", c.Name, c.OverwriteSemantic, wantsOverwrite)
		}
	}
	for _, name := range []string{
		"markdown inbox create unconfirmed mutates",
		"markdown protected create unconfirmed errors",
		"markdown replace unconfirmed previews",
		"markdown delete unconfirmed errors",
		"markdown plan previews",
		"card create confirmed dry run previews",
	} {
		if _, ok := seen[name]; !ok {
			t.Fatalf("missing behavior vector %q", name)
		}
	}
}
