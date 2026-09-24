package metrics

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/otel/metric"
)

// runtimeSnapshot is the single runtime sample used for one OTel collection.
type runtimeSnapshot struct {
	memStats     runtime.MemStats
	numGoroutine int64
	numCgoCall   int64
}

type runtimeSnapshotReader func() runtimeSnapshot

func readRuntimeSnapshot() runtimeSnapshot {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return runtimeSnapshot{
		memStats:     memStats,
		numGoroutine: int64(runtime.NumGoroutine()),
		numCgoCall:   runtime.NumCgoCall(),
	}
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
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.Alloc) },
	},
	{
		name:        "process/total_memory_alloc",
		description: "Number of allocations in total",
		unit:        UnitBytes,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.TotalAlloc) },
	},
	{
		name:        "process/sys_memory_alloc",
		description: "Number of bytes given to the process to use in total",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.Sys) },
	},
	{
		name:        "process/memory_lookups",
		description: "Cumulative number of pointer lookups performed by the runtime",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.Lookups) },
	},
	{
		name:        "process/memory_malloc",
		description: "Cumulative count of heap objects allocated",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.Mallocs) },
	},
	{
		name:        "process/memory_frees",
		description: "Cumulative count of heap objects freed",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.Frees) },
	},
	{
		name:        "process/heap_alloc",
		description: "Process heap allocation",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.HeapAlloc) },
	},
	{
		name:        "process/sys_heap",
		description: "Bytes of heap memory obtained from the OS",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.HeapSys) },
	},
	{
		name:        "process/heap_idle",
		description: "Bytes in idle (unused) spans",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.HeapIdle) },
	},
	{
		name:        "process/heap_inuse",
		description: "Bytes in in-use spans",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.HeapInuse) },
	},
	{
		name:        "process/heap_release",
		description: "The cumulative number of objects released from the heap",
		unit:        UnitBytes,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.HeapReleased) },
	},
	{
		name:        "process/heap_objects",
		description: "The number of objects allocated on the heap",
		unit:        UnitDimensionless,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.HeapObjects) },
	},
	{
		name:        "process/stack_inuse",
		description: "Bytes in stack spans",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.StackInuse) },
	},
	{
		name:        "process/sys_stack",
		description: "The memory used by stack spans and OS thread stacks",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.StackSys) },
	},
	{
		name:        "process/stack_mspan_inuse",
		description: "Bytes of allocated mspan structures",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.MSpanInuse) },
	},
	{
		name:        "process/sys_stack_mspan",
		description: "Bytes of memory obtained from the OS for mspan structures",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.MSpanSys) },
	},
	{
		name:        "process/stack_mcache_inuse",
		description: "Bytes of allocated mcache structures",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.MCacheInuse) },
	},
	{
		name:        "process/sys_stack_mcache",
		description: "Bytes of memory obtained from the OS for mcache structures",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.MCacheSys) },
	},
	{
		name:        "process/gc_sys",
		description: "Bytes of memory in garbage collection metadatas",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.GCSys) },
	},
	{
		name:        "process/other_sys",
		description: "Bytes of memory in miscellaneous off-heap runtime allocations",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.OtherSys) },
	},
	{
		name:        "process/num_gc",
		description: "Cumulative count of completed GC cycles",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.NumGC) },
	},
	{
		name:        "process/num_forced_gc",
		description: "Cumulative count of GC cycles forced by the application",
		unit:        UnitDimensionless,
		cumulative:  true,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.NumForcedGC) },
	},
	{
		name:        "process/next_gc_heap_size",
		description: "Target heap size of the next GC cycle in bytes",
		unit:        UnitBytes,
		value:       func(s runtimeSnapshot) int64 { return int64(s.memStats.NextGC) },
	},
	{
		name:        "process/last_gc_finished_timestamp",
		description: "Time the last garbage collection finished, as milliseconds since 1970 (the UNIX epoch)",
		unit:        UnitMilliseconds,
		value: func(s runtimeSnapshot) int64 {
			return int64(s.memStats.LastGC) / int64(time.Millisecond)
		},
	},
	{
		name:        "process/pause_total",
		description: "Cumulative milliseconds spent in GC stop-the-world pauses",
		unit:        UnitMilliseconds,
		cumulative:  true,
		value: func(s runtimeSnapshot) int64 {
			return int64(s.memStats.PauseTotalNs) / int64(time.Millisecond)
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
		value:       func(s runtimeSnapshot) float64 { return s.memStats.GCCPUFraction },
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
