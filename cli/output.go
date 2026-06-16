package cli

import (
	"io"

	render "github.com/tamnd/thegradient-cli/pkg"
)

type Format = render.Format

const (
	FormatTable = render.FormatTable
	FormatJSON  = render.FormatJSON
	FormatJSONL = render.FormatJSONL
	FormatCSV   = render.FormatCSV
	FormatTSV   = render.FormatTSV
	FormatURL   = render.FormatURL
	FormatRaw   = render.FormatRaw
)

func NewRenderer(w io.Writer, format Format, fields []string, noHeader bool, tmpl string) *render.Renderer {
	return render.New(w, format, fields, noHeader, tmpl)
}

// newRendererTo builds a renderer writing to w using the App's current settings.
func (a *App) newRendererTo(w io.Writer) *render.Renderer {
	format := Format(a.output)
	if !format.Valid() {
		format = FormatJSONL
	}
	return NewRenderer(w, format, a.fields, a.noHeader, a.template)
}
