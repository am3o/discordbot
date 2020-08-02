package collector

import "github.com/prometheus/client_golang/prometheus"

const (
	channel    = "channel"
	userHandle = "userhandel"
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
		}, []string{channel, userHandle}),
		totalBotUsage: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "discord",
			Name:      "bot_usage_total",
			Help:      "tracks the usage of the bot",
		}, []string{channel, userHandle}),
	}
}

func (d *DiscordCollector) Describe(descs chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(d, descs)
}

func (d *DiscordCollector) Collect(ch chan<- prometheus.Metric) {
	d.totalMessageCounter.Collect(ch)
	d.totalBotUsage.Collect(ch)
}

func (d *DiscordCollector) TrackMessage(channel, user string) {
	d.totalMessageCounter.With(prometheus.Labels{
		channel:    channel,
		userHandle: user,
	}).Inc()
}

func (d *DiscordCollector) TrackBotUsage(channel, userHandle string) {
	d.totalBotUsage.With(prometheus.Labels{
		channel:    channel,
		userHandle: userHandle,
	}).Inc()
}
