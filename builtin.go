package protocol

// BuiltinCapability describes a node-owned optional tool group, never an external MCP server.
// Available is derived by the node from Provided, Enabled, Ready and Transitioning.
type BuiltinCapability struct {
	ID            string   `json:"id"`
	Provided      bool     `json:"provided"`
	Enabled       bool     `json:"enabled"`
	Ready         bool     `json:"ready"`
	Available     bool     `json:"available"`
	Transitioning bool     `json:"transitioning"`
	Reason        string   `json:"reason"`
	Tools         []string `json:"tools"`
}

type BuiltinUpdate struct {
	ID      string `json:"id"`
	Enabled *bool  `json:"enabled"`
}

const MessageNodeUpdated = "node.updated"
