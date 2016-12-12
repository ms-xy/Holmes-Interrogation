package monitoring

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/HolmesProcessing/Holmes-Interrogation/context"
	// "github.com/gocql/gocql"
)

func GetRoutes() map[string]func(*context.Ctx, *json.RawMessage) *context.Response {
	r := make(map[string]func(*context.Ctx, *json.RawMessage) *context.Response)

	r["get_uuids"] = GetUuids
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

	if resp, err = http.Get(c.HolmesStatus + url); err == nil {
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

func GetUuids(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	return forward(c, "/status/get_uuids")
}

type SysinfoParams struct {
	Uuid string
}

func GetSysinfo(c *context.Ctx, parametersRaw *json.RawMessage) *context.Response {
	params := &SysinfoParams{}
	json.Unmarshal(*parametersRaw, params)
	return forward(c, "/status/get_sysinfo/"+params.Uuid)
}
