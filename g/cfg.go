package g

import (
	"os"
    "net"
    "log"
    "sync"
    "time"
    "strings"
    "encoding/json"

    "github.com/toolkits/file"
)

var (
	LocalIp string
    ConfigFile string
    config *GlobalConfig
    lock = new(sync.RWMutex)
)

type LogConfig struct {
    LogLevel string     //!< debug, info, warn, error, fatal 日志级别由低到高
    Output   string     //!< tty, file
    Type     string     //!< text, json
}

type HeartbeatConfig struct {
    Enabled  bool
    Addr     string
    Interval int
    Timeout  int
}

type TransferConfig struct {
    Enabled  bool
	Type	 string     //!< localFile, Redis, RPC, MQ
    Addrs    []string   //!< file,      ip:port
    Interval int        //!< 监控采集周期 
}

type GlobalConfig struct {
	Hostname		string
	Ip				string
    UnixSockFile    string
    Log             *LogConfig
    Heartbeat		*HeartbeatConfig
	Transfer		*TransferConfig
}

func InitLocalIp() {
    if Config().Transfer.Enabled {
        conn, err := net.DialTimeout("tcp", Config().Transfer.Addrs[0], time.Second*10)
        if err != nil {
            log.Println("get local addr failed !")
        } else {
            LocalIp = strings.Split(conn.LocalAddr().String(), ":")[0]
            conn.Close()
        }
    } else {
        log.Println("hearbeat is not enabled, can't get localip")
    }
}

func Config() *GlobalConfig {
    lock.RLock()
    defer lock.RUnlock()
    return config
}

func Hostname() (string, error) {
    hostname := Config().Hostname
    if hostname != "" {
        return hostname, nil
    }

    hostname, err := os.Hostname()
    if err != nil {
        log.Println("ERROR: os.Hostname() fail", err)
    }
    return hostname, err
}

func IP() string {
    ip := Config().Ip
    if ip != "" {
        // use ip in configuration
        return ip
    }

    if len(LocalIp) > 0 {
        ip = LocalIp
    }

    return ip
}

func ParseConfig(cfg string) {
    if cfg == "" {
        log.Fatalln("use -c to specify configuration file")
    }

    if !file.IsExist(cfg) {
        log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
    }

    ConfigFile = cfg

    configContent, err := file.ToTrimString(cfg)
    if err != nil {
        log.Fatalln("read config file:", cfg, "fail:", err)
    }

    var c GlobalConfig
    err = json.Unmarshal([]byte(configContent), &c)
    if err != nil {
        log.Fatalln("parse config file:", cfg, "fail:", err)
    }

    lock.Lock()
    defer lock.Unlock()

    config = &c

    log.Println("read config file:", cfg, "successfully")
}
