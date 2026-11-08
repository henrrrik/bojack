package view

import (
	"strconv"
	"strings"

	"github.com/henrrrik/bojack/internal/runway"
)

// NewTemplateDetailModel shows one app template.
func NewTemplateDetailModel(name string, styles DetailStyles, width, height int) ItemModel[runway.Template] {
	return NewItemModel("template", name, templateFields, styles, width, height)
}

func templateFields(t runway.Template) []field {
	return []field{
		{"Name:", t.Name},
		{"Display name:", orDash(t.DisplayName)},
		{"Category:", orDash(t.Category)},
		{"Maintainer:", orDash(t.Maintainer)},
		{"Min plan:", orDash(t.MinPlan)},
		{"Latest:", orDash(t.Latest)},
		{"Versions:", orDash(strings.Join(t.Versions, ", "))},
		{"Description:", orDash(t.Description)},
	}
}

// NewVolumeDetailModel shows one volume.
func NewVolumeDetailModel(name string, styles DetailStyles, width, height int) ItemModel[runway.Volume] {
	return NewItemModel("volume", name, volumeFields, styles, width, height).WithActions(volumeActions)
}

// volumeActions offers resizing when the volume allows it.
func volumeActions(v runway.Volume) []actionSpec {
	if !v.Resizable {
		return nil
	}
	return []actionSpec{resizeVolumeAction(v.SizeMB)}
}

func volumeFields(v runway.Volume) []field {
	return []field{
		{"Name:", v.Name},
		{"App:", orDash(v.App)},
		{"Mountpoint:", orDash(v.Mountpoint)},
		{"Size:", formatMB(strconv.Itoa(v.SizeMB))},
		{"Extra:", extraMB(v.SizeExtraMB)},
		{"Replicas:", strconv.Itoa(len(v.Replicas))},
		{"Is replica:", yesNo(v.IsReplica)},
		{"Attachable:", yesNo(v.Attachable)},
		{"Detachable:", yesNo(v.Detachable)},
		{"Deletable:", yesNo(v.Deletable)},
		{"Resizable:", yesNo(v.Resizable)},
	}
}

// extraMB formats extra volume storage, "-" when there is none.
func extraMB(mb int) string {
	if mb <= 0 {
		return "-"
	}
	return formatMB(strconv.Itoa(mb))
}
