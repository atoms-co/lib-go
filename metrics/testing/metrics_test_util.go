package testing

import (
	"fmt"
	"sort"
	"strings"

	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricproducer"
)

// MetricsToString convert metrics to string for specified metrics names. For testing only.
func MetricsToString(names ...string) string {
	var sb strings.Builder
	for _, name := range names {
		metric := FindMetrics(name)
		if metric != nil {
			sb.WriteString(MetricToString(metric))
		}
	}
	return sb.String()
}

// FindMetrics find metrics by name. For testing only.
func FindMetrics(name string) *metricdata.Metric {
	metricproducer.GlobalManager().GetAll()
	metrics := metricproducer.GlobalManager().GetAll()[0].Read()
	for _, metric := range metrics {
		if metric.Descriptor.Name == name {
			return metric
		}
	}
	return nil
}

// MetricToString convert metrics to a string. For test only
func MetricToString(metric *metricdata.Metric) string {
	var tsString []string
	for _, ts := range metric.TimeSeries {
		tsString = append(tsString, timeSeriesToString(ts))
	}
	sort.Strings(tsString)

	var sb strings.Builder
	name := metric.Descriptor.Name
	sb.WriteString(name + "(")
	for _, label := range metric.Descriptor.LabelKeys {
		sb.WriteString(label.Key + ",")
	}
	sb.WriteString(") {\n")
	sb.WriteString(strings.Join(tsString, "\n"))
	sb.WriteString("\n}\n")
	return sb.String()
}

func timeSeriesToString(ts *metricdata.TimeSeries) string {
	var sb strings.Builder
	sb.WriteString("  (")
	for _, v := range ts.LabelValues {
		sb.WriteString(v.Value + ",")
	}
	sb.WriteString(") : ")
	sb.WriteString(fmt.Sprintf("%v", ts.Points[0].Value))
	return sb.String()
}

// FindMetricDataPoint utility to find collected metrics by name and tag values for testing only
func FindMetricDataPoint(name string, tagVal ...string) *metricdata.Point {
	metric := FindMetrics(name)
	for _, ts := range metric.TimeSeries {
		if !tagValMatches(ts, tagVal...) {
			continue
		}
		return &ts.Points[0]
	}
	return nil
}

func tagValMatches(ts *metricdata.TimeSeries, tagVal ...string) bool {
	if len(ts.LabelValues) != len(tagVal) {
		return false
	}
	for i, v := range ts.LabelValues {
		if v.Value != tagVal[i] {
			return false
		}
	}
	return true
}
