# grafana-image-renderer-sdk-go

> This package is in early alpha. Please use with caution.

This package can be used to interact with the [Grafana Image Renderer](https://github.com/grafana/grafana-image-renderer) from Go applications.

It also includes:

- a sequencer package to make rendering sets of images easier
- a CLI for both image and sequence rendering

This package contains code & general language from [grafana-tools/sdk](https://github.com/grafana-tools/sdk).

<a href="#"><img src="docs/img/banner.gif"></a>

## CLI

Install with:

```text
go get -u github.com/MacroPower/grafana-image-renderer-sdk-go/cmd/grafana-image-renderer-cli
```

The CLI contains two subcommands, `image` for single renders, and `sequence` for consecutive renders.

```text
$ grafana-image-renderer-cli image --help

Usage of image:
  -api-key-or-basic-auth string
        Grafana authorization, either an API key or basic auth
  -api-url string
        Grafana API URL
  -dashboard string
        ID of the dashboard
  -end-time int
        The ending timestamp (Unix MS) of the render
  -height int
        The height of the image (default 1080)
  -out-file string
        The file to write (default "img.png")
  -panel int
        ID of the panel, 0 = Entire dashboard
  -start-time int
        The starting timestamp (Unix MS) of the render
  -timeout duration
        Timeout of the render request (default 1m0s)
  -width int
        The width of the image (default 1920)
```

```text
$ grafana-image-renderer-cli sequence --help

Usage of sequence:
  -api-key-or-basic-auth string
        Grafana authorization, either an API key or basic auth
  -api-url string
        Grafana API URL
  -dashboard string
        ID of the dashboard
  -end-padding duration
        Duration to add to the end of the frame
  -frame-interval duration
        Time progression between frames, positive = forward, negative = backward (default 5m0s)
  -frames string
        The frames to render, pass a range and/or a set (e.g. 1-10,12,15) (default "1-2")
  -height int
        The height of the image (default 1080)
  -max-concurrency int
        Maximum number of concurrent render requests (default 5)
  -out-directory string
        Directory to write rendered frames to (default "frames")
  -panel int
        ID of the panel, 0 = Entire dashboard
  -start-padding duration
        Duration to add to the start of the frame
  -start-time int
        The starting timestamp (Unix MS) of the render
  -timeout duration
        Timeout of the render request (default 1m0s)
  -width int
        The width of the image (default 1920)
  -worker-delay duration
        Delay worker startup (default 2s)
```
