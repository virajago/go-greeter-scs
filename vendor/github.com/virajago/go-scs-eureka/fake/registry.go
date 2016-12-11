package fake

import (
	"log"
	"net/http"
	"os"

	"github.com/emicklei/go-restful"

	"github.com/st3v/go-eureka"
)

type registry struct {
	apps map[string]*eureka.App
}

func NewRegistry() *registry {
	return &registry{
		apps: map[string]*eureka.App{},
	}
}

func (r *registry) HTTPServer(addr string, debug bool) *http.Server {
	if debug {
		restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))
	}

	s := new(restful.WebService)

	s.Path("/").Produces(restful.MIME_XML)
	s.Route(s.POST("/apps/{app-name}").To(r.register).Consumes(restful.MIME_XML))
	s.Route(s.DELETE("/apps/{app-name}/{instance-id}").To(r.deregister))
	s.Route(s.PUT("/apps/{app-name}/{instance-id}").To(r.heartbeat))
	s.Route(s.GET("/apps").To(r.list))
	s.Route(s.GET("/apps/{app-name}").To(r.app))
	s.Route(s.GET("/apps/{app-name}/{instance-id}").To(r.appInstance))
	s.Route(s.PUT("/apps/{app-name}/{instance-id}/status").To(r.statusOverride))
	s.Route(s.DELETE("/apps/{app-name}/{instance-id}/status").To(r.removeStatusOverride))
	s.Route(s.GET("/instances/{instance-id}").To(r.instance))

	return &http.Server{
		Addr:    addr,
		Handler: restful.NewContainer().Add(s),
	}
}

func (r *registry) deregister(req *restful.Request, resp *restful.Response) {
	resp.AddHeader("Content-Type", "text/plain")

	name := req.PathParameter("app-name")
	instanceID := req.PathParameter("instance-id")

	if app, found := r.apps[name]; found {
		for i, instance := range app.Instances {
			if instance.ID == instanceID {
				app.Instances = append(app.Instances[0:i], app.Instances[i+1:]...)

				if len(app.Instances) == 0 {
					delete(r.apps, name)
				}

				resp.WriteHeader(http.StatusOK)
				return
			}
		}
	}

	resp.WriteErrorString(http.StatusNotFound, "Instance not found.")
}

func (r *registry) register(req *restful.Request, resp *restful.Response) {
	resp.AddHeader("Content-Type", "text/plain")

	name := req.PathParameter("app-name")

	instance := new(eureka.Instance)
	err := req.ReadEntity(instance)
	if err != nil {
		resp.WriteHeader(http.StatusNotAcceptable)
		return
	}

	app, found := r.apps[name]
	if !found {
		app = &eureka.App{
			Name:      name,
			Instances: make([]*eureka.Instance, 0, 1),
		}
	}

	for _, i := range app.Instances {
		if i.ID == instance.ID {
			resp.WriteErrorString(http.StatusMethodNotAllowed, "Instance already registered")
			return
		}
	}

	app.Instances = append(app.Instances, instance)

	r.apps[name] = app
	resp.WriteHeader(http.StatusNoContent)
}

func (r *registry) list(req *restful.Request, resp *restful.Response) {
	apps := make([]*eureka.App, 0, len(r.apps))

	for _, app := range r.apps {
		apps = append(apps, app)
	}

	result := eureka.AppsResponse{
		Apps: apps,
	}

	resp.WriteEntity(result)
}

func (r *registry) app(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("app-name")

	app, found := r.apps[name]
	if !found {
		resp.AddHeader("Content-Type", "text/plain")
		resp.WriteErrorString(http.StatusNotFound, "App not found.")
		return
	}

	resp.WriteEntity(app)
}

func (r *registry) heartbeat(req *restful.Request, resp *restful.Response) {
	resp.AddHeader("Content-Type", "text/plain")

	name := req.PathParameter("app-name")
	instanceID := req.PathParameter("instance-id")

	if _, found := r.findAppInstance(name, instanceID); !found {
		resp.WriteErrorString(http.StatusNotFound, "Instance not found.")
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (r *registry) instance(req *restful.Request, resp *restful.Response) {
	instanceID := req.PathParameter("instance-id")

	if i, found := findInstance(instanceID, r.apps); found {
		resp.WriteEntity(i)
		return
	}

	resp.AddHeader("Content-Type", "text/plain")
	resp.WriteErrorString(http.StatusNotFound, "Instance not found.")
}

func (r *registry) appInstance(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("app-name")
	instanceID := req.PathParameter("instance-id")

	if i, found := r.findAppInstance(name, instanceID); found {
		resp.WriteEntity(i)
		return
	}

	resp.AddHeader("Content-Type", "text/plain")
	resp.WriteErrorString(http.StatusNotFound, "Instance not found.")
}

func (r *registry) statusOverride(req *restful.Request, resp *restful.Response) {
	value := req.QueryParameter("value")
	status, err := eureka.ParseStatus(value)
	if value == "" || err != nil {
		resp.WriteErrorString(http.StatusNotAcceptable, "Invalid status")
		return
	}

	name := req.PathParameter("app-name")
	instanceID := req.PathParameter("instance-id")

	instance, found := r.findAppInstance(name, instanceID)
	if !found {
		resp.WriteErrorString(http.StatusNotFound, "Instance not registered")
		return
	}

	instance.Status = status
	instance.StatusOverride = status
}

func (r *registry) removeStatusOverride(req *restful.Request, resp *restful.Response) {
	value := req.QueryParameter("value")

	status, err := eureka.ParseStatus(value)
	if err != nil {
		resp.WriteErrorString(http.StatusNotAcceptable, "Invalid status")
		return
	}

	name := req.PathParameter("app-name")
	instanceID := req.PathParameter("instance-id")

	instance, found := r.findAppInstance(name, instanceID)
	if !found {
		resp.WriteErrorString(http.StatusNotFound, "Instance not registered")
		return
	}

	instance.Status = status
	instance.StatusOverride = eureka.StatusUnknown
}

func (r *registry) findAppInstance(appName, instanceID string) (*eureka.Instance, bool) {
	if app, found := r.apps[appName]; found {
		return findInstance(instanceID, map[string]*eureka.App{app.Name: app})
	}

	return nil, false
}

func findInstance(instanceID string, apps map[string]*eureka.App) (*eureka.Instance, bool) {
	for _, a := range apps {
		for _, i := range a.Instances {
			if i.ID == instanceID {
				return i, true
			}
		}
	}

	return nil, false
}
