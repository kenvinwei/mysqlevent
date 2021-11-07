package slave

import (
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
)

type Sync struct {
	canal.DummyEventHandler
	m     *masterInfo
	Event chan interface{}
	IsStop bool
}

func (h *Sync) OnRow(e *canal.RowsEvent) error {
	//log.Infof("%s,%v, %+v\n", e.Action, e.Rows, e.Table.Columns)
	if !h.IsStop {
		h.Event <- e
	}
	//log.Infoln(e.Table.String())
	return nil
}

func (h *Sync) OnXID(nextPos mysql.Position) error {
	log.Infof("OnXID: %v, %v", nextPos, !h.IsStop)
	h.Event <- nextPos
	return nil
}

func (h *Sync) String() string {
	return "Sync"
}

func (h *Sync) OnPosSynced(p mysql.Position, gtid mysql.GTIDSet, force bool) error {
	//log.Infof("OnPosSynced: %v, %v", p, !h.IsStop)
	//if !h.IsStop {
	//	h.Event <- p
	//}
	return nil
}

func NewSync(dataDir string) (s *Sync, err error) {
	masterInfo, err := loadMasterInfo(dataDir)
	return &Sync{
		m: masterInfo,
		Event: make(chan interface{}, 4096),
	}, nil
}
