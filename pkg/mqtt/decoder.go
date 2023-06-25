package mqtt

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/eclipse/paho.golang/paho"
	"k8s.io/apimachinery/pkg/types"

	mqttv1 "github.com/morvencao/event-controller/pkg/apis/v1"
)

// Decodes MQTT messages into unstructured.Unstructured.
type Decoder struct {
	sink DecoderSink
}

func NewDecoder(sink DecoderSink) *Decoder {
	return &Decoder{
		sink: sink,
	}
}

func (d *Decoder) Decode(m *paho.Publish) error {
	if len(m.Payload) == 0 || m.Payload == nil {
		return d.sink.Delete(uidFromTopic(m.Topic))
	}

	obj, err := d.decode(m)
	if err != nil {
		return err
	}
	return d.sink.Update(obj)
}

func uidFromTopic(topic string) types.UID {
	raw := path.Base(strings.TrimSuffix(topic, "/content"))
	return types.UID(raw)
}

func (d *Decoder) decode(m *paho.Publish) (
	rm *mqttv1.ResourceMessage, err error) {
	// TODO: implement contentType

	rm = &mqttv1.ResourceMessage{}
	err = json.Unmarshal(m.Payload, rm)
	return
}

type DecoderSink interface {
	Update(msg *mqttv1.ResourceMessage) error
	Delete(uid types.UID) error
}

type DecoderSinkList []DecoderSink

func (l DecoderSinkList) Update(msg *mqttv1.ResourceMessage) error {
	for _, d := range l {
		if err := d.Update(msg); err != nil {
			return err
		}
	}
	return nil
}

func (l DecoderSinkList) Delete(uid types.UID) error {
	for _, d := range l {
		if err := d.Delete(uid); err != nil {
			return err
		}
	}
	return nil
}
