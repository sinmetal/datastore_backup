package datastore_backup

import (
	"context"
	"net/http"

	"github.com/sinmetal/ds2bq"

	"github.com/favclip/ucon"
	"github.com/favclip/ucon/swagger"

	"google.golang.org/appengine"
)

func UseAppengineContext(b *ucon.Bubble) error {
	b.Context = appengine.NewContext(b.R)
	return b.Next()
}

func init() {
	ucon.Middleware(UseAppengineContext)
	ucon.Orthodox()

	swPlugin := swagger.NewPlugin(&swagger.Options{
		Object: &swagger.Object{
			Info: &swagger.Info{
				Title:   "ds2bq",
				Version: "1",
			},
		},
	})
	ucon.Plugin(swPlugin)

	{
		s, err := ds2bq.NewGCSWatcherService(
			ds2bq.GCSWatcherWithURLs(
				"/cloud-datastore/gcs/object-change-notification",
				"/tq/gcs/object-to-bq",
			),
			ds2bq.GCSWatcherWithAfterContext(func(c context.Context) (ds2bq.GCSWatcherOption, error) {
				bucketName := appengine.AppID(c) + "-datastore-backup"
				return ds2bq.GCSWatcherWithBackupBucketName(bucketName), nil
			}),
			ds2bq.GCSWatcherWithDatasetID("datastore_imports"),
			ds2bq.GCSWatcherWithQueueName("datastore-to-bq"),
			ds2bq.GCSWatcherWithTargetKindNames("Item"),
		)
		if err != nil {
			panic(err)
		}
		s.SetupWithUcon()
	}

	ucon.DefaultMux.Prepare()
	http.Handle("/", ucon.DefaultMux)
}
