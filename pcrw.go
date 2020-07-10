package pcrw

import (
	"io/ioutil"
	"net/url"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/prometheus/prometheus/tsdb"
)

func Push(logger log.Logger, registerer prometheus.Registerer, gatherer prometheus.Gatherer, interval time.Duration, remoteURL string) error {
	tempDir, err := ioutil.TempDir("", "pcrw")
	if err != nil {
		return err
	}

	localStorage, err := tsdb.Open(tempDir, logger, registerer, &tsdb.Options{})
	if err != nil {
		return err
	}

	remoteStorage := remote.NewStorage(
		logger,
		registerer,
		func() (int64, error) {
			return 0, nil
		},
		tempDir,
		time.Minute,
	)

	remote, err := url.Parse(remoteURL)
	if err != nil {
		return err
	}
	if err := remoteStorage.ApplyConfig(&config.Config{
		RemoteWriteConfigs: []*config.RemoteWriteConfig{
			&config.RemoteWriteConfig{
				URL: &config_util.URL{
					remote,
				},
				RemoteTimeout: model.Duration(30 * time.Second),
				QueueConfig:   config.DefaultQueueConfig,
			},
		},
	}); err != nil {
		return err
	}

	fanoutStorage := storage.NewFanout(logger, localStorage, remoteStorage)

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			logger.Log("msg", "pushing metrics via remote_write")

			fams, err := gatherer.Gather()
			if err != nil {
				logger.Log("err", err)
				continue
			}

			decodeOptions := &expfmt.DecodeOptions{
				Timestamp: model.Now(),
			}
			vector, err := expfmt.ExtractSamples(decodeOptions, fams...)
			if err != nil {
				logger.Log("err", err)
				continue
			}

			appender := fanoutStorage.Appender()

			for _, sample := range vector {
				if _, err := appender.Add(metricToLabel(sample.Metric), int64(sample.Timestamp), float64(sample.Value)); err != nil {
					logger.Log("err", err)
					continue
				}
			}

			if err := appender.Commit(); err != nil {
				logger.Log("err", err)
				continue
			}
		}
	}

	return nil
}

func metricToLabel(m model.Metric) labels.Labels {
	l := []labels.Label{}

	for name, value := range m {
		l = append(l, labels.Label{
			Name:  string(name),
			Value: string(value),
		})
	}

	return l
}
