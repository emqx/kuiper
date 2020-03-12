// +build edgex

package extensions

import (
	"context"
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
	"github.com/emqx/kuiper/common"
	"github.com/emqx/kuiper/xstream/api"
	"strconv"
	"strings"
)

type EdgexSource struct {
	client     messaging.MessageClient
	subscribed bool
	vdc        coredata.ValueDescriptorClient
	topic      string
	valueDescs map[string]string
}

func (es *EdgexSource) Configure(device string, props map[string]interface{}) error {
	var protocol = "tcp";
	if p, ok := props["protocol"]; ok {
		protocol = p.(string)
	}
	var server = "localhost";
	if s, ok := props["server"]; ok {
		server = s.(string)
	}
	var port = 5570
	if p, ok := props["port"]; ok {
		port = p.(int)
	}

	if tpc, ok := props["topic"]; ok {
		es.topic = tpc.(string)
	}

	var mbusType = messaging.ZeroMQ
	if t, ok := props["type"]; ok {
		mbusType = t.(string)
	}

	if messaging.ZeroMQ != strings.ToLower(mbusType) {
		mbusType = messaging.MQTT
	}

	if serviceServer, ok := props["serviceServer"]; ok {
		svr := serviceServer.(string) + clients.ApiValueDescriptorRoute
		common.Log.Infof("Connect to value descriptor service at: %s \n", svr)
		es.vdc = coredata.NewValueDescriptorClient(local.New(svr))
		es.valueDescs = make(map[string]string)
	} else {
		return fmt.Errorf("The service server cannot be empty.")
	}

	mbconf := types.MessageBusConfig{SubscribeHost: types.HostInfo{Protocol: protocol, Host: server, Port: port}, Type: messaging.ZeroMQ}
	common.Log.Infof("Use configuration for edgex messagebus %v\n", mbconf)

	var optional = make(map[string]string)
	if ops, ok := props["optional"]; ok {
		if ops1, ok1 := ops.(map[interface{}]interface{}); ok1 {
			for k, v := range ops1 {
				k1 := k.(string)
				v1 := v.(string)
				optional[k1] = v1
			}
		}
		mbconf.Optional = optional
	}

	if client, err := messaging.NewMessageClient(mbconf); err != nil {
		return err
	} else {
		es.client = client
		return nil
	}

}

func (es *EdgexSource) Open(ctx api.StreamContext, consumer chan<- api.SourceTuple, errCh chan<- error) {
	log := ctx.GetLogger()
	if err := es.client.Connect(); err != nil {
		errCh <- fmt.Errorf("Failed to connect to edgex message bus: " + err.Error())
	}
	log.Infof("The connection to edgex messagebus is established successfully.")
	messages := make(chan types.MessageEnvelope)
	topics := []types.TopicChannel{{Topic: es.topic, Messages: messages}}
	err := make(chan error)
	if e := es.client.Subscribe(topics, err); e != nil {
		log.Errorf("Failed to subscribe to edgex messagebus topic %s.\n", e)
		errCh <- e
	} else {
		es.subscribed = true
		log.Infof("Successfully subscribed to edgex messagebus topic %s.", es.topic)
		for {
			select {
			case e1 := <-err:
				errCh <- e1
				return
			case env := <-messages:
				if strings.ToLower(env.ContentType) == "application/json" {
					e := models.Event{}
					if err := e.UnmarshalJSON(env.Payload); err != nil {
						log.Warnf("payload %s unmarshal fail: %v", env.Payload, err)
					} else {
						result := make(map[string]interface{})
						meta := make(map[string]interface{})

						log.Debugf("receive message from device %s", e.Device)
						for _, r := range e.Readings {
							if r.Name != "" {
								if v, err := es.getValue(r, log); err != nil {
									log.Warnf("fail to get value for %s: %v", r.Name, err)
								} else {
									result[strings.ToLower(r.Name)] = v
								}
								r_meta := map[string]interface{}{}
								r_meta["id"] = r.Id
								r_meta["created"] = r.Created
								r_meta["modified"] = r.Modified
								r_meta["origin"] = r.Origin
								r_meta["pushed"] = r.Pushed
								r_meta["device"] = r.Device
								meta[strings.ToLower(r.Name)] = r_meta
							} else {
								log.Warnf("The name of readings should not be empty!")
							}
						}
						if len(result) > 0 {
							meta["id"] = e.ID
							meta["pushed"] = e.Pushed
							meta["device"] = e.Device
							meta["created"] = e.Created
							meta["modified"] = e.Modified
							meta["origin"] = e.Origin
							meta["CorrelationID"] = env.CorrelationID

							select {
							case consumer <- api.NewDefaultSourceTuple(result, meta):
								log.Debugf("send data to device node")
							case <-ctx.Done():
								return
							}
						} else {
							log.Warnf("got an empty result, ignored")
						}
					}
				} else {
					log.Errorf("Unsupported data type %s.", env.ContentType)
				}
			}
		}
	}
}

func (es *EdgexSource) getValue(r models.Reading, logger api.Logger) (interface{}, error) {
	t, err := es.getType(r.Name, logger)
	if err != nil {
		return nil, err
	}
	t = strings.ToUpper(t)
	logger.Debugf("name %s with type %s", r.Name, t)
	v := r.Value
	switch t {
	case "BOOL":
		if r, err := strconv.ParseBool(v); err != nil {
			return nil, err
		} else {
			return r, nil
		}
	case "INT8", "INT16", "INT32", "INT64", "UINT8", "UINT16", "UINT32", "UINT64":
		if r, err := strconv.Atoi(v); err != nil {
			return nil, err
		} else {
			return r, nil
		}
	case "FLOAT32", "FLOAT64":
		if r, err := strconv.ParseFloat(v, 64); err != nil {
			return nil, err
		} else {
			return r, nil
		}
	case "STRING":
		return v, nil
	default:
		logger.Warnf("unknown type %s return the string value", t)
		return v, nil
	}
}

func (es *EdgexSource) fetchAllDataDescriptors() error {
	if vdArr, err := es.vdc.ValueDescriptors(context.Background()); err != nil {
		return err
	} else {
		for _, vd := range vdArr {
			es.valueDescs[vd.Id] = vd.Type
		}
		if len(vdArr) == 0 {
			common.Log.Infof("Cannot find any value descriptors from value descriptor services.")
		} else {
			common.Log.Infof("Get %d of value descriptors from service.", len(vdArr))
		}
	}
	return nil
}

func (es *EdgexSource) getType(id string, logger api.Logger) (string, error) {
	if t, ok := es.valueDescs[id]; ok {
		return t, nil
	} else {
		if e := es.fetchAllDataDescriptors(); e != nil {
			return "", e
		}
		if t, ok := es.valueDescs[id]; ok {
			return t, nil
		} else {
			return "", fmt.Errorf("cannot find type info for %s in value descriptor.", id)
		}
	}
}

func (es *EdgexSource) Close(ctx api.StreamContext) error {
	if es.subscribed {
		if e := es.client.Disconnect(); e != nil {
			return e
		}
	}
	return nil
}