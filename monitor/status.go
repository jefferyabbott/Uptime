package monitor

import (
	"context"
	"time"
)

type SiteStatus struct {
	Up        bool      `json:"up"`
	CheckedAt time.Time `json:"checked_at"`
}

type StatusResponse struct {
	Sites map[int]SiteStatus `json:"sites"`
}

//encore:api public method=GET path=/status
func Status(ctx context.Context) (*StatusResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT DISTINCT ON (site_id) site_id, up, checked_at
		FROM checks
		ORDER BY site_id, checked_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]SiteStatus)
	for rows.Next() {
		var siteID int
		var status SiteStatus
		if err := rows.Scan(&siteID, &status.Up, &status.CheckedAt); err != nil {
			return nil, err
		}
		result[siteID] = status
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &StatusResponse{Sites: result}, nil
}
