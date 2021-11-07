package slave

import (
	"fmt"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go/log"
	"os"
	"os/signal"
	"syscall"
)

type Canal struct {
	*canal.Canal
}

func Start(filePath string) (err error) {
	c, err := LoadFromFile(filePath)
	if err != nil {
		log.Infof("load from:%s err", filePath)
	}

	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", c.MysqlHost, c.MysqlPort)
	cfg.User = c.MysqlUser
	cfg.Password = c.MysqlPassword
	cfg.IncludeTableRegex = c.IncludeTableRegex

	if len(c.Mysqldump) > 0 {
		cfg.Dump.ExecutionPath = c.Mysqldump
	}

	cfg.Dump.DiscardErr = false
	cfg.Dump.SkipMasterData = c.SkipMasterData

	// 平滑重启
	isStop := make(chan os.Signal)
	signal.Notify(isStop, syscall.SIGUSR1)

	newCanal, err := canal.NewCanal(cfg)
	if err != nil {
		return
	}

	sync, err := NewSync(c.DataDir)
	if err != nil {
		return
	}

	newCanal.SetEventHandler(sync)

	// 消费 mysqlEvent
	go func() {
		for {
			select {
			case ev := <-sync.Event:
				switch ev := ev.(type) {
				case *canal.RowsEvent:
					//log.Infof("chan:%+v", ev.Table.String())
					destTable := ev.Table.String()
					if _, ok := c.Tables[destTable]; !ok {
						log.Infof("err table config: %s", destTable)
					}

					h := NewHandle(c, ev)
					err = h.Do(destTable)
					if err != nil {
						log.Errorf("error pushMq:%v", err)
					}
				case mysql.Position:
					err = sync.m.Save(ev)
					if err != nil {
						log.Errorf("save mysql.Position :%v", err)
					}
				}
			case <-isStop:
				log.Infof("receive signal isStop")
				sync.IsStop = true
				/*
					default:
					// 如果都消费完毕，需要关闭整个程序，重新拉起一个新的进行消费
					if len(sync.Event) == 0 && sync.IsStop {
						Start("config.json")
					}
				*/
			}

			if len(sync.Event) == 0 && sync.IsStop {
				log.Infof("newCanal isStop")
				Start(filePath)
				newCanal.Close()
				sync.IsStop = false
				break
			}
		}

	}()

	return newCanal.RunFrom(sync.m.Position())
}
