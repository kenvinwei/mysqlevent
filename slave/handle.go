package slave

import (
	"github.com/ddliu/go-httpclient"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/go-mysql/canal"
	"io/ioutil"
	"strings"
)

const (
	SubmitTypeForm = "form"
	SubmitTypeJson = "json"
	MethodGet = "get"
	MethodPost = "post"
)

type Handle struct {
	Config *Config
	Ev     *canal.RowsEvent
}

func NewHandle(Config *Config, ev *canal.RowsEvent) *Handle {
	return &Handle{Config, ev}
}

func (h *Handle) Do(tableName string) (err error) {
	amqpResCh := make(chan bool)
	callRemoteResCh := make(chan bool)
	if _,ok := h.Config.AMQP[tableName]; ok {
		go func() {
			amqpCnf := h.Config.AMQP[tableName]
			err = h.SendAmqp(amqpCnf)
			amqpResCh <- true
		}()
	} else {
		log.Infof("not found amqp : %s of remote url cfg", tableName)
		close(amqpResCh)
	}

	if _,ok := h.Config.RemoteUrl[tableName]; ok {
		go func() {
			remoteUrlCnf := h.Config.RemoteUrl[tableName]
			err = h.CallRemoteUrl(remoteUrlCnf)
			callRemoteResCh <- true
		}()
	} else {
		close(callRemoteResCh)
		log.Infof("not found table : %s of remote url cfg", tableName)
	}

	log.Infof("mqRes:%v", <-amqpResCh)
	log.Infof("urlRes:%v", <-callRemoteResCh)

	return
}

func (h *Handle) SendAmqp(amqpConf AMQP) (err error) {
	ev := h.Ev

	fields := amqpConf.Fields

	//log.Infof("rows: %+v", ev.Rows)
	cli, err := NewClient(amqpConf.RabbitmqHost, amqpConf.RabbitmqPort, amqpConf.RabbitmqUser, amqpConf.RabbitmqPassword)
	if err != nil {
		return
	}

	data := ev.Rows[0]
	if ev.Action == canal.UpdateAction {
		data = ev.Rows[1]
	}

	body := make(map[string]interface{})
	body["action"] = ev.Action
	for ci, cv := range fields {
		columnValue, err := ev.Table.GetColumnValue(ci, data)
		if err != nil {
			continue
		}
		body[cv] = columnValue
	}

	//log.Infof("body==========:%v", body)

	err = cli.pushMq(amqpConf.Exchange, amqpConf.RoutingKey, body)
	return
}

func (h *Handle) CallRemoteUrl(remoteUrlCnf RemoteUrl) (err error) {
	var res *httpclient.Response

	url := remoteUrlCnf.Url
	method := strings.ToLower(remoteUrlCnf.Method)
	contentType := remoteUrlCnf.ContentType
	fields := remoteUrlCnf.Fields


	fields["action"] = h.Ev.Action

	//log.Infof("timeout:%t", remoteUrlCnf.Timeout)
	cli := httpclient.NewHttpClient().WithOption(httpclient.OPT_TIMEOUT, remoteUrlCnf.Timeout)

	if method == MethodGet {
		res, err := cli.Get(url, fields)
		if err != nil {
			return err
		}
		result, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		log.Infof(" url : %s, args: %v, call remote url res:%s",
			url,
			fields,
			string(result),
		)
		return nil
	}


	switch contentType {
	case SubmitTypeForm:
			res, err = cli.Post(url, fields)
		break
	case SubmitTypeJson:
			res, err = cli.PostJson(url, fields)
		break
	}

	result, _ := ioutil.ReadAll(res.Body)
	log.Infof("call remote url res:%s", result)

	return
}
