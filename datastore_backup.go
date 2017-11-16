package datastore_backup

import (
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"

	datastore "google.golang.org/api/datastore/v1beta1"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/log"
)

func init() {
	http.HandleFunc("/cloud-datastore-export", Export)
}

func Export(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	for k, v := range r.Header {
		log.Infof(ctx, "%s = %s", k, v)
	}

	if r.Header.Get("X-Appengine-Cron") != "true" {
		if user.IsAdmin(ctx) == false {
			l, err := user.LoginURL(ctx, r.URL.Path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, l, http.StatusFound)
			return
		}
	}

	outputStoragePath := r.FormValue("outputStoragePath")
	kind := r.FormValue("kind")

	ctxWithDeadline, _ := context.WithTimeout(ctx, 10*time.Minute)
	client, err := google.DefaultClient(ctxWithDeadline, datastore.DatastoreScope)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	service, err := datastore.New(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := appengine.AppID(ctx)
	op, err := service.Projects.Export(p, &datastore.GoogleDatastoreAdminV1beta1ExportEntitiesRequest{
		EntityFilter: &datastore.GoogleDatastoreAdminV1beta1EntityFilter{
			NamespaceIds: []string{},
			Kinds:        []string{kind},
		},
		OutputUrlPrefix: outputStoragePath,
	}).Do()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(op.HTTPStatusCode)
	b, err := op.Response.MarshalJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
