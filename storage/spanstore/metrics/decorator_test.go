// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package metrics_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/uber/jaeger/model"
	"github.com/uber/jaeger/storage/spanstore"
	. "github.com/uber/jaeger/storage/spanstore/metrics"
	"github.com/uber/jaeger/storage/spanstore/mocks"
)

func TestSuccessfulUnderlyingCalls(t *testing.T) {
	mf := metrics.NewLocalFactory(0)

	mockReader := mocks.Reader{}
	mrs := NewReadMetricsDecorator(&mockReader, mf)
	mockReader.On("GetServices").Return([]string{}, nil)
	mrs.GetServices()
	mockReader.On("GetOperations", "something").Return([]string{}, nil)
	mrs.GetOperations("something")
	mockReader.On("GetTrace", model.TraceID{}).Return(&model.Trace{}, nil)
	mrs.GetTrace(model.TraceID{})
	mockReader.On("FindTraces", &spanstore.TraceQueryParameters{}).Return([]*model.Trace{}, nil)
	mrs.FindTraces(&spanstore.TraceQueryParameters{})
	counters, gauges := mf.Snapshot()
	expecteds := map[string]int{
		"GetOperations.attempts":  1,
		"GetOperations.successes": 1,
		"GetOperations.errors":    0,
		"GetTrace.attempts":       1,
		"GetTrace.successes":      1,
		"GetTrace.errors":         0,
		"FindTraces.attempts":     1,
		"FindTraces.successes":    1,
		"FindTraces.errors":       0,
		"GetServices.attempts":    1,
		"GetServices.successes":   1,
		"GetServices.errors":      0,
	}

	for k, v := range expecteds {
		assert.EqualValues(t, v, counters[k], k)
	}

	existingKeys := []string{
		"GetOperations.okLatency.P50",
		"GetTrace.responses.P50",
		"FindTraces.okLatency.P50", // this is not exhaustive
	}
	nonExistentKeys := []string{
		"GetOperations.errLatency.P50",
	}
	for _, k := range existingKeys {
		_, ok := gauges[k]
		assert.True(t, ok)
	}

	for _, k := range nonExistentKeys {
		_, ok := gauges[k]
		assert.False(t, ok)
	}
}

func TestFailingUnderlyingCalls(t *testing.T) {
	mf := metrics.NewLocalFactory(0)

	mockReader := mocks.Reader{}
	mrs := NewReadMetricsDecorator(&mockReader, mf)
	mockReader.On("GetServices").Return(nil, errors.New("Failure"))
	mrs.GetServices()
	mockReader.On("GetOperations", "something").Return(nil, errors.New("Failure"))
	mrs.GetOperations("something")
	mockReader.On("GetTrace", model.TraceID{}).Return(nil, errors.New("Failure"))
	mrs.GetTrace(model.TraceID{})
	mockReader.On("FindTraces", &spanstore.TraceQueryParameters{}).Return(nil, errors.New("Failure"))
	mrs.FindTraces(&spanstore.TraceQueryParameters{})
	counters, gauges := mf.Snapshot()
	expecteds := map[string]int{
		"GetOperations.attempts":  1,
		"GetOperations.successes": 0,
		"GetOperations.errors":    1,
		"GetTrace.attempts":       1,
		"GetTrace.successes":      0,
		"GetTrace.errors":         1,
		"FindTraces.attempts":     1,
		"FindTraces.successes":    0,
		"FindTraces.errors":       1,
		"GetServices.attempts":    1,
		"GetServices.successes":   0,
		"GetServices.errors":      1,
	}

	for k, v := range expecteds {
		assert.EqualValues(t, v, counters[k], k)
	}

	existingKeys := []string{
		"GetOperations.errLatency.P50",
	}

	nonExistentKeys := []string{
		"GetOperations.okLatency.P50",
		"GetTrace.responses.P50",
		"Query.okLatency.P50", // this is not exhaustive
	}

	for _, k := range existingKeys {
		_, ok := gauges[k]
		assert.True(t, ok, k)
	}

	for _, k := range nonExistentKeys {
		_, ok := gauges[k]
		assert.False(t, ok)
	}
}