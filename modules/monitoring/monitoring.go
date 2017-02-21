package monitoring

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/HolmesProcessing/Holmes-Interrogation/context"

	"fmt"
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
	r["get_netinfo"] = GetNetinfo
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
		default:
			defer resp.Body.Close()
			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				errStr = "Storage Response: [HTTP " + string(resp.StatusCode) + "] " + string(data)
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
	limit := strconv.FormatInt(int64(params.Limit), 10)
	url := "/status/get_sysinfo/" + params.MachineUuid + "/" + limit
	return forward(c, url)
}

type GetSysinfoParams struct {
	MachineUuid string
	Limit       int
}

// Get network status information
// ------------------------------
//
// Result:
// ```````
//
// type NetworkStatus struct {
//   Interfaces []*NetworkInterface
// }
//
// type NetworkInterface struct {
//   ID        int
//   Name      string
//   IP        net.IP
//   Netmask   net.IPMask
//   Broadcast net.IP
//   Scope     string
// }
//
func GetNetinfo(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	params := &GetNetinfoParams{}
	json.Unmarshal(*parametersRaw, params)
	return forward(c, "/status/get_netinfo/"+params.MachineUuid)
}

type GetNetinfoParams struct {
	MachineUuid string
}

// Get planners information
// ------------------------
//
// Result:
// ```````
//
// map[uint64]*PlannerInformation
//
// type PlannerInformation struct {
//   Name          string
//   PID           uint64
//   IP            net.IP
//   Port          int
//   Configuration string
//   Logs          *LogBuffer
//   Services      map[uint16]*ServiceInformation
// }
//
// type ServiceInformation struct {
//   Configuration string
//   Name          string
//   Port          uint16
//   Task          string
//   Logs          *LogBuffer
// }
//
func GetPlanners(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	params := &GetPlannersParams{}
	json.Unmarshal(*parametersRaw, params)
	return forward(c, "/status/get_planners/"+params.MachineUuid)
}

type GetPlannersParams struct {
	MachineUuid string
}
