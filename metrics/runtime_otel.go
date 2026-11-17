package metrics

import (
	"context"
	"fmt"
	"math"
	"runtime"
	rtmetrics "runtime/metrics"
	"sync"
	"time"

	"go.opentelemetry.io/otel/metric"
)

// runtimeSnapshot is the single runtime sample used for one OTel collection.
//
// The values mirror the runtime.MemStats fields the legacy metrics exported,
// but are read through runtime/metrics. runtime.ReadMemStats stops the world
// on every call, which stalled all goroutines on every /metrics scrape;
// runtime/metrics.Read does not.
type runtimeSnapshot struct {
	totalAlloc    uint64
	sys           uint64
	mallocs       uint64
	frees         uint64
	heapAlloc     uint64
	heapSys       uint64
	heapIdle      uint64
	heapInuse     uint64
	heapReleased  uint64
	heapObjects   uint64
	stackInuse    uint64
	stackSys      uint64
	mSpanInuse    uint64
	mSpanSys      uint64
	mCacheInuse   uint64
	mCacheSys     uint64
	gcSys         uint64
	otherSys      uint64
	numGC         uint64
	numForcedGC   uint64
	nextGC        uint64
	pauseTotal    time.Duration
	gcCPUFraction float64
	numGoroutine  int64
	numCgoCall    int64
}

type runtimeSnapshotReader func() runtimeSnapshot

// runtime/metrics names read for each snapshot. See
// https://pkg.go.dev/runtime/metrics#hdr-Supported_metrics.
const (
	rmHeapAllocsBytes       = "/gc/heap/allocs:bytes"
	rmHeapAllocsObjects     = "/gc/heap/allocs:objects"
	rmHeapFreesObjects      = "/gc/heap/frees:objects"
	rmHeapTinyAllocsObjects = "/gc/heap/tiny/allocs:objects"
	rmHeapObjects           = "/gc/heap/objects:objects"
	rmHeapGoal              = "/gc/heap/goal:bytes"
	rmGCCycles              = "/gc/cycles/total:gc-cycles"
	rmGCCyclesForced        = "/gc/cycles/forced:gc-cycles"
	rmMemTotal              = "/memory/classes/total:bytes"
	rmMemHeapObjects        = "/memory/classes/heap/objects:bytes"
	rmMemHeapUnused         = "/memory/classes/heap/unused:bytes"
	rmMemHeapFree           = "/memory/classes/heap/free:bytes"
	rmMemHeapReleased       = "/memory/classes/heap/released:bytes"
	rmMemHeapStacks         = "/memory/classes/heap/stacks:bytes"
	rmMemOSStacks           = "/memory/classes/os-stacks:bytes"
	rmMemMSpanInuse         = "/memory/classes/metadata/mspan/inuse:bytes"
	rmMemMSpanFree          = "/memory/classes/metadata/mspan/free:bytes"
	rmMemMCacheInuse        = "/memory/classes/metadata/mcache/inuse:bytes"
	rmMemMCacheFree         = "/memory/classes/metadata/mcache/free:bytes"
	rmMemMetadataOther      = "/memory/classes/metadata/other:bytes"
	rmMemOther              = "/memory/classes/other:bytes"
	rmGCPauses              = "/sched/pauses/total/gc:seconds"
	rmCPUGCTotal            = "/cpu/classes/gc/total:cpu-seconds"
	rmCPUGCMarkIdle         = "/cpu/classes/gc/mark/idle:cpu-seconds"
	rmCPUTotal              = "/cpu/classes/total:cpu-seconds"
)

var runtimeMetricNames = []string{
	rmHeapAllocsBytes,
	rmHeapAllocsObjects,
	rmHeapFreesObjects,
	rmHeapTinyAllocsObjects,
	rmHeapObjects,
	rmHeapGoal,
	rmGCCycles,
	rmGCCyclesForced,
	rmMemTotal,
	rmMemHeapObjects,
	rmMemHeapUnused,
	rmMemHeapFree,
	rmMemHeapReleased,
	rmMemHeapStacks,
	rmMemOSStacks,
	rmMemMSpanInuse,
	rmMemMSpanFree,
	rmMemMCacheInuse,
	rmMemMCacheFree,
	rmMemMetadataOther,
	rmMemOther,
	rmGCPauses,
	rmCPUGCTotal,
	rmCPUGCMarkIdle,
	rmCPUTotal,
}

