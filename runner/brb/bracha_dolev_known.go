package brb

import (
	"fmt"
	"math"
	"rp-runner/brb/algo"
	"rp-runner/graphs"
)

type BrachaDolevKnown struct {
	bracha Protocol
	dolev  Protocol

	n   Network
	app Application
	cfg Config

	brachaBroadcast map[int]struct{}
}

var _ Protocol = (*BrachaDolevKnown)(nil)
var _ Network = (*BrachaDolevKnown)(nil)
var _ Application = (*BrachaDolevKnown)(nil)

type BrachaDolevConfig struct {
	Included []uint64
}

type BrachaDolevMessage struct {
	Src   uint64
	Id    uint32
	Type  uint8
	Paths []algo.DolevPath
}

type BrachaDolevWrapper struct {
	Msgs            []BrachaDolevMessage
	OriginalSrc     uint64
	OriginalId      uint32
	OriginalPayload interface{}
	Included        []uint64
}

// Unpack creates all DolevKnownImprovedMessage instances from the current BrachaDolevWrapper
func (b BrachaDolevWrapper) Unpack() []DolevKnownImprovedMessage {
	res := make([]DolevKnownImprovedMessage, 0, len(b.Msgs))

	for _, msg := range b.Msgs {
		bm := BrachaImprovedMessage{
			BrachaMessage: BrachaMessage{
				Src:     b.OriginalSrc,
				Id:      b.OriginalId,
				Payload: b.OriginalPayload,
			},
			Included: b.Included,
		}
		dm := DolevKnownImprovedMessage{
			Src: msg.Src,
			Id:  msg.Id,
			Payload: brachaWrapper{
				messageType: msg.Type,
				msg:         bm,
			},
			Paths: msg.Paths,
		}

		res = append(res, dm)
	}

	return res
}

// Pack fills the current BrachaDolevWrapper with information from a slice of DolevKnownImprovedMessage instances
func Pack(original []DolevKnownImprovedMessage) BrachaDolevWrapper {
	msgs := make([]BrachaDolevMessage, 0, len(original))
	bdw := BrachaDolevWrapper{}

	for _, msg := range original {
		bw := msg.Payload.(brachaWrapper)

		msgs = append(msgs, BrachaDolevMessage{
			Src:   msg.Src,
			Id:    msg.Id,
			Type:  bw.messageType,
			Paths: msg.Paths,
		})

		bm := bw.msg.(BrachaImprovedMessage)
		bdw.OriginalSrc = bm.Src
		bdw.OriginalId = bm.Id
		bdw.OriginalPayload = bm.Payload
		bdw.Included = bm.Included
	}

	bdw.Msgs = msgs
	return bdw
}

func (bd *BrachaDolevKnown) Init(n Network, app Application, cfg Config) {
	bd.n = n
	bd.app = app
	bd.cfg = cfg
	bd.brachaBroadcast = make(map[int]struct{})

	sil := cfg.Silent
	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Bracha-Dolev Byzantine node\n", cfg.Id)
	}

	cfg.Silent = true

	// Create bracha instance with BD as the network
	if bd.bracha == nil {
		bd.bracha = &BrachaImproved{}
	}

	bCfg := cfg
	nodes := cfg.Graph.Nodes()
	bCfg.Neighbours = make([]uint64, 0, nodes.Len())

	echoReq := int(math.Ceil((float64(cfg.N)+float64(cfg.F)+1)/2)) + cfg.F
	bCfg.AdditionalConfig = BrachaDolevConfig{
		Included: graphs.ClosestNodes(int64(cfg.Id), graphs.FindAdjMap(graphs.Directed(cfg.Graph), graphs.MaxId(cfg.Graph)), echoReq),
	}

	for nodes.Next() {
		i := uint64(nodes.Node().ID())
		if i != cfg.Id {
			bCfg.Neighbours = append(bCfg.Neighbours, i)
		}
	}

	bd.bracha.Init(bd, app, bCfg)

	// Create dolev (improved) instance with BD as the application
	if bd.dolev == nil {
		bd.dolev = &DolevKnownImprovedBD{}
	}
	cfg.AdditionalConfig = BrachaDolevConfig{}
	bd.dolev.Init(n, bd, cfg)

	cfg.Silent = sil
}

func (bd *BrachaDolevKnown) Send(messageType uint8, dest uint64, uid uint32, data interface{}, bc BroadcastInfo) {
	if _, ok := bd.brachaBroadcast[bc.Id]; !ok {
		// A message is broadcast only once to all
		bd.brachaBroadcast[bc.Id] = struct{}{}

		//if messageType == BrachaEcho {
		//	fmt.Printf("proc %v is sending echo\n", bd.cfg.Id)
		//} else if messageType == BrachaReady {
		//	fmt.Printf("proc %v is sending ready\n", bd.cfg.Id)
		//}

		partial := messageType == BrachaSend || messageType == BrachaEcho
		bcType := BrachaEveryone

		if partial {
			bcType = BrachaPartial
		}

		// Bracha is sending a message through Dolev
		bd.dolev.Broadcast(uid, brachaWrapper{
			messageType: messageType,
			msg:         data,
		}, BroadcastInfo{Type: bcType})
	}
}

func (bd *BrachaDolevKnown) Deliver(uid uint32, payload interface{}, src uint64) {
	// Dolev is delivering a message, so send it to Bracha
	m := payload.(brachaWrapper)

	if src == bd.cfg.Id {
		return
	}

	//if m.messageType == BrachaSend {
	//	fmt.Printf("proc %v got initial send\n", bd.cfg.Id)
	//} else if m.messageType == BrachaEcho {
	//	fmt.Printf("proc %v got echo from %v\n", bd.cfg.Id, src)
	//} else if m.messageType == BrachaReady {
	//	fmt.Printf("proc %v got ready from %v\n", bd.cfg.Id, src)
	//}

	bd.bracha.Receive(m.messageType, src, uid, m.msg)
}

func (bd *BrachaDolevKnown) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
	// Network is delivering a messages, pass to Dolev
	bd.dolev.Receive(0, src, uid, data)
}

func (bd *BrachaDolevKnown) Broadcast(uid uint32, payload interface{}, _ BroadcastInfo) {
	// Application is requesting a broadcast, pass to Bracha
	bd.bracha.Broadcast(uid, payload, BroadcastInfo{})
}

func (bd *BrachaDolevKnown) Category() ProtocolCategory {
	return BrachaDolevCat
}

func (bd *BrachaDolevKnown) TriggerStat(uid uint32, n NetworkStat) {
	bd.n.TriggerStat(uid, n)
}
