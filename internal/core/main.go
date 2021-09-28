package core

import (
	"fmt"
	"github.com/bufsnake/aiqicha/config"
	"github.com/bufsnake/aiqicha/pkg/aiqicha"
	"github.com/bufsnake/aiqicha/pkg/log"
	"os"
	"os/signal"
	"sync"
	"time"
)

func Core(conf *config.Config) error {
	log.SetOutFile(conf.Output)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case _ = <-c:
			log.SaveData()
		}
	}()

	browser, err := aiqicha.NewBrowser(conf)
	if err != nil {
		return err
	}
	go browser.SwitchTab()
	defer browser.CloseBrowser()

	group := sync.WaitGroup{}
	first := true
	for company, _ := range conf.Targets {
		group.Add(1)
		flag := make(chan bool)
		go func() {
			defer group.Done()
			err = browser.Search(company)
			flag <- true
			if err != nil && first {
				first = false
				fmt.Println(err)
			}
		}()
		select {
		case <-time.After(5 * time.Minute):
		case <-flag:
		}
	}
	group.Wait()
	log.SaveData()
	return nil
}
