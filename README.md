# pcrw

`prometheus-client-remote-write`

A prototype of a Prometheus client using remote write functionality to 'push' metrics directly to a remote write server.

Don't use this. It's dumb. Barely a POC.

## 'Features'

- Don't need to run a Prometheus server!
- Just add one line, and push metrics to a remote_write compatible server (e.g: Cortex)!

Basically, it's possibly useful if you're in an environment where you can't / don't want to run a Prometheus server, but want to push metrics somewhere.

## Demo

### Starting a remote_write server

Build the `example_write_adapter` (see [here](https://github.com/prometheus/prometheus/blob/master/documentation/examples/remote_storage/example_write_adapter/server.go)) to use as a remote write server.

It is recommended to start the `example_write_adapter` as so:

```
./example_write_adapter | grep -A 1 pcrw_example_gauge
```
This helps to clean up the output somewhat by removing the other metrics.

The demo assumes the adapter is available at `http://localhost:1234/receive`. See [here](https://www.robustperception.io/using-the-remote-write-path) for more on the `example_write_adapter`.

### Starting a push client

To run the demo, execute

```
go run demo/demo.go
```

This will start the client. You should see output similar to the following from the server within a minute or so (default 'push' interval is 30 seconds):

```
$ ./example_write_adapter | grep -A 1 pcrw_example_gauge
pcrw_example_gauge
  46.000000 1594394926487
--
pcrw_example_gauge
  16.000000 1594394956488
```

This shows the client 'scraping' the metrics (with `pcrw_example_gauge` being set to the value of seconds when scraped) and pushing them to the server.

## Memory Usage

With pcrw, the demo uses `54440` RSS (~55Mb), compared to `816` (~1Mb) when run without.