func readRuntimeSnapshot() runtimeSnapshot {
	// Allocated per call: OTel may invoke the callback from several readers
	// concurrently, and the slice is small.
	samples := make([]rtmetrics.Sample, len(runtimeMetricNames))
	for i, name := range runtimeMetricNames {
		samples[i].Name = name
	}
	rtmetrics.Read(samples)
	return runtimeSnapshotFromSamples(samples, int64(runtime.NumGoroutine()), runtime.NumCgoCall())
}

// runtimeSnapshotFromSamples derives the MemStats-equivalent values the same
// way runtime.ReadMemStats does internally (see readmemstats_m in
// runtime/mstats.go).
func runtimeSnapshotFromSamples(samples []rtmetrics.Sample, numGoroutine, numCgoCall int64) runtimeSnapshot {
	u := make(map[string]uint64, len(samples))
	f := make(map[string]float64, len(samples))
	var pauses *rtmetrics.Float64Histogram
	for _, sample := range samples {
		switch sample.Value.Kind() {
		case rtmetrics.KindUint64:
			u[sample.Name] = sample.Value.Uint64()
		case rtmetrics.KindFloat64:
			f[sample.Name] = sample.Value.Float64()
		case rtmetrics.KindFloat64Histogram:
			if sample.Name == rmGCPauses {
				pauses = sample.Value.Float64Histogram()
			}
		}
		// KindBad means the metric is unsupported by this Go version; leave it 0.
	}

	var gcCPUFraction float64
	if total := f[rmCPUTotal]; total > 0 {
		// MemStats.GCCPUFraction omits idle mark time, which is "free".
		gcCPUFraction = (f[rmCPUGCTotal] - f[rmCPUGCMarkIdle]) / total
	}

	return runtimeSnapshot{
		totalAlloc: u[rmHeapAllocsBytes],
		sys:        u[rmMemTotal],
		// MemStats counts tiny allocations as both allocated and freed.
		mallocs:       u[rmHeapAllocsObjects] + u[rmHeapTinyAllocsObjects],
		frees:         u[rmHeapFreesObjects] + u[rmHeapTinyAllocsObjects],
		heapAlloc:     u[rmMemHeapObjects],
		heapSys:       u[rmMemHeapObjects] + u[rmMemHeapUnused] + u[rmMemHeapFree] + u[rmMemHeapReleased],
		heapIdle:      u[rmMemHeapFree] + u[rmMemHeapReleased],
		heapInuse:     u[rmMemHeapObjects] + u[rmMemHeapUnused],
		heapReleased:  u[rmMemHeapReleased],
		heapObjects:   u[rmHeapObjects],
		stackInuse:    u[rmMemHeapStacks],
		stackSys:      u[rmMemHeapStacks] + u[rmMemOSStacks],
		mSpanInuse:    u[rmMemMSpanInuse],
		mSpanSys:      u[rmMemMSpanInuse] + u[rmMemMSpanFree],
		mCacheInuse:   u[rmMemMCacheInuse],
		mCacheSys:     u[rmMemMCacheInuse] + u[rmMemMCacheFree],
		gcSys:         u[rmMemMetadataOther],
		otherSys:      u[rmMemOther],
		numGC:         u[rmGCCycles],
		numForcedGC:   u[rmGCCyclesForced],
		nextGC:        u[rmHeapGoal],
		pauseTotal:    histogramTotalDuration(pauses),
		gcCPUFraction: gcCPUFraction,
		numGoroutine:  numGoroutine,
		numCgoCall:    numCgoCall,
	}
}

// histogramTotalDuration approximates the sum of all observations in a
// runtime/metrics time histogram. runtime/metrics exposes no exact pause sum,
// so each bucket contributes count * midpoint. Go's time histograms have four
// sub-buckets per power of two, which bounds the error at roughly 12%.
func histogramTotalDuration(h *rtmetrics.Float64Histogram) time.Duration {
	if h == nil {
		return 0
	}
	var seconds float64
	for i, count := range h.Counts {
		if count == 0 {
			continue
		}
		lo, hi := h.Buckets[i], h.Buckets[i+1]
		var value float64
		switch {
		case math.IsInf(lo, -1):
			value = math.Max(hi, 0)
		case math.IsInf(hi, 1):
			value = lo
		default:
			value = (lo + hi) / 2
		}
		seconds += float64(count) * value
	}
	return time.Duration(seconds * float64(time.Second))
}

