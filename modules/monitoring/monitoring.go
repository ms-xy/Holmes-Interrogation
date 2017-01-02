package monitoring

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/HolmesProcessing/Holmes-Interrogation/context"
	// "github.com/gocql/gocql"
)

// Sometimes the http.Get fails with EADDRNOTAVAIL.
// This is because too many keep-alive connections are created and only time out
// after 2 minutes. Fixing this is possible by not using the default transport
// but rather a custom client that disables keep-alives altogether.
// Additionally, setting the idle connections count up results in less waiting
// time for lots of small parallel connections (default is only 2, one operator
// using Holmes-Frontend and switching to the monitoring tab will use ~12
// parallel connections (roughly, CLOSE-WAIT counted as well as connections
// still in ESTAB state, even if transfer is done, can be easily confirmed using
// `watch "ss -eap | grep Holmes"`)).
var (
	client *http.Client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			MaxIdleConnsPerHost: 0x400,
		},
		Timeout: 5 * time.Second,
	}
)

func GetRoutes() map[string]func(*context.Ctx, *json.RawMessage) *context.Response {
	r := make(map[string]func(*context.Ctx, *json.RawMessage) *context.Response)

	r["get_machines"] = GetMachineUuids
	r["get_planners"] = GetPlanners
	r["get_sysinfo"] = GetSysinfo

	return r
}

func forward(c *context.Ctx, url string) *context.Response {
	var (
		errStr string
		result *json.RawMessage

		resp *http.Response
		err  error
	)

	// Switched to client because it might solve the socket problem we run into
	if resp, err = client.Get(c.HolmesStatus + url); err == nil {
		switch resp.StatusCode {
		case 200:
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				x := json.RawMessage(data)
				result = &x
			}
		case 404:
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				errStr = string(data)
				result = nil
			}
		}
	}

	if err != nil {
		errStr = err.Error()
	}

	return &context.Response{
		Error:  errStr,
		Result: result,
	}
}

func GetMachineUuids(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	return forward(c, "/status/get_machines")
}

type GetMachineUuidsParams struct {
}

func GetPlanners(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	params := &GetPlannersParams{}
	json.Unmarshal(*parametersRaw, params)
	return forward(c, "/status/get_planners/"+params.MachineUuid)
}

type GetPlannersParams struct {
	MachineUuid string
}

// Get system status information
// -----------------------------
//
// Result:
// ```````
// type SystemStatus struct {
//   Uptime int64
//
//   CpuIOWait uint64
//   CpuIdle   uint64
//   CpuBusy   uint64
//   CpuTotal  uint64
//
//   MemoryUsage uint64
//   MemoryMax   uint64
//   SwapUsage   uint64
//   SwapMax     uint64
//
//   Harddrives []*Harddrive
//
//   Loads1  float64 // System load as reported by sysinfo syscall
//   Loads5  float64
//   Loads15 float64
// }
//
// type Harddrive struct {
//   FsType     string
//   MountPoint string
//   Used       uint64
//   Total      uint64
//   Free       uint64
// }
//
func GetSysinfo(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	params := &GetSysinfoParams{}
	json.Unmarshal(*parametersRaw, params)
	return forward(c, "/status/get_sysinfo/"+params.MachineUuid)
}

type GetSysinfoParams struct {
	MachineUuid string
}
