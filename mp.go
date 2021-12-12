package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricProcessor struct {
	logger log.Logger
}

func (mp MetricProcessor) GetMetricType(metricType string, metricsType map[string]string) prometheus.ValueType {
	var strToPromType = map[string]prometheus.ValueType{
		"gauge":     prometheus.GaugeValue,
		"counter":   prometheus.CounterValue,
		"histogram": prometheus.UntypedValue,
	}

	strType, ok := metricsType[strings.ToLower(metricType)]
	if !ok {
		return prometheus.GaugeValue
	}
	valueType, ok := strToPromType[strings.ToLower(strType)]
	if !ok {
		panic(errors.New("Error while getting prometheus type " + strings.ToLower(strType)))
	}
	return valueType
}

// interface method to call ScrapeGenericValues using Metric struct values
func (mp MetricProcessor) ScrapeMetric(ch chan<- prometheus.Metric, md MultiDatabase, metricDefinition Metric) error {
	level.Debug(mp.logger).Log("Calling function ScrapeGenericValues()")
	return mp.ScrapeGenericValues(ch, md, metricDefinition.Context, metricDefinition.Labels,
		metricDefinition.MetricsDesc, metricDefinition.MetricsType, metricDefinition.MetricsBuckets,
		metricDefinition.FieldToAppend, metricDefinition.IgnoreZeroResult,
		metricDefinition.Request)
}

// generic method for retrieving metrics.
func (mp MetricProcessor) ScrapeGenericValues(ch chan<- prometheus.Metric, md MultiDatabase, context string, labels []string,
	metricsDesc map[string]string, metricsType map[string]string, metricsBuckets map[string]map[string]string, fieldToAppend string, ignoreZeroResult bool, sqlText string) error {
	metricsCount := 0
	genericParser := func(row map[string]string) error {
		// Construct labels value
		labelsValues := []string{}
		for _, label := range labels {
			labelsValues = append(labelsValues, row[label])
		}
		// Construct Prometheus values to sent back
		for metric, metricHelp := range metricsDesc {

			value, err := strconv.ParseFloat(strings.TrimSpace(row[metric]), 64)
			// If not a float, skip current metric
			if err != nil {
				level.Error(mp.logger).Log("Unable to convert current value to float (metric=" + metric + ",metricHelp=" + metricHelp + ",value=<" + row[metric] + ">)")
				continue
			}
			level.Debug(mp.logger).Log("Query result looks like: ", value)
			// If metric do not use a field content in metric's name
			if strings.Compare(fieldToAppend, "") == 0 {
				desc := prometheus.NewDesc(
					prometheus.BuildFQName(namespace, context, metric),
					metricHelp,
					labels, nil,
				)
				if metricsType[strings.ToLower(metric)] == "histogram" {
					count, err := strconv.ParseUint(strings.TrimSpace(row["count"]), 10, 64)
					if err != nil {
						level.Error(mp.logger).Log("Unable to convert count value to int (metric=" + metric + ",metricHelp=" + metricHelp + ",value=<" + row["count"] + ">)")
						continue
					}
					buckets := make(map[float64]uint64)
					for field, le := range metricsBuckets[metric] {
						lelimit, err := strconv.ParseFloat(strings.TrimSpace(le), 64)
						if err != nil {
							level.Error(mp.logger).Log("Unable to convert bucket limit value to float (metric=" + metric + ",metricHelp=" + metricHelp + ",bucketlimit=<" + le + ">)")
							continue
						}
						counter, err := strconv.ParseUint(strings.TrimSpace(row[field]), 10, 64)
						if err != nil {
							level.Error(mp.logger).Log("Unable to convert ", field, " value to int (metric="+metric+",metricHelp="+metricHelp+",value=<"+row[field]+">)")
							continue
						}
						buckets[lelimit] = counter
					}
					ch <- prometheus.MustNewConstHistogram(desc, count, value, buckets, labelsValues...)
				} else {
					ch <- prometheus.MustNewConstMetric(desc, mp.GetMetricType(metric, metricsType), value, labelsValues...)
				}
				// If no labels, use metric name
			} else {
				desc := prometheus.NewDesc(
					prometheus.BuildFQName(namespace, context, mp.cleanName(row[fieldToAppend])),
					metricHelp,
					nil, nil,
				)
				if metricsType[strings.ToLower(metric)] == "histogram" {
					count, err := strconv.ParseUint(strings.TrimSpace(row["count"]), 10, 64)
					if err != nil {
						level.Error(mp.logger).Log("Unable to convert count value to int (metric=" + metric + ",metricHelp=" + metricHelp + ",value=<" + row["count"] + ">)")
						continue
					}
					buckets := make(map[float64]uint64)
					for field, le := range metricsBuckets[metric] {
						lelimit, err := strconv.ParseFloat(strings.TrimSpace(le), 64)
						if err != nil {
							level.Error(mp.logger).Log("Unable to convert bucket limit value to float (metric=" + metric + ",metricHelp=" + metricHelp + ",bucketlimit=<" + le + ">)")
							continue
						}
						counter, err := strconv.ParseUint(strings.TrimSpace(row[field]), 10, 64)
						if err != nil {
							level.Error(mp.logger).Log("Unable to convert ", field, " value to int (metric="+metric+",metricHelp="+metricHelp+",value=<"+row[field]+">)")
							continue
						}
						buckets[lelimit] = counter
					}
					ch <- prometheus.MustNewConstHistogram(desc, count, value, buckets)
				} else {
					ch <- prometheus.MustNewConstMetric(desc, mp.GetMetricType(metric, metricsType), value)
				}
			}
			metricsCount++
		}
		return nil
	}
	err := mp.GeneratePrometheusMetrics(genericParser, md, sqlText)
	level.Debug(mp.logger).Log("ScrapeGenericValues() - metricsCount: ", metricsCount)
	if err != nil {
		return err
	}
	// if !ignoreZeroResult && metricsCount == 0 {
	// 	return errors.New("No metrics found while parsing")
	// }
	return err
}

// inspired by https://kylewbanks.com/blog/query-result-to-map-in-golang
// Parse SQL result and call parsing function to each row
func (mp MetricProcessor) GeneratePrometheusMetrics(parse func(row map[string]string) error, md MultiDatabase, sqlText string) error {
	rows, err := md.Query(sqlText)

	// if ctx.Err() == context.DeadlineExceeded {
	// 	return errors.New("Oracle query timed out")
	// }

	if err != nil {
		return err
	}

	for _, row := range rows {

		m := make(map[string]string)
		for col, val := range row {
			m[strings.ToLower(col)] = fmt.Sprintf("%v", val)
		}

		if err := parse(m); err != nil {
			return err
		}
	}
	return nil
}

// Oracle gives us some ugly names back. This function cleans things up for Prometheus.
func (mp MetricProcessor) cleanName(s string) string {
	s = strings.Replace(s, " ", "_", -1) // Remove spaces
	s = strings.Replace(s, "(", "", -1)  // Remove open parenthesis
	s = strings.Replace(s, ")", "", -1)  // Remove close parenthesis
	s = strings.Replace(s, "/", "", -1)  // Remove forward slashes
	s = strings.Replace(s, "*", "", -1)  // Remove asterisks
	s = strings.ToLower(s)
	return s
}
