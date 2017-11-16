package datastore_backup

import (
	"net/http"

	"github.com/sinmetal/ds2bq"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func init() {
	kindNames := []string{"Item"}
	queueName := "datastore-to-bq"
	path := "/tq/gcs/object-to-bq"

	f := func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)

		obj, err := ds2bq.DecodeGCSObject(r.Body)
		if err != nil {
			log.Errorf(c, "ds2bq: failed to decode request: %s", err)
			return
		}
		defer r.Body.Close()

		bucketName := appengine.AppID(c) + "-datastore-backup"
		if !obj.IsImportTarget(c, r, bucketName, kindNames) {
			return
		}

		err = ds2bq.ReceiveOCN(c, obj, queueName, path)
		if err != nil {
			log.Errorf(c, "ds2bq: failed to receive OCN: %s", err)
			return
		}
	}

	http.HandleFunc("/cloud-datastore/gcs/object-change-notification", f) // from GCS, This API must not requires admin role.
	http.HandleFunc("/tq/gcs/object-to-bq", ds2bq.ImportBigQueryHandleFunc("datastore_imports"))
}
