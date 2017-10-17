// +build linux

package linux

import (
	"regexp"
	"time"

	"github.com/mackerelio/go-osstat/network"
	"github.com/mackerelio/golib/logging"
	"github.com/mackerelio/mackerel-agent/metrics"
)

/*
collect network interface I/O

`interface.{interface}.{metric}.delta`: The increased amount of network I/O per minute retrieved from /proc/net/dev

interface = "eth0", "eth1" and so on...

see interface_test.go for sample input/output
*/

// InterfaceGenerator XXX
type InterfaceGenerator struct {
	Interval time.Duration
}

var interfaceMetrics = []string{
	"rxBytes", "rxPackets", "rxErrors", "rxDrops",
	"rxFifo", "rxFrame", "rxCompressed", "rxMulticast",
	"txBytes", "txPackets", "txErrors", "txDrops",
	"txFifo", "txColls", "txCarrier", "txCompressed",
}

// metrics for posting to Mackerel
var postInterfaceMetricsRegexp = regexp.MustCompile(`^interface\..+\.(?:rxBytes|txBytes)$`)

var interfaceLogger = logging.GetLogger("metrics.interface")

// Generate XXX
func (g *InterfaceGenerator) Generate() (metrics.Values, error) {
	prevValues, err := g.collectInterfacesValues()
	if err != nil {
		return nil, err
	}

	time.Sleep(g.Interval)

	currValues, err := g.collectInterfacesValues()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]float64)
	for name, prevValue := range prevValues {
		if currValue, ok := currValues[name]; ok && currValue >= prevValue {
			ret[name+".delta"] = float64(currValue-prevValue) / g.Interval.Seconds()
		}
	}

	return metrics.Values(ret), nil
}

func (g *InterfaceGenerator) collectInterfacesValues() (map[string]uint64, error) {
	networks, err := network.Get()
	if err != nil {
		interfaceLogger.Errorf("failed to get network statistics: %s", err)
		return nil, err
	}
	if len(networks) == 0 {
		return nil, nil
	}
	results := make(map[string]uint64, len(networks)*2)
	for _, network := range networks {
		results["interface."+network.Name+".rxBytes"] = network.RxBytes
		results["interface."+network.Name+".txBytes"] = network.TxBytes
	}
	return results, nil
}
