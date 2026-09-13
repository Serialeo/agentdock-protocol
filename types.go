package protocol

import "encoding/json"

// Message is the AgentDock ↔ NexusDock WebSocket envelope.
type Message struct {
	Type             string            `json:"type"`
	RequestID        string            `json:"request_id,omitempty"`
	Operation        string            `json:"operation,omitempty"`
	ExecutionContext *ExecutionContext `json:"execution_context,omitempty"`
	Arguments        json.RawMessage   `json:"arguments,omitempty"`
	Result           json.RawMessage   `json:"result,omitempty"`
	Error            *RemoteError      `json:"error,omitempty"`
	Hello            *Hello            `json:"hello,omitempty"`
	ProtocolVersion  string            `json:"protocol_version,omitempty"`
	HeartbeatMS      int               `json:"heartbeat_ms,omitempty"`
}

// Hello is the complete node capability snapshot sent at connection time.
type Hello struct {
	DeviceID           string                 `json:"device_id"`
	Version            string                 `json:"version"`
	ProtocolVersion    string                 `json:"protocol_version"`
	OS                 string                 `json:"os"`
	Arch               string                 `json:"arch"`
	Capabilities       []string               `json:"capabilities"`
	BridgeCapabilities []string               `json:"bridge_capabilities,omitempty"`
	ToolContractHash   string                 `json:"tool_contract_hash"`
	Tools              []ToolDescriptor       `json:"tools"`
	UIResources        []UIResourceCapability `json:"ui_resources"`
}

// ToolDescriptor is the Bridge projection of an MCP tool descriptor.
// UI resource capability is deliberately absent: per-tool _meta.ui is presentation binding only.
type ToolDescriptor struct {
	Name         string         `json:"name"`
	Title        string         `json:"title,omitempty"`
	Description  string         `json:"description,omitempty"`
	InputSchema  map[string]any `json:"inputSchema"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
	Meta         map[string]any `json:"_meta,omitempty"`
	Annotations  map[string]any `json:"annotations,omitempty"`
}

// UIResourceCapability advertises a resource the node can actually serve through resource.read.
type UIResourceCapability struct {
	URI      string `json:"uri"`
	Contract string `json:"contract"`
	MIMEType string `json:"mime_type"`
}

type RemoteError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Category  string         `json:"category,omitempty"`
	Retryable bool           `json:"retryable,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

func (e *RemoteError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}
