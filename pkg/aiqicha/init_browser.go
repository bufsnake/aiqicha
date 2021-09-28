package aiqicha

import (
	"context"
	"github.com/bufsnake/aiqicha/config"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"os"
	"time"
)

type aiqicha struct {
	ctx    context.Context
	cancel context.CancelFunc
	conf_  config.Config
}

// 初始化浏览器
func NewBrowser(conf_ *config.Config) (aiqicha, error) {
	aiqicha_ := aiqicha{conf_: *conf_}
	flags := []chromedp.ExecAllocatorOption{
		chromedp.Flag("block-new-web-contents", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-webgl", true),
		chromedp.Flag("disable-xss-auditor", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("headless", false),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("proxy-bypass-list", "<-loopback>"),
		chromedp.Flag("restore-on-startup", false),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.DisableGPU,
	}
	if conf_.Proxy != "" {
		flags = append(flags, chromedp.ProxyServer(conf_.Proxy))
	}
	if conf_.ChromePath != "" {
		flags = append(flags, chromedp.ExecPath(conf_.ChromePath))
	}

	login_option_ := append(flags, chromedp.Flag("headless", false))
	login_options := append(chromedp.DefaultExecAllocatorOptions[:], login_option_...)
	// start chrome login
	// 如果Cookie文件存在，则不进行登录
	_, err := os.ReadFile("/tmp/cookies.aiqicha")
	if err != nil {
		aiqicha_.ctx, aiqicha_.cancel = chromedp.NewExecAllocator(context.Background(), login_options...)
		aiqicha_.ctx, aiqicha_.cancel = chromedp.NewContext(aiqicha_.ctx)
		err = chromedp.Run(aiqicha_.ctx, aiqicha_.login())
		if err != nil {
			return aiqicha_, err
		}
	}
	// restart chrome
	restart_option_ := append(flags, chromedp.Flag("headless", !conf_.DisableHeadless))
	restart_options := append(chromedp.DefaultExecAllocatorOptions[:], restart_option_...)
	aiqicha_.ctx, aiqicha_.cancel = chromedp.NewExecAllocator(context.Background(), restart_options...)
	aiqicha_.ctx, aiqicha_.cancel = chromedp.NewContext(aiqicha_.ctx)
	err = chromedp.Run(aiqicha_.ctx, page.Close())
	if err != nil {
		return aiqicha_, err
	}
	return aiqicha_, nil
}

func (a *aiqicha) CloseBrowser() {
	defer a.cancel()
}

func (a *aiqicha) SwitchTab() {
	targets_ := make(map[target.ID]time.Time)
	for {
		targets, err := chromedp.Targets(a.ctx)
		if err != nil {
			continue
		}
		for i := 0; i < len(targets); i++ {
			if _, ok := targets_[targets[i].TargetID]; !ok {
				targets_[targets[i].TargetID] = time.Now()
			}
			err = target.ActivateTarget(targets[i].TargetID).Do(cdp.WithExecutor(a.ctx, chromedp.FromContext(a.ctx).Browser))
			if err != nil {
				continue
			}
			if time.Now().Sub(targets_[targets[i].TargetID]).Seconds() >= float64(a.conf_.Timeout) {
				_ = target.CloseTarget(targets[i].TargetID).Do(cdp.WithExecutor(a.ctx, chromedp.FromContext(a.ctx).Browser))
			}
		}
		time.Sleep(1 * time.Second / 5)
	}
}
