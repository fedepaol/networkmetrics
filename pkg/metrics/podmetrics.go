package podmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricStoreInitSize int = 330
	initialMetricsCount int = 0
	metricsIncVal       int = 1
)

var (
	netAttachDefInstanceEnabledCount      = initialMetricsCount
	netAttachDefInstanceSriovEnabledCount = initialMetricsCount
	//Change this when we set metrics per node.
	objStore = make(map[string]string, metricStoreInitSize) // Preallocate room 110 entires per node*3
	//NetAttachDefInstanceCounter ...  Total no of network attachment definition instance in the cluster

	NetAttachDefPerPod = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "network_attachment_definition_enabled_instance_up",
			Help: "Metric to identify clusters with network attachment definition enabled instances.",
		}, []string{"pod",
			"namespace",
			"interface",
			"nad"})
)

//UpdateNetAttachDefInstanceMetrics ...
func UpdateNetAttachDefInstanceMetrics(podName, namespace, nicName, networkName string, add bool) {
	labels := prometheus.Labels{
		"pod":       podName,
		"namespace": namespace,
		"interface": nicName,
		"nad":       networkName,
	}

	if add {
		NetAttachDefPerPod.With(labels).Add(0)
	}

	NetAttachDefPerPod.Delete(labels)
}
