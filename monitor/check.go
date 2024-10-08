package monitor

import (
	"context"

	"encore.app/site"
	"encore.dev/cron"
	"encore.dev/storage/sqldb"
	"golang.org/x/sync/errgroup"
)

//encore:api public method=POST path=/check/:siteID
func Check(ctx context.Context, siteID int) error {
	site, err := site.Get(ctx, siteID)
	if err != nil {
		return err
	}
	return check(ctx, site)
}

//encore:api public method=POST path=/checkall
func CheckAll(ctx context.Context) error {
	// get the tracked sites
	resp, err := site.List(ctx)
	if err != nil {
		return err
	}

	// check up to 8 sites concurrently
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	for _, site := range resp.Sites {
		site := site
		g.Go(func() error {
			return check(ctx, site)
		})
	}
	return g.Wait()
}

var db = sqldb.NewDatabase("monitor", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

func check(ctx context.Context, site *site.Site) error {
	result, err := Ping(ctx, site.URL)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `
		INSERT INTO checks (site_id, up, checked_at)
		VALUES ($1, $2, NOW())
	`, site.ID, result.Up)
	return err
}

// cron job to check every hour
var _ = cron.NewJob("check-all", cron.JobConfig{
	Title:    "Check all sites",
	Endpoint: CheckAll,
	Every:    1 * cron.Hour,
})
