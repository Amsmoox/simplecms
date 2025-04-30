package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// AdminDashboardHandler serves the admin dashboard
func AdminDashboardHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("admin/dashboard.plush.html"))
} 