package config

import (
	"os"
	"strconv"

	"github.com/spf13/viper"
	"log"
	"test.com/project-common/logs"
)

var C = InitConfig()

type Config struct {
	viper       *viper.Viper
	SC          *ServerConfig
	GC          *GrpcConfig
	EtcdConfig  *EtcdConfig
	MysqlConfig *MysqlConfig
	RpcConfig   *RpcConfig
	RedisConfig *RedisConfig
	MailConfig  *MailConfig
}

type ServerConfig struct {
	Name string
	Addr string
}

type GrpcConfig struct {
	Name string
	Addr string
}

type EtcdConfig struct {
	Addrs []string
}

type MysqlConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Db       string
}

type RpcConfig struct {
	UserAddr    string
	ProjectAddr string
}

type RedisConfig struct {
	Addr     string
	Password string
	Db       int
}

type MailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

func InitConfig() *Config {
	conf := &Config{viper: viper.New()}
	workDir, _ := os.Getwd()
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	conf.viper.AddConfigPath("/etc/ms_project/user")
	conf.viper.AddConfigPath(workDir + "/config")
	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	conf.ReadServerConfig()
	conf.InitZapLog()
	conf.ReadEtcdConfig()
	conf.ReadRpcConfig()
	conf.InitMysqlConfig()
	conf.ReadRedisConfig()
	conf.ReadMailConfig()
	return conf
}

func (c *Config) InitZapLog() {
	//从配置中读取日志配置，初始化日志
	lc := &logs.LogConfig{
		DebugFileName: c.viper.GetString("zap.debugFileName"),
		InfoFileName:  c.viper.GetString("zap.infoFileName"),
		WarnFileName:  c.viper.GetString("zap.warnFileName"),
		MaxSize:       c.viper.GetInt("maxSize"),
		MaxAge:        c.viper.GetInt("maxAge"),
		MaxBackups:    c.viper.GetInt("maxBackups"),
	}
	err := logs.InitLogger(lc)
	if err != nil {
		log.Fatalln(err)
	}
}

func (c *Config) ReadServerConfig() {
	sc := &ServerConfig{}
	sc.Name = c.viper.GetString("server.name")
	sc.Addr = c.viper.GetString("server.addr")
	c.SC = sc
}

func (c *Config) ReadEtcdConfig() {
	ec := &EtcdConfig{}
	var addrs []string
	err := c.viper.UnmarshalKey("etcd.addrs", &addrs)
	if err != nil {
		log.Fatalln(err)
	}
	ec.Addrs = addrs
	c.EtcdConfig = ec
}

func (c *Config) ReadRpcConfig() {
	rc := &RpcConfig{}
	rc.UserAddr = c.viper.GetString("rpc.userAddr")
	rc.ProjectAddr = c.viper.GetString("rpc.projectAddr")
	c.RpcConfig = rc
}

func (c *Config) InitMysqlConfig() {
	mc := &MysqlConfig{
		Username: c.viper.GetString("mysql.username"),
		Password: c.viper.GetString("mysql.password"),
		Host:     c.viper.GetString("mysql.host"),
		Port:     c.viper.GetInt("mysql.port"),
		Db:       c.viper.GetString("mysql.db"),
	}
	c.MysqlConfig = mc
}

func (c *Config) ReadRedisConfig() {
	rc := &RedisConfig{
		Addr:     c.viper.GetString("redis.addr"),
		Password: c.viper.GetString("redis.password"),
		Db:       c.viper.GetInt("redis.db"),
	}
	if rc.Addr != "" {
		c.RedisConfig = rc
	}
}

func (c *Config) ReadMailConfig() {
	mc := &MailConfig{
		Host:     c.viper.GetString("mail.host"),
		Port:     c.viper.GetInt("mail.port"),
		User:     c.viper.GetString("mail.user"),
		Password: c.viper.GetString("mail.password"),
		From:     c.viper.GetString("mail.from"),
	}
	// 支持环境变量覆盖邮件配置，便于Docker部署和不同环境切换
	if v := os.Getenv("MAIL_HOST"); v != "" {
		mc.Host = v
	}
	if v := os.Getenv("MAIL_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			mc.Port = port
		}
	}
	if v := os.Getenv("MAIL_USER"); v != "" {
		mc.User = v
	}
	if v := os.Getenv("MAIL_PASSWORD"); v != "" {
		mc.Password = v
	}
	if v := os.Getenv("MAIL_FROM"); v != "" {
		mc.From = v
	}
	// 只有配置了有效的SMTP主机才启用邮件服务
	if mc.Host != "" && mc.Password != "" && mc.Password != "placeholder_replace_with_real_auth_code" {
		c.MailConfig = mc
	}
}
