package actions

import (
	"simplecms/public"
	"simplecms/templates"

	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/plush/v4"
	"errors"
	"html/template"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.plush.html",

		// fs.FS containing templates
		TemplatesFS: templates.FS(),

		// fs.FS containing assets
		AssetsFS: public.FS(),

		// Add template helpers here:
		Helpers: render.Helpers{
			// for non-bootstrap form helpers uncomment the lines
			// below and import "github.com/gobuffalo/helpers/forms"
			// forms.FormKey:     forms.Form,
			// forms.FormForKey:  forms.FormFor,
			
			// Add CSRF helper
			"csrf": func(ctx plush.HelperContext) (template.HTML, error) {
				tok, ok := ctx.Value("authenticity_token").(string)
				if !ok {
					return "", errors.New("CSRF token not found in context")
				}
				return template.HTML(`<input type="hidden" name="authenticity_token" value="` + tok + `" />`), nil
			},
		},
	})
}
