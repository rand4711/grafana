//go:build !windows && !nacl && !plan9
// +build !windows,!nacl,!plan9

package log

import (
	"log/syslog"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	gokitsyslog "github.com/go-kit/log/syslog"
	"gopkg.in/ini.v1"
)

type SysLogHandler struct {
	syslog   *syslog.Writer
	Network  string
	Address  string
	Facility string
	Tag      string
	Format   Formatedlogger
	logger   log.Logger
}

var selector = func(keyvals ...interface{}) syslog.Priority {
	for i := 0; i < len(keyvals); i += 2 {
		if keyvals[i] == level.Key() {
			if v, ok := keyvals[i+1].(string); ok {
				switch v {
				case "emergency":
					return syslog.LOG_EMERG
				case "alert":
					return syslog.LOG_ALERT
				case "critical":
					return syslog.LOG_CRIT
				case "error":
					return syslog.LOG_ERR
				case "warning":
					return syslog.LOG_WARNING
				case "notice":
					return syslog.LOG_NOTICE
				case "info":
					return syslog.LOG_INFO
				case "debug":
					return syslog.LOG_DEBUG
				}
				return syslog.LOG_LOCAL0
			}
		}
	}
	return syslog.LOG_LOCAL0
}

func NewSyslog(sec *ini.Section, format Formatedlogger) *SysLogHandler {
	handler := &SysLogHandler{}

	handler.Format = format
	handler.Network = sec.Key("network").MustString("")
	handler.Address = sec.Key("address").MustString("")
	handler.Facility = sec.Key("facility").MustString("local7")
	handler.Tag = sec.Key("tag").MustString("")

	if err := handler.Init(); err != nil {
		_ = level.Error(root).Log("Failed to init syslog log handler", "error", err)
		os.Exit(1)
	}
	handler.logger = gokitsyslog.NewSyslogLogger(handler.syslog, format, gokitsyslog.PrioritySelectorOption(selector))
	return handler
}

func (sw *SysLogHandler) Init() error {
	// the facility is the origin of the syslog message
	prio := parseFacility(sw.Facility)

	w, err := syslog.Dial(sw.Network, sw.Address, prio, sw.Tag)
	if err != nil {
		return err
	}

	sw.syslog = w
	return nil
}

func (sw *SysLogHandler) Log(keyvals ...interface{}) error {
	err := sw.logger.Log(keyvals...)
	return err
}

func (sw *SysLogHandler) Close() error {
	return sw.syslog.Close()
}

var facilities = map[string]syslog.Priority{
	"user":   syslog.LOG_USER,
	"daemon": syslog.LOG_DAEMON,
	"local0": syslog.LOG_LOCAL0,
	"local1": syslog.LOG_LOCAL1,
	"local2": syslog.LOG_LOCAL2,
	"local3": syslog.LOG_LOCAL3,
	"local4": syslog.LOG_LOCAL4,
	"local5": syslog.LOG_LOCAL5,
	"local6": syslog.LOG_LOCAL6,
	"local7": syslog.LOG_LOCAL7,
}

func parseFacility(facility string) syslog.Priority {
	v, found := facilities[facility]
	if !found {
		// default the facility level to LOG_LOCAL7
		return syslog.LOG_LOCAL7
	}
	return v
}