type runtimeInt64MetricDefinition struct {
	name        Name
	description string
	unit        UnitType
	cumulative  bool
	value       func(runtimeSnapshot) int64
}

type runtimeFloat64MetricDefinition struct {
	name        Name
	description string
	unit        UnitType
	value       func(runtimeSnapshot) float64
}

// These definitions preserve the legacy runtime metric names, descriptions,
// units, and cumulative classification.
var runtimeInt64MetricDefinitions = []runtimeInt64MetricDefinition{
	{
		name:        "process/memory_alloc",
		description: "Number of bytes currently allocated in use",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.heapAlloc) },
	},
	{
		name:        "process/total_memory_alloc",
		description: "Number of allocations in total",
		unit:        UnitBytes,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.totalAlloc) },
	},
	{
		name:        "process/sys_memory_alloc",
		description: "Number of bytes given to the process to use in total",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.sys) },
	},
	{
		name:        "process/memory_malloc",
		description: "Cumulative count of heap objects allocated",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.mallocs) },
	},
	{
		name:        "process/memory_frees",
		description: "Cumulative count of heap objects freed",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.frees) },
	},
	{
		name:        "process/heap_alloc",
		description: "Process heap allocation",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.heapAlloc) },
	},
	{
		name:        "process/sys_heap",
		description: "Bytes of heap memory obtained from the OS",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.heapSys) },
	},
	{
		name:        "process/heap_idle",
		description: "Bytes in idle (unused) spans",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.heapIdle) },
	},
	{
		name:        "process/heap_inuse",
		description: "Bytes in in-use spans",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.heapInuse) },
	},
	{
		name:        "process/heap_release",
		description: "The cumulative number of objects released from the heap",
		unit:        UnitBytes,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.heapReleased) },
	},
	{
		name:        "process/heap_objects",
		description: "The number of objects allocated on the heap",
		unit:        UnitDimensionless,
		value:       func(s runtimeSnapshot) int64 { return int64(s.heapObjects) },
	},
	{
		name:        "process/stack_inuse",
		description: "Bytes in stack spans",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.stackInuse) },
	},
	{
		name:        "process/sys_stack",
		description: "The memory used by stack spans and OS thread stacks",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.stackSys) },
	},
	{
		name:        "process/stack_mspan_inuse",
		description: "Bytes of allocated mspan structures",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.mSpanInuse) },
	},
	{
		name:        "process/sys_stack_mspan",
		description: "Bytes of memory obtained from the OS for mspan structures",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.mSpanSys) },
	},
	{
		name:        "process/stack_mcache_inuse",
		description: "Bytes of allocated mcache structures",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.mCacheInuse) },
	},
	{
		name:        "process/sys_stack_mcache",
		description: "Bytes of memory obtained from the OS for mcache structures",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.mCacheSys) },
	},
	{
		name:        "process/gc_sys",
		description: "Bytes of memory in garbage collection metadatas",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.gcSys) },
	},
	{
		name:        "process/other_sys",
		description: "Bytes of memory in miscellaneous off-heap runtime allocations",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.otherSys) },
	},
	{
		name:        "process/num_gc",
		description: "Cumulative count of completed GC cycles",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.numGC) },
	},
	{
		name:        "process/num_forced_gc",
		description: "Cumulative count of GC cycles forced by the application",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.numForcedGC) },
	},
	{
		name:        "process/next_gc_heap_size",
		description: "Target heap size of the next GC cycle in bytes",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.nextGC) },
	},
	{
		name:        "process/pause_total",
		description: "Cumulative milliseconds spent in GC stop-the-world pauses",
		unit:        UnitMilliseconds,
		cumulative:  true,
		value: func(s runtimeSnapshot) int64 {
			return s.pauseTotal.Milliseconds()
		},
	},
	{
		name:        "process/cpu_goroutines",
		description: "Number of goroutines that currently exist",
		unit:        UnitDimensionless,
		value:       func(s runtimeSnapshot) int64 { return s.numGoroutine },
	},
	{
		name:        "process/cpu_cgo_calls",
		description: "Number of cgo calls made by the current process",
		unit:        UnitDimensionless,
		value:       func(s runtimeSnapshot) int64 { return s.numCgoCall },
	},
}

