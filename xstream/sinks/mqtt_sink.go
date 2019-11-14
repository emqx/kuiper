package sinks

import (
	"crypto/tls"
	"engine/xstream/api"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strings"
)

type MQTTSink struct {
	srv      string
	tpc      string
	clientid string
	pVersion uint
	uName 	string
	password string
	certPath string
	pkeyPath string

	conn MQTT.Client
}

func NewMqttSink(properties interface{}) (*MQTTSink, error) {
	ps, ok := properties.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expect map[string]interface{} type for the mqtt sink properties")
	}
	srv, ok := ps["server"]
	if !ok {
		return nil, fmt.Errorf("mqtt sink is missing property server")
	}
	tpc, ok := ps["topic"]
	if !ok {
		return nil, fmt.Errorf("mqtt sink is missing property topic")
	}
	clientid, ok := ps["clientId"]
	if !ok{
		if uuid, err := uuid.NewUUID(); err != nil {
			return nil, fmt.Errorf("mqtt sink fails to get uuid, the error is %s", err)
		}else{
			clientid = uuid.String()
		}
	}
	var pVersion uint = 3
	pVersionStr, ok := ps["protocol_version"];
	if ok {
		v, _ := pVersionStr.(string)
		if v == "3.1" {
			pVersion = 3
		} else if v == "3.1.1" {
			pVersion = 4
		} else {
			return nil, fmt.Errorf("Unknown protocol version {0}, the value could be only 3.1 or 3.1.1 (also refers to MQTT version 4).", pVersionStr)
		}
	}

	uName := ""
	un, ok := ps["username"];
	if ok {
		v, _ := un.(string)
		if strings.Trim(v, " ") != "" {
			uName = v
		}
	}

	password := ""
	pwd, ok := ps["password"];
	if ok {
		v, _ := pwd.(string)
		if strings.Trim(v, " ") != "" {
			password = v
		}
	}

	certPath := ""
	if cp, ok := ps["certification_path"]; ok {
		if v, ok := cp.(string); ok {
			certPath = v
		}
	}

	pKeyPath := ""
	if pk, ok := ps["private_key_path"]; ok {
		if v, ok := pk.(string); ok {
			pKeyPath = v
		}
	}

	ms := &MQTTSink{srv: srv.(string), tpc: tpc.(string), clientid: clientid.(string), pVersion:pVersion, uName:uName, password:password, certPath:certPath, pkeyPath:pKeyPath}
	return ms, nil
}

func processPath(p string) (string, error) {
	if abs, err := filepath.Abs(p); err != nil {
		return "", nil
	} else {
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			return "", err;
		}
		return abs, nil
	}
}

func (ms *MQTTSink) Open(ctx api.StreamContext) error {
	log := ctx.GetLogger()
	log.Printf("Opening mqtt sink for rule %s.", ctx.GetRuleId())
	opts := MQTT.NewClientOptions().AddBroker(ms.srv).SetClientID(ms.clientid)

	if ms.certPath != "" || ms.pkeyPath != "" {
		log.Printf("Connect MQTT broker with certification and keys.")
		if cp, err := processPath(ms.certPath); err == nil {
			if kp, err1 := processPath(ms.pkeyPath); err1 == nil {
				if cer, err2 := tls.LoadX509KeyPair(cp, kp); err2 != nil {
					return err2
				} else {
					opts.SetTLSConfig(&tls.Config{Certificates: []tls.Certificate{cer}})
				}
			} else {
				return err1
			}
		} else {
			return err
		}
	} else {
		log.Printf("Connect MQTT broker with username and password.")
		if ms.uName != "" {
			opts = opts.SetUsername(ms.uName)
		}

		if ms.password != "" {
			opts = opts.SetPassword(ms.password)
		}
	}

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Found error: %s", token.Error())
	}
	log.Printf("The connection to server %s was established successfully", ms.srv)
	ms.conn = c
	return nil
}

func (ms *MQTTSink) Collect(ctx api.StreamContext, item interface{}) error {
	logger := ctx.GetLogger()
	c := ms.conn
	logger.Infof("publish %s", item)
	if token := c.Publish(ms.tpc, 0, false, item); token.Wait() && token.Error() != nil {
		return fmt.Errorf("publish error: %s", token.Error())
	}
	return nil
}

func (ms *MQTTSink) Close(ctx api.StreamContext) error {
	logger := ctx.GetLogger()
	logger.Infof("Closing mqtt sink")
	ms.conn.Disconnect(5000)
	return nil
}


