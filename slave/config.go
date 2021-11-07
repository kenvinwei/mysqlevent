package slave

import (
	"encoding/json"
	"github.com/siddontang/go-log/log"
	"io/ioutil"
	"os"
)

type Config struct {
	MysqlHost        string `json:"mysql_host"`
	MysqlPort        int32  `json:"mysql_port"`
	MysqlUser        string `json:"mysql_user"`
	MysqlPassword    string `json:"mysql_password"`
	DataDir          string `json:"data_dir"`
	Mysqldump		 string `json:"mysqldump"`
	SkipMasterData   bool `json:"skip_master_data"`
	RabbitmqHost     string `json:"rabbitmq_host"`
	RabbitmqPort     int32  `json:"rabbitmq_port"`
	RabbitmqUser     string `json:"rabbitmq_user"`
	RabbitmqPassword string `json:"rabbitmq_password"`
	Tables           map[string]map[string]map[string]interface{} `json:"tables"`
	AMQP              map[string]AMQP
	RemoteUrl           map[string]RemoteUrl
	IncludeTableRegex []string
}

type AMQP struct {
	RabbitmqHost     string
	RabbitmqPort     int32
	RabbitmqUser     string
	RabbitmqPassword string
	Exchange         string
	RoutingKey       string
	Fields           map[string]string
}

type RemoteUrl struct {
	Url         string
	Method      string
	ContentType string
	Timeout     int
	Fields      map[string]string
}

func LoadFromFile(filePath string) (c *Config, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	conf, err := ioutil.ReadAll(file)
	if err != nil {
		log.Infof("Read file err, err =", err)
		return
	}

	defer file.Close()
	log.Infoln(string(conf))

	c = NewConfig()
	c.SetConfigByJson(string(conf))
	return c, nil
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) SetConfigByJson(config string) *Config {
	err := json.Unmarshal([]byte(config), c)
	if err != nil {
		panic("parser config.json error!")
	}

	// init
	c.AMQP = make(map[string]AMQP)
	c.RemoteUrl = make(map[string]RemoteUrl)

	for tk, tv := range c.Tables {
		// 只监听有配置的表
		c.IncludeTableRegex = append(c.IncludeTableRegex, tk)

		if _, ok := tv["send_amqp"]; ok {
			err := c.ParseAmqpConf(tk, tv["send_amqp"])
			if err != nil {
				panic(err)
			}
		}
		if _, ok := tv["remote_url"]; ok {
			err := c.ParseRemoteUrlConf(tk, tv["remote_url"])
			if err != nil {
				panic(err)
			}
		}
	}

	return c
}

func (c *Config) ParseAmqpConf(tableName string, tableAttr map[string]interface{}) error {
	fields := make(map[string]string)

	// 转化 fields 类型
	for fk, fv := range tableAttr["fields"].(map[string]interface{}) {
		fields[fk] = fv.(string)
	}

	amqpCnf := AMQP{
		RabbitmqHost:     tableAttr["rabbitmq_host"].(string),
		RabbitmqPort:     int32(tableAttr["rabbitmq_port"].(float64)),
		RabbitmqUser:     tableAttr["rabbitmq_user"].(string),
		RabbitmqPassword: tableAttr["rabbitmq_password"].(string),
		Exchange:         tableAttr["exchange"].(string),
		RoutingKey:       tableAttr["routing_key"].(string),
		Fields:           fields,
	}
	c.AMQP[tableName] = amqpCnf
	return nil
}

func (c *Config) ParseRemoteUrlConf(tableName string, tableAttr map[string]interface{}) error {
	fields := make(map[string]string)

	// 转化 fields 类型
	for fk, fv := range tableAttr["fields"].(map[string]interface{}) {
		fields[fk] = fv.(string)
	}

	remoteUrl := RemoteUrl{
		Url: tableAttr["url"].(string),
		Method: tableAttr["method"].(string),
		ContentType: tableAttr["content_type"].(string),
		Timeout: int(tableAttr["timeout"].(float64)),
		Fields: fields,
	}

	c.RemoteUrl[tableName] = remoteUrl

	return nil
}
