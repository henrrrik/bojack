package view

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/bojack/internal/msg"
	"github.com/henrrrik/bojack/internal/runway"
)

// NewTemplateListModel lists app templates.
func NewTemplateListModel(styles DetailStyles, width, height int) ListModel[runway.Template] {
	return NewListModel("templates", []table.Column{
		{Title: "Name", Width: 16},
		{Title: "Category", Width: 12},
		{Title: "Latest", Width: 8},
		{Title: "Min Plan", Width: 16},
		{Title: "Description", Width: 50},
	}, templateRow, styles, width, height).WithSelect(func(t runway.Template) tea.Msg {
		return msg.NavigateMsg{View: msg.ViewTemplateDetail, Param: t.Name}
	})
}

func templateRow(t runway.Template) table.Row {
	return table.Row{t.Name, t.Category, t.Latest, t.MinPlan, t.Description}
}

// NewRegistryListModel lists container registries.
func NewRegistryListModel(styles DetailStyles, width, height int) ListModel[runway.Registry] {
	return NewListModel("registries", []table.Column{
		{Title: "Hostname", Width: 60},
		{Title: "Organization", Width: 28},
		{Title: "Username", Width: 20},
		{Title: "Managed", Width: 8},
	}, registryRow, styles, width, height)
}

func registryRow(r runway.Registry) table.Row {
	managed := "no"
	if r.Managed {
		managed = "yes"
	}
	return table.Row{r.Hostname, r.Organization, orDash(r.Username), managed}
}

// NewSSHKeyListModel lists SSH deploy keys.
func NewSSHKeyListModel(styles DetailStyles, width, height int) ListModel[runway.SSHKey] {
	return NewListModel("SSH keys", []table.Column{
		{Title: "ID", Width: 36},
		{Title: "Type", Width: 20},
		{Title: "Fingerprint", Width: 51},
		{Title: "Created", Width: 10},
	}, sshKeyRow, styles, width, height)
}

func sshKeyRow(k runway.SSHKey) table.Row {
	keyType := "-"
	if f := strings.Fields(k.Public); len(f) > 0 {
		keyType = f[0]
	}
	return table.Row{k.ID, keyType, sshFingerprint(k.Public), formatDate(k.Created)}
}

// sshFingerprint returns the key's fingerprint as "ssh-keygen -l" prints it
// ("SHA256:<unpadded base64>"), or "-" if public is not a valid key line.
func sshFingerprint(public string) string {
	f := strings.Fields(public)
	if len(f) < 2 {
		return "-"
	}
	blob, err := base64.StdEncoding.DecodeString(f[1])
	if err != nil {
		return "-"
	}
	sum := sha256.Sum256(blob)
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(sum[:])
}

// NewAPIKeyListModel lists API keys.
func NewAPIKeyListModel(styles DetailStyles, width, height int) ListModel[runway.APIKey] {
	return NewListModel("API keys", []table.Column{
		{Title: "Name", Width: 24},
		{Title: "Status", Width: 10},
		{Title: "Created", Width: 10},
		{Title: "Expires", Width: 10},
		{Title: "Last Used", Width: 10},
	}, apiKeyRow, styles, width, height)
}

func apiKeyRow(k runway.APIKey) table.Row {
	return table.Row{k.Name, k.Status, formatDate(k.Created), formatDate(k.Expires), formatDate(k.LastUsed)}
}

// formatDate shortens an RFC 3339 timestamp to its date, or "-" if empty.
func formatDate(ts string) string {
	if len(ts) < 10 {
		return orDash(ts)
	}
	return ts[:10]
}
