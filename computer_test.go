package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestComputerPermissionsExplicitAndValidated(t *testing.T) {
	for _, permission := range []ComputerPermission{ComputerPermissionNone, ComputerPermissionObserve, ComputerPermissionControl} {
		p := DeploymentPermissions{Files: FileCapabilityNone, Computer: permission}
		if err := p.Validate(); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		var decoded DeploymentPermissions
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded != p {
			t.Fatalf("permissions changed: %s", encoded)
		}
	}
	for _, p := range []DeploymentPermissions{{Files: FileCapabilityNone}, {Files: FileCapabilityNone, Computer: "root"}, {Files: "root", Computer: ComputerPermissionControl}, {FullAccess: true, Files: FileCapabilityNone, Computer: "invalid"}} {
		if p.Validate() == nil {
			t.Fatalf("invalid permissions accepted: %+v", p)
		}
	}
}
func TestComputerPermissionMatrix(t *testing.T) {
	for _, permission := range []ComputerPermission{ComputerPermissionNone, ComputerPermissionObserve, ComputerPermissionControl} {
		for _, full := range []bool{false, true} {
			p := DeploymentPermissions{Files: FileCapabilityNone, Computer: permission, FullAccess: full}
			for _, tool := range []string{ToolComputerStatus, ToolComputerStop, ToolComputerObserve, ToolComputerSession, ToolComputerAct, "computer_shell"} {
				want := tool == ToolComputerStatus || tool == ToolComputerStop || (IsComputerTool(tool) && (full || permission == ComputerPermissionControl || (permission == ComputerPermissionObserve && tool == ToolComputerObserve)))
				if got := p.AllowsComputerTool(tool); got != want {
					t.Fatalf("%+v %s: got %v want %v", p, tool, got, want)
				}
			}
		}
	}
}
func TestHistoricalComputerControlIsMetadataOrStopOnly(t *testing.T) {
	for _, tc := range []struct {
		name string
		args map[string]any
		want bool
	}{
		{ToolComputerStop, map[string]any{"session_id": "cs-1"}, true},
		{ToolComputerStatus, map[string]any{"operation_id": "op-1"}, true},
		{ToolComputerStatus, nil, false}, {ToolComputerStatus, map[string]any{"operation_id": "  "}, false},
		{ToolComputerStatus, map[string]any{"operation_id": 1}, false},
		{ToolComputerObserve, map[string]any{"operation_id": "op-1"}, false},
		{ToolComputerSession, map[string]any{"action": "release"}, false}, {ToolComputerAct, nil, false},
	} {
		if got := AllowsHistoricalComputerControl(tc.name, tc.args); got != tc.want {
			t.Fatalf("%+v got %v", tc, got)
		}
	}
}

func TestComputerScrollWirePreservesZeroAxes(t *testing.T) {
	zero, delta := 0, 120
	action := ComputerAction{Kind: "scroll", Point: &ComputerPoint{}, DeltaX: &zero, DeltaY: &delta}
	encoded, err := json.Marshal(action)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"delta_x":0`) || !strings.Contains(string(encoded), `"delta_y":120`) {
		t.Fatalf("scroll lost zero axis: %s", encoded)
	}
}
