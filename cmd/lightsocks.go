package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xmapst/lightsocks/internal/api"
	"github.com/xmapst/lightsocks/internal/conf"
	"github.com/xmapst/lightsocks/internal/dns"
	"github.com/xmapst/lightsocks/internal/mixed"
	"github.com/xmapst/lightsocks/internal/resolver"
	"github.com/xmapst/lightsocks/internal/server"
	"github.com/xmapst/lightsocks/internal/tunnel"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	s *server.Listener
	c *mixed.Listener
)

var (
	cmd = &cobra.Command{
		Use:               os.Args[0],
		Short:             "Support socks4, socks4a, socks5, socks5h, http proxy all in one,",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// load conf
			err := conf.Load()
			if err != nil {
				logrus.Fatalln(err)
			}
			tunnel.Start(conf.App.Local.Token)
			api.Server(conf.App.Api)
			conf.App.LoadTLS()
			// start socks server
			c = mixed.New()
			err = c.ListenAndServe()
			if err != nil {
				logrus.Fatalln(err)
			}
		},
	}
	serverCmd = &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			// load conf
			err := conf.Load()
			if err != nil {
				logrus.Fatalln(err)
			}
			tunnel.Start(conf.App.Local.Token)
			api.Server(conf.App.Api)
			conf.App.Mode = conf.ServerMode
			conf.App.LoadTLS()
			// start socks server
			s = server.New()
			err = s.ListenAndServe()
			if err != nil {
				logrus.Fatalln(err)
			}
		},
	}
	clientCmd = &cobra.Command{
		Use:               "client",
		Short:             "Support socks4, socks4a, socks5, socks5h, http proxy all in one.",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			// load conf
			err := conf.Load()
			if err != nil {
				logrus.Fatalln(err)
			}
			tunnel.Start(conf.App.Server.Token)
			api.Server(conf.App.Api)
			conf.App.Mode = conf.ClientMode
			if conf.App.Server.Port == 0 || conf.App.Server.Host == "" {
				conf.App.Mode = conf.DirectMode
				conf.App.TLS.Enable = false
			}
			conf.App.LoadTLS()
			// start socks server
			c = mixed.New()
			err = c.ListenAndServe()
			if err != nil {
				logrus.Fatalln(err)
			}
		},
	}
)

func init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		DisableColors:   true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			file = fmt.Sprintf("%s:%d", path.Base(frame.File), frame.Line)
			function = trimFunctionSuffix(frame.Function)
			return
		},
	})
	registerSignalHandlers()
	cmd.PersistentFlags().StringVarP(&conf.Path, "config", "c", "config.yaml", "config file path")
	cmd.AddCommand(serverCmd, clientCmd)
	resolver.DefaultResolver = dns.NewResolver(dns.Config{
		NameServers: []dns.NameServer{
			{
				Net:  "tcp",
				Addr: "119.29.29.29:53",
			},
			{
				Net:  "tcp",
				Addr: "119.28.28.28:53",
			},
			{
				Net:  "tcp",
				Addr: "223.5.5.5:53",
			},
			{
				Net:  "tcp",
				Addr: "223.6.6.6:53",
			},
			{
				Net:  "tcp",
				Addr: "1.0.0.1:53",
			},
			{
				Net:  "tcp",
				Addr: "1.1.1.1:53",
			},
			{
				Net:  "tcp",
				Addr: "8.8.8.8:53",
			},
			{
				Net:  "tcp",
				Addr: "8.8.4.4:53",
			},
		},
	})
}

func main() {
	cobra.CheckErr(cmd.Execute())
}

func registerSignalHandlers() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigs
		//err := ml.Shutdown()
		logrus.Infoln("received signal, exiting...")
		if s != nil {
			_ = s.ShutdownWithTimeout(time.Second * 15)
		}
		if c != nil {
			_ = c.ShutdownWithTimeout(time.Second * 15)
		}
		os.Exit(0)
	}()
}

func trimFunctionSuffix(s string) string {
	if strings.Contains(s, ".func") {
		index := strings.Index(s, ".func")
		s = s[:index]
	}
	s = strings.TrimSuffix(s, ".")
	slice := strings.Split(s, ".")
	return slice[len(slice)-1]
}
