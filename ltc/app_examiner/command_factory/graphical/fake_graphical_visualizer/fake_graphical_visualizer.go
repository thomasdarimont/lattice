// This file was generated by counterfeiter
package fake_graphical_visualizer

import (
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/graphical"
)

type FakeGraphicalVisualizer struct {
	PrintDistributionChartStub        func(rate time.Duration) error
	printDistributionChartMutex       sync.RWMutex
	printDistributionChartArgsForCall []struct {
		rate time.Duration
	}
	printDistributionChartReturns struct {
		result1 error
	}
}

func (fake *FakeGraphicalVisualizer) PrintDistributionChart(rate time.Duration) error {
	fake.printDistributionChartMutex.Lock()
	fake.printDistributionChartArgsForCall = append(fake.printDistributionChartArgsForCall, struct {
		rate time.Duration
	}{rate})
	fake.printDistributionChartMutex.Unlock()
	if fake.PrintDistributionChartStub != nil {
		return fake.PrintDistributionChartStub(rate)
	} else {
		return fake.printDistributionChartReturns.result1
	}
}

func (fake *FakeGraphicalVisualizer) PrintDistributionChartCallCount() int {
	fake.printDistributionChartMutex.RLock()
	defer fake.printDistributionChartMutex.RUnlock()
	return len(fake.printDistributionChartArgsForCall)
}

func (fake *FakeGraphicalVisualizer) PrintDistributionChartArgsForCall(i int) time.Duration {
	fake.printDistributionChartMutex.RLock()
	defer fake.printDistributionChartMutex.RUnlock()
	return fake.printDistributionChartArgsForCall[i].rate
}

func (fake *FakeGraphicalVisualizer) PrintDistributionChartReturns(result1 error) {
	fake.PrintDistributionChartStub = nil
	fake.printDistributionChartReturns = struct {
		result1 error
	}{result1}
}

var _ graphical.GraphicalVisualizer = new(FakeGraphicalVisualizer)
