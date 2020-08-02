package collector

import "github.com/prometheus/client_golang/prometheus"

const (
	metricChannel    = "channel"
	metricUserHandle = "userhandle"
)

type DiscordCollector struct {
	totalMessageCounter *prometheus.CounterVec
	totalBotUsage       *prometheus.CounterVec
}

func New() *DiscordCollector {
	return &DiscordCollector{
		totalMessageCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "discord",
			Name:      "message_total",
			Help:      "tracks the messages in the channels",
		}, []string{metricChannel, metricUserHandle}),
		totalBotUsage: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "discord",
			Name:      "bot_usage_total",
			Help:      "tracks the usage of the bot",
		}, []string{metricChannel, metricUserHandle}),
	}
}

func (d *DiscordCollector) Describe(descs chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(d, descs)
}

func (d *DiscordCollector) Collect(ch chan<- prometheus.Metric) {
	d.totalMessageCounter.Collect(ch)
	d.totalBotUsage.Collect(ch)
}

func (d *DiscordCollector) TrackMessage(channel, userHandle string) {
	d.totalMessageCounter.With(prometheus.Labels{
		metricChannel:    channel,
		metricUserHandle: userHandle,
	}).Inc()
}

func (d *DiscordCollector) TrackBotUsage(channel, userHandle string) {
	d.totalBotUsage.With(prometheus.Labels{
		metricChannel:    channel,
		metricUserHandle: userHandle,
	}).Inc()
}
