package config

import (
	"github.com/pkg/errors"
	"github.com/struckoff/SFCFramework/curve"
	"strings"
)

type Config struct {
	Address    string `envconfig:"ADDRESS"`
	RPCAddress string `envconfig:"RPC_ADDRESS"`
	Balancer   *BalancerConfig
}

// If config implies use of consul, this options will be taken from consul KV.
// Otherwise it will be taken from config file.
type BalancerConfig struct {
	//Amount of space filling curve dimensions
	Dimensions uint64 `envconfig:"KVROUTER_SFC_DIMENSIONS"`
	//Size of space filling curve
	Size uint64 `envconfig:"KVROUTER_SFC_SIZE"`
	//Space filling curve type
	Curve CurveType `envconfig:"KVROUTER_SFC_CURVE"`
}

type CurveType struct {
	curve.CurveType
}

func (ct *CurveType) UnmarshalJSON(cb []byte) error {
	c := strings.ToLower(string(cb))
	c = strings.Trim(c, "\"")
	switch c {
	case "morton":
		ct.CurveType = curve.Morton
		return nil
	case "hilbert":
		ct.CurveType = curve.Hilbert
		return nil
	default:
		return errors.New("unknown curve type")
	}
}
