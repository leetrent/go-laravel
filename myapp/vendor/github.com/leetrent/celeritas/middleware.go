package celeritas

import "net/http"

func (c *Celeritas) SessionLoad(next http.Handler) http.Handler {
	//c.InfoLog.Println("[celeritas][middleware][SessionLoad] =>")
	return c.Session.LoadAndSave(next)
}
