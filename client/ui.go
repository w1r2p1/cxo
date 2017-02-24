package client

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/skycoin/cxo/gui"
	"github.com/skycoin/cxo/node"
	"github.com/skycoin/skycoin/src/util"
)

const (
	logModuleName = "skyhash.main"
)

var (
	logger     = util.MustGetLogger(logModuleName)
	logModules = []string{logModuleName}
)

// default configurations
const (
	webInterfaceEnable = true
	webInterfacePort   = 6481
	webInterfaceAddr   = "127.0.0.1"
	webInterfaceHttps  = false
	launchBrowser      = true
	guiDirectory       = "gui/static/"
	dataDirectory      = ".skyhash"
)

// TODO:
//    Q: to move WebInterfaceConfig to its related package
//       (github.com/skycoin/skycoin/src/mesh/gui) ?
// Remote web interface
type WebInterfaceConfig struct {
	Enable bool
	Port   int
	Addr   string
	Cert   string
	Key    string
	HTTPS  bool
	// Launch system default browser after client startup
	LaunchBrowser bool
	// static htmls + assets location
	GUIDirectory string
}

type Config struct {
	// secret key
	SecretKey string
	// node configs
	NodeConfig node.Config
	// WebInterface configs
	WebInterface WebInterfaceConfig
	// Data directory holds app data -- defaults to ~/.skycoin
	DataDirectory string
	// logging configs
	Log *util.LogConfig
}

// TODO: defaultConfig or may be (developConfig + productConfig) ?

func defaultConfig() *Config {
	cfg := &Config{
		NodeConfig: node.NewConfig(),
		WebInterface: WebInterfaceConfig{
			Enable: webInterfaceEnable,
			Port:   webInterfacePort,
			Addr:   webInterfaceAddr,
			// Cert: "",
			// Key: "",
			HTTPS:         webInterfaceHttps,
			LaunchBrowser: launchBrowser,
			GUIDirectory:  guiDirectory,
		},
		// Data directory holds app data -- defaults to ~/.skycoin
		DataDirectory: dataDirectory,
		// TODO: dev/prod vs default?
		//       see src/util/logging.go 'TODO' note before DevLogConfig()
		//       for details
		Log: util.DevLogConfig(logModules),
	}
	return cfg
}

func (c *Config) Parse() {
	// obtain values from flags
	c.fromFlags()
	// post process
	c.DataDirectory = util.InitDataDir(c.DataDirectory)
	// if HTTPS is turned off then cerk/key are never used
	if c.WebInterface.HTTPS == true {
		if c.WebInterface.Cert == "" {
			c.WebInterface.Cert = filepath.Join(c.DataDirectory, "cert.pem")
		}
		if c.WebInterface.Key == "" {
			c.WebInterface.Key = filepath.Join(c.DataDirectory, "key.pem")
		}
	}

	// initialize logger
	c.Log.InitLogger()
}

func (c *Config) fromFlags() {
	c.NodeConfig.FromFlags()

	flag.StringVar(&c.SecretKey,
		"secret-key",
		"",
		"secret key of node")

	flag.BoolVar(&c.WebInterface.Enable, "web-interface", c.WebInterface.Enable,
		"enable the web interface")
	flag.IntVar(&c.WebInterface.Port, "web-interface-port",
		c.WebInterface.Port, "port to serve web interface on")
	flag.StringVar(&c.WebInterface.Addr, "web-interface-addr",
		c.WebInterface.Addr, "addr to serve web interface on")
	flag.StringVar(&c.WebInterface.Cert, "web-interface-cert",
		c.WebInterface.Cert, "cert.pem file for web interface HTTPS. "+
			"If not provided, will use cert.pem in -data-directory")
	flag.StringVar(&c.WebInterface.Key, "web-interface-key",
		c.WebInterface.Key, "key.pem file for web interface HTTPS. "+
			"If not provided, will use key.pem in -data-directory")
	flag.BoolVar(&c.WebInterface.HTTPS, "web-interface-https",
		c.WebInterface.HTTPS, "enable HTTPS for web interface")
	flag.BoolVar(&c.WebInterface.LaunchBrowser, "launch-browser",
		c.WebInterface.LaunchBrowser,
		"launch system default webbrowser at client startup")

	flag.StringVar(&c.DataDirectory, "data-dir", c.DataDirectory,
		"directory to store app data (defaults to ~/.skycoin)")

	flag.StringVar(&c.WebInterface.GUIDirectory, "gui-dir",
		c.WebInterface.GUIDirectory,
		"static content directory for the html gui")

	flag.StringVar(&c.Log.Level, "log-level", c.Log.Level,
		"Choices are: debug, info, notice, warning, error, critical")
	flag.BoolVar(&c.Log.Colors, "color-log", c.Log.Colors,
		"Add terminal colors to log output")

	flag.Parse()
}

// returns "http" ot "https";
// result depends on c.WebInterface.HTTPS
func (c *Config) scheme() string {
	if c.WebInterface.HTTPS == true {
		return "https"
	}
	return "http"
}

func RunAPI(c *Config, node node.Node, controllers ...gui.IRouterApi) {
	c.WebInterface.GUIDirectory = util.ResolveResourceDirectory(
		c.WebInterface.GUIDirectory)

	// print this message regardless of the log level
	fmt.Printf("Full address: %s\n", c.fullAddress())

	// start node_manager
	fmt.Printf("Starting Skyhash Manager Service...\n")

	if c.WebInterface.Enable == true {
		var err error
		if c.WebInterface.HTTPS == true {
			log.Panic("HTTPS support is not implemented yet")
		} else {
			err = gui.LaunchWebInterfaceAPI(c.host(),
				c.WebInterface.GUIDirectory,
				node,
				controllers...)
		}

		if err != nil {
			logger.Error(err.Error())
			logger.Error("Failed to start web GUI")
			os.Exit(1)
		}

		c.launchBrowser()
	}

	// subscribe to SIGINT (Ctrl+C)
	sigint := catchInterrupt()

	// waiting for SIGINT (Ctrl+C)
	logger.Info("Got signal %q, shutting down...", <-sigint)

	logger.Info("Goodbye")
}

// returns c.WebInterfaceAddr:c.WebInterface.Port string
func (c *Config) host() string {
	return fmt.Sprintf("%s:%d", c.WebInterface.Addr, c.WebInterface.Port)
}

// returns full address: scheme://address:port
func (c *Config) fullAddress() string {
	return fmt.Sprintf("%s://%s", c.scheme(), c.host())
}

// launches browser if it's enabled
func (c *Config) launchBrowser() {
	if c.WebInterface.LaunchBrowser == true {
		go func() {
			// TODO: wait is BS, is it really needed?
			//
			// Wait a moment just to make sure the http interface is up
			time.Sleep(time.Millisecond * 100)
			//

			logger.Info("Launching System Browser with %s", c.fullAddress())
			if err := util.OpenBrowser(c.fullAddress()); err != nil {
				logger.Error(err.Error())
			}
		}()
	}
}

// subscribe to SIGINT
func catchInterrupt() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	return sig
}
