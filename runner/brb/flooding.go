package brb

import "fmt"

// Simple 'brb' (NOT BRB) testing protocol
type Flooding struct {
	n   Network
	app Application
	cfg Config

	seen map[uint32]struct{}
}

func (f *Flooding) Init(n Network, app Application, cfg Config) {
	f.n = n
	f.app = app
	f.cfg = cfg
	f.seen = make(map[uint32]struct{})

	if cfg.Byz {
		fmt.Printf("process %v is a Byzantine node\n", cfg.Id)
	}
}

func (f *Flooding) flood(uid uint32, data []byte, ex uint64) {
	for _, n := range f.cfg.Neighbours {
		if n != ex {
			f.n.Send(0, n, uid, data)
		}
	}
}

func (f *Flooding) Receive(_ uint8, src uint64, uid uint32, data []byte) {
	if f.cfg.Byz {
		return
	}

	if _, ok := f.seen[uid]; !ok {
		f.seen[uid] = struct{}{}
		f.app.Deliver(uid, data)

		f.flood(uid, data, src)
	}
}

func (f *Flooding) Send(uid uint32, payload []byte) {
	if _, ok := f.seen[uid]; !ok {
		f.seen[uid] = struct{}{}
		f.app.Deliver(uid, payload)

		f.flood(uid, payload, f.cfg.Id)
	}
}
