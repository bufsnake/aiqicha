package main

import (
	"flag"
	"fmt"
	"github.com/bufsnake/aiqicha/config"
	"github.com/bufsnake/aiqicha/internal/core"
	"os"
	"strings"
	"time"
)

func main() {
	conf := config.Config{}
	flag.StringVar(&conf.ChromePath, "chrome-path", "", "chrome path")
	flag.StringVar(&conf.Target, "target", "", "target")
	flag.StringVar(&conf.TargetList, "target-list", "", "target list")
	flag.StringVar(&conf.Proxy, "proxy", "", "proxy")
	flag.StringVar(&conf.Output, "output", time.Now().Format("2006_01_02_15_04_05"), "output file name")
	flag.IntVar(&conf.Timeout, "timeout", 3600, "timeout")
	flag.BoolVar(&conf.DisableHeadless, "disable-headless", false, "disable chrome headless model")
	flag.Parse()
	conf.Targets = make(map[string]bool)
	if conf.Target != "" {
		conf.Targets[conf.Target] = true
	} else if conf.TargetList != "" {
		file, err := os.ReadFile(conf.TargetList)
		if err != nil {
			fmt.Println(err)
			return
		}
		split := strings.Split(string(file), "\n")
		for i := 0; i < len(split); i++ {
			split[i] = strings.Trim(split[i], "\t\r ")
			if split[i] == "" {
				continue
			}
			conf.Targets[split[i]] = true
		}
	} else {
		flag.Usage()
		return
	}
	if conf.Output == "" {
		conf.Output = time.Now().Format("2006_01_02_15_04_05")
	}
	err := core.Core(&conf)
	if err != nil {
		fmt.Println(err)
	}
}
