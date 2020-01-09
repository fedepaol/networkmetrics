package controller

import (
	"github.com/fedepaol/network-metrics/pkg/controller/metrics"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, metrics.Add)
}
