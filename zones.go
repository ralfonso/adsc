package adsc

import (
	"strings"
	"sync"
	"time"
)

const (
	timeoutDur = 30 * time.Second
)

type ZoneTracker struct {
	in           chan interface{}
	clientStopFn func()

	onFault   func(int)
	onRestore func(int)

	faultsmu sync.Mutex
	faults   map[int]time.Time
}

func NewZoneTracker(client *TCPClient, onFault, onRestore func(int)) (*ZoneTracker, error) {
	in, stop := client.Messages()
	zt := &ZoneTracker{
		in:           in,
		clientStopFn: stop,

		onFault:   onFault,
		onRestore: onRestore,

		faults: make(map[int]time.Time),
	}

	go zt.start()
	return zt, nil
}

func (z *ZoneTracker) Stop() {
	close(z.in)
	z.clientStopFn()
}

func (z *ZoneTracker) start() {
	for raw := range z.in {
		if msg, ok := raw.(Message); ok {
			if kmsg, ok := msg.(*KeypadMessage); ok {
				z.handleMessage(kmsg)
			}
		}
	}
}

func (z *ZoneTracker) handleMessage(kmsg *KeypadMessage) {
	if strings.Contains(kmsg.Message, "FAULT") {
		z.handleFault(kmsg)
	}

	if kmsg.Fields.Ready {
		z.clearZones()
	}

	z.expireZones()
}

func (z *ZoneTracker) clearZones() {
	z.faultsmu.Lock()
	defer z.faultsmu.Unlock()

	for zone, _ := range z.faults {
		delete(z.faults, zone)
		go z.onRestore(zone)
	}
}

func (z *ZoneTracker) handleFault(kmsg *KeypadMessage) {
	z.faultsmu.Lock()
	defer z.faultsmu.Unlock()
	z.faults[kmsg.Code.Value()] = time.Now().UTC().Add(timeoutDur)
	go z.onFault(kmsg.Code.Value())
}

func (z *ZoneTracker) expireZones() {
	z.faultsmu.Lock()
	defer z.faultsmu.Unlock()

	for zone, expiration := range z.faults {
		if time.Now().UTC().After(expiration) {
			delete(z.faults, zone)
			go z.onRestore(zone)
		}
	}
}
