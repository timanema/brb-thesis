package brb

import "fmt"

// Simple 'brb' (NOT BRB) testing protocol
type Flooding struct {
	n   Network
	app Application
	cfg Config

	seen map[uint32]struct{}
}

var _ Protocol = (*Flooding)(nil)

func (f *Flooding) Init(n Network, app Application, cfg Config) {
	f.n = n
	f.app = app
	f.cfg = cfg
	f.seen = make(map[uint32]struct{})

	if cfg.Byz {
		fmt.Printf("process %v is a flooding Byzantine node\n", cfg.Id)
	}
}

func (f *Flooding) flood(uid uint32, data Size, ex uint64) {
	for _, n := range f.cfg.Neighbours {
		if n != ex {
			f.n.Send(0, n, uid, data, BroadcastInfo{})
		}
	}
}

func (f *Flooding) Receive(_ uint8, src uint64, uid uint32, data Size) {
	if f.cfg.Byz {
		return
	}

	if _, ok := f.seen[uid]; !ok {
		f.seen[uid] = struct{}{}
		f.app.Deliver(uid, data, src)

		f.flood(uid, data, src)
	}
}

func (f *Flooding) Broadcast(uid uint32, payload Size, _ BroadcastInfo) {
	if _, ok := f.seen[uid]; !ok {
		f.seen[uid] = struct{}{}
		f.app.Deliver(uid, payload, 0)

		f.flood(uid, payload, f.cfg.Id)
	}
}

func (f *Flooding) Category() ProtocolCategory {
	return BrachaCat
}
