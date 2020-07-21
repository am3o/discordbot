package collector

import "github.com/prometheus/client_golang/prometheus"

const (
	Channel    = "channel"
	UserHandel = "userhandel"
)

type DiscordCollector struct {
	totalMessageCounter *prometheus.CounterVec
}

func New() *DiscordCollector {
	return &DiscordCollector{
		totalMessageCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "discord",
			Name:      "message_counter",
			Help:      "tracks the messages in the channels",
		}, []string{Channel, UserHandel}),
	}
}

func (d *DiscordCollector) Describe(descs chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(d, descs)
}

func (d *DiscordCollector) Collect(metrics chan<- prometheus.Metric) {
	d.totalMessageCounter.Collect(metrics)
}

func (d *DiscordCollector) TrackMessage(channel, user string) {
	d.totalMessageCounter.With(prometheus.Labels{
		Channel:    channel,
		UserHandel: user,
	}).Inc()
}
