package middleware

import (
	"myapp/data"

	"github.com/leetrent/celeritas"
)

type Middleware struct {
	App    *celeritas.Celeritas
	Models data.Models
}
