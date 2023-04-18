package conf

import (
	"crypto/tls"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var (
	App       *Config
	Path      string
	logOutput *lumberjack.Logger
)

const (
	DirectMode = iota
	ClientMode
	ServerMode
)

type Config struct {
	Local  Server `yaml:""` // 服务端及客户端监听的本地端口
	Server Server `yaml:""` // 远端服务器地址
	Api    Server `yaml:""` // RESTful API
	TLS    TLS    `yaml:""` // 证书
	// 可动态配置
	Timeout time.Duration `yaml:""` // 连接超时时间
	CIDR    []string      `yaml:""` // 服务端或客户端使用的ip白名单
	Users   []User        `yaml:""` // 客户端的sock(s)/http认证
	Log     Log           `yaml:""` // 日志输出

	// self
	Mode    int
	TLSConf *tls.Config
}

type TLS struct {
	Enable bool   `yaml:""`
	Key    string `yaml:""`
	Cert   string `yaml:""`
}

type Server struct {
	Host  string `yaml:""`
	Port  int64  `yaml:""`
	Token string `yaml:""`
}

type User struct {
	UserName string
	Password string
	CIDR     []string
}

type Log struct {
	Filename   string `yaml:""`
	Level      string `yaml:",default=info"`
	MaxBackups int    `yaml:",default=7"`
	MaxSize    int    `yaml:",default=500"`
	MaxAge     int    `yaml:",default=28"`
	Compress   bool   `yaml:",default=true"`
}

func viperLoadConf() error {
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	var conf = &Config{
		Mode: DirectMode,
		TLSConf: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS13,
			CurvePreferences: []tls.CurveID{
				tls.CurveP521,
				tls.CurveP384,
				tls.CurveP256,
			},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
		Log: Log{
			Level:      "info",
			MaxBackups: 7,
			MaxSize:    500,
			MaxAge:     28,
			Compress:   true,
		},
	}
	err = viper.Unmarshal(conf)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	App = conf
	return nil
}

func Load() error {
	viper.SetConfigFile(Path)
	err := viperLoadConf()
	if err != nil {
		return err
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if !e.Has(fsnotify.Write) {
			return
		}
		logrus.Infoln(e.Name, "config file modified")
		err = viperLoadConf()
		if err != nil {
			logrus.Warningln(err)
			return
		}
		err = App.reload()
		if err != nil {
			logrus.Warningln(err)
			return
		}
	})

	err = App.reload()
	if err != nil {
		return err
	}
	c := cron.New()
	_, _ = c.AddFunc("@daily", func() {
		if logOutput != nil {
			_ = logOutput.Rotate()
		}
	})
	c.Start()
	return nil
}

func (c *Config) LoadTLS() {
	if !c.TLS.Enable {
		return
	}
	cer, err := tls.LoadX509KeyPair(c.TLS.Cert, c.TLS.Key)
	if err != nil {
		if c.Mode != ServerMode {
			return
		}
		logrus.Fatalln(err)
	}

	c.TLSConf = &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cer},
	}
}

func (c *Config) reload() error {
	level, err := logrus.ParseLevel(c.Log.Level)
	if err != nil {
		logrus.Warningln(err)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	if c.Log.Filename != "" {
		logOutput = &lumberjack.Logger{
			Filename:   c.Log.Filename,
			MaxBackups: c.Log.MaxBackups,
			MaxSize:    c.Log.MaxSize,  // megabytes
			MaxAge:     c.Log.MaxAge,   // days
			Compress:   c.Log.Compress, // disabled by default
			LocalTime:  true,           // use local time zone
		}
		logrus.SetOutput(logOutput)
	} else {
		logOutput = nil
		logrus.SetOutput(os.Stdout)
	}
	return nil
}
