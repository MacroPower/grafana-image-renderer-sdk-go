package client

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"time"
)

// Render renders a panel. If panelId is set to 0, the entire dashboard is rendered.
func (r *Client) Render(ctx context.Context, dashboard string, from, to time.Time, panelId, width, height int) ([]byte, int, error) {
	params := url.Values{}
	params.Set("width", fmt.Sprint(width))
	params.Set("height", fmt.Sprint(height))
	params.Set("from", fmt.Sprint(from.UnixNano()/int64(time.Millisecond)))
	params.Set("to", fmt.Sprint(to.UnixNano()/int64(time.Millisecond)))

	t := "d"
	if panelId != 0 {
		params.Set("panelId", fmt.Sprint(panelId))
		t = "d-solo"
	}

	deadline, ok := ctx.Deadline()
	if ok {
		timeout := time.Until(deadline)
		params.Set("timeout", fmt.Sprint(int(timeout.Seconds())))
	}

	b, code, err := r.Get(ctx, path.Join("render", t, dashboard), params)
	if err != nil {
		return nil, code, err
	}

	return b, code, nil
}
