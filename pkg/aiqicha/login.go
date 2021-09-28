package aiqicha

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"os"
	"strings"
)

// 登录模块
// 登录后为用户名:
// body > div.base.page-index.has-search-tab > header > div > div.header-func > div.header-login > div > a
// body > div.base.page-index.has-search-tab > header > div > div.header-func > div.header-login > span.login
func (a *aiqicha) login() chromedp.Tasks {
	// 登录获取Cookie信息，保存到临时文件，关闭浏览器
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(bypassHeadlessDetect).Do(ctx)
			if err != nil {
				return errors.New(fmt.Sprintf("AddScriptToEvaluateOnNewDocument %s", err))
			}
			err = chromedp.Navigate("https://aiqicha.baidu.com/").Do(ctx)
			if err != nil {
				if !strings.Contains(err.Error(), "page load error net::ERR_INVALID_AUTH_CREDENTIALS") {
					errs := target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
					if errs != nil {
						return errors.New(fmt.Sprintf("Navigate CloseTarget %s", err))
					}
					return errors.New(fmt.Sprintf("Navigate Target Error %s", err))
				}
			}
			return nil
		}),
		chromedp.WaitVisible(`body > div.base.page-index.has-search-tab > header > div > div.header-func > div.header-login > div > a`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			defer target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
			cookies, err := network.GetAllCookies().Do(ctx)
			if err != nil {
				return err
			}
			data, err := network.GetAllCookiesReturns{Cookies: cookies}.MarshalJSON()
			if err != nil {
				return err
			}
			return os.WriteFile("/tmp/cookies.aiqicha", data, 0666)
		}),
	}
}