var runtimeFloat64MetricDefinitions = []runtimeFloat64MetricDefinition{
	{
		name:        "process/gc_cpu_fraction",
		description: "Fraction of this program's available CPU time used by the GC since the program started",
		unit:        UnitDimensionless,
		value:       func(s runtimeSnapshot) float64 { return s.gcCPUFraction },
	},
}

type runtimeInt64Instrument struct {
	definition runtimeInt64MetricDefinition
	instrument metric.Int64Observable
}

type runtimeFloat64Instrument struct {
	definition runtimeFloat64MetricDefinition
	instrument metric.Float64Observable
}

var (
	otelRuntimeRegistrationMu sync.Mutex
	otelRuntimeRegistration   metric.Registration
)

func enableOTelRuntimeMetrics() error {
	otelRuntimeRegistrationMu.Lock()
	defer otelRuntimeRegistrationMu.Unlock()

	// Replace the previous registration when Init is called more than once.
	// Otherwise duplicate callbacks could observe the same series.
	if otelRuntimeRegistration != nil {
		if err := otelRuntimeRegistration.Unregister(); err != nil {
			return fmt.Errorf("unregister previous OTel runtime metrics callback: %w", err)
		}
		otelRuntimeRegistration = nil
	}

	registration, err := newOTelRuntimeMetrics(otelMeter, readRuntimeSnapshot)
	if err != nil {
		return err
	}
	otelRuntimeRegistration = registration
	return nil
}

func newOTelRuntimeMetrics(
	meter metric.Meter,
	readSnapshot runtimeSnapshotReader,
) (metric.Registration, error) {
	if readSnapshot == nil {
		return nil, fmt.Errorf("create OTel runtime metrics: snapshot reader must not be nil")
	}

	int64Instruments := make([]runtimeInt64Instrument, 0, len(runtimeInt64MetricDefinitions))
	observables := make([]metric.Observable, 0, len(runtimeInt64MetricDefinitions)+len(runtimeFloat64MetricDefinitions))
	for _, definition := range runtimeInt64MetricDefinitions {
		options := []metric.Int64ObservableGaugeOption{
			metric.WithDescription(definition.description),
			metric.WithUnit(string(definition.unit)),
		}
		if definition.cumulative {
			instrument, err := meter.Int64ObservableCounter(definition.name, metric.WithDescription(definition.description), metric.WithUnit(string(definition.unit)))
			if err != nil {
				return nil, fmt.Errorf("create OTel runtime counter %q: %w", definition.name, err)
			}
			int64Instruments = append(int64Instruments, runtimeInt64Instrument{
				definition: definition,
				instrument: instrument,
			})
			observables = append(observables, instrument)
			continue
		}

		instrument, err := meter.Int64ObservableGauge(definition.name, options...)
		if err != nil {
			return nil, fmt.Errorf("create OTel runtime gauge %q: %w", definition.name, err)
		}
		int64Instruments = append(int64Instruments, runtimeInt64Instrument{
			definition: definition,
			instrument: instrument,
		})
		observables = append(observables, instrument)
	}

	float64Instruments := make([]runtimeFloat64Instrument, 0, len(runtimeFloat64MetricDefinitions))
	for _, definition := range runtimeFloat64MetricDefinitions {
		instrument, err := meter.Float64ObservableGauge(definition.name, metric.WithDescription(definition.description), metric.WithUnit(string(definition.unit)))
		if err != nil {
			return nil, fmt.Errorf("create OTel runtime gauge %q: %w", definition.name, err)
		}
		float64Instruments = append(float64Instruments, runtimeFloat64Instrument{
			definition: definition,
			instrument: instrument,
		})
		observables = append(observables, instrument)
	}

	attributes := newOTelAttributes(nil)
	registration, err := meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			snapshot := readSnapshot()
			option := attributes.option(nil)
			for _, instrument := range int64Instruments {
				observer.ObserveInt64(instrument.instrument, instrument.definition.value(snapshot), option)
			}
			for _, instrument := range float64Instruments {
				observer.ObserveFloat64(instrument.instrument, instrument.definition.value(snapshot), option)
			}
			return nil
		},
		observables...,
	)
	if err != nil {
		return nil, fmt.Errorf("register OTel runtime metrics callback: %w", err)
	}
	return registration, nil
}
