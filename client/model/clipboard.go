package model

type Clipboard struct {
	Op            string   `json:"op"`
	ClipboardType string   `json:"clipboard_type,omitempty"`
	ClipboardData []string `json:"clipboard_data,omitempty"`
}
