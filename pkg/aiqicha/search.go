package aiqicha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bufsnake/aiqicha/pkg/log"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"os"
	"strings"
	"time"
)

// 查询模块
func (a *aiqicha) Search(data string) error {
	tab_ctx, tab_cancel := chromedp.NewContext(a.ctx)
	defer tab_cancel()
	tab_ctx, tab_cancel = context.WithTimeout(tab_ctx, time.Duration(a.conf_.Timeout)*time.Second)
	defer tab_cancel()
	chromedp.ListenTarget(tab_ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			for i := 0; i < len(e.Args); i++ {
				var val string
				err := json.Unmarshal(e.Args[i].Value, &val)
				if err != nil {
					continue
				}
				switch {
				case strings.HasPrefix(val, "bufsnake control "):
					log.Control(data, val)
					break
				case strings.HasPrefix(val, "bufsnake invest "):
					log.Invest(data, val)
					break
				case strings.HasPrefix(val, "bufsnake business "):
					log.Business(data, val)
					break
				case strings.HasPrefix(val, "bufsnake shareholders "):
					log.Shareholders(data, val)
					break
				case strings.HasPrefix(val, "bufsnake webRecord "):
					log.WebRecord(data, val)
					break
				}
			}
		}
	})
	return chromedp.Run(tab_ctx, tasks(data))
}

func tasks(data string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(bypassHeadlessDetect).Do(ctx)
			if err != nil {
				return err
			}
			cookie_data, err := os.ReadFile("/tmp/cookies.aiqicha")
			if err != nil {
				return err
			}
			cookie_params := network.SetCookiesParams{}
			err = cookie_params.UnmarshalJSON(cookie_data)
			if err != nil {
				return err
			}
			err = network.SetCookies(cookie_params.Cookies).Do(ctx)
			if err != nil {
				return err
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
		chromedp.WaitVisible("#aqc-search-input"),
		// 判断是否登录
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					_ = target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
				}
			}()
			result, exception, err := runtime.Evaluate(`document.querySelector('body > div.base.page-index.has-search-tab > header > div > div.header-func > div.header-login > div > a').innerText`).Do(ctx)
			if err != nil {
				return err
			}
			if exception != nil {
				err = exception
				return
			}
			if !result.Value.IsDefined() {
				return errors.New("user is undefined ?")
			}
			return nil
		}),
		// 输入公司名，点击查询
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					_ = target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
				}
			}()
			return chromedp.SendKeys("#aqc-search-input", data).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`document.querySelector('body > div.base.page-index.has-search-tab > div.search-panel > div > div.index-search > div.index-search-input > button').click()`, nil).Do(ctx)
		}),
		// 等待查询结果
		chromedp.ActionFunc(func(ctx context.Context) error {
			fin := make(chan bool)
			go func() {
				_ = chromedp.WaitVisible("body > div.base.page-search.has-search-tab > div.aqc-content-wrapper.has-footer > div > div.main > div.list-wrap > div.header > span").Do(ctx)
				fin <- true
			}()
			select {
			case <-time.After(10 * time.Second):
				return errors.New("wait body > div.base.page-search.has-search-tab > div.aqc-content-wrapper.has-footer > div > div.main > div.list-wrap > div.header > span error")
			case <-fin:
				break
			}
			return nil
		}),
		chromedp.Sleep(1 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 判断是否为 0 家
			var res string
			err := chromedp.Evaluate(`document.querySelector('body > div.base.page-search.has-search-tab > div.aqc-content-wrapper.has-footer > div > div.main > div.list-wrap > div.header > span').innerText`, &res).Do(ctx)
			if err != nil {
				return err
			}
			if res == "0 " {
				return errors.New("not found companies")
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					_ = target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
				}
			}()
			var auto = `
var companys = document.querySelectorAll(
    'body > div.base.page-search.has-search-tab > div.aqc-content-wrapper.has-footer > div > div.main > div.list-wrap > div.company-list > div > div'
);
for (var i = 0; i < companys.length; i++) {
    if (companys[i].querySelector('div.info > div > h3 > a').innerText == "` + strings.ReplaceAll(strings.ReplaceAll(data, "）", ")"), "（", "(") + `" || companys[i].querySelector('div.info > div > h3 > a').innerText == "` + data + `") {
        companys[i].querySelector('div.info > div > h3 > a').target = '_self';
        companys[i].querySelector('div.info > div > h3 > a').click();
    }
}`
			_, exception, err := runtime.Evaluate(auto).Do(ctx)
			if err != nil {
				return err
			}
			if exception != nil {
				return exception
			}
			return
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fin := make(chan bool)
			go func() {
				_ = chromedp.WaitVisible("body > div.base.page-detail.has-search-tab > div.aqc-content-wrapper.has-footer > div > div.detail-header-container > div.detail-header > div.header-top > div.header-content > div.business-info > div.registered-capital.ellipsis-line-1 > span").Do(ctx)
				fin <- true
			}()
			select {
			case <-time.After(10 * time.Second):
				return errors.New("wait body > div.base.page-detail.has-search-tab > div.aqc-content-wrapper.has-footer > div > div.detail-header-container > div.detail-header > div.header-top > div.header-content > div.business-info > div.registered-capital.ellipsis-line-1 > span error")
			case <-fin:
				break
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fin := make(chan bool)
			go func() {
				_ = chromedp.WaitVisible("#basic-stockchart > h3 > span").Do(ctx)
				fin <- true
			}()
			select {
			case <-time.After(10 * time.Second):
				return errors.New("wait #basic-stockchart > h3 > span error")
			case <-fin:
				break
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fin := make(chan bool)
			go func() {
				_ = chromedp.WaitVisible("#basic-doubtcontroller > h3 > span").Do(ctx)
				fin <- true
			}()
			select {
			case <-time.After(10 * time.Second):
				return errors.New("wait #basic-doubtcontroller > h3 > span error")
			case <-fin:
				break
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, exception, err := runtime.Evaluate(`
const sleep = async (ms) => {
  return new Promise(resolve => setTimeout(resolve, ms));
}
`).Do(ctx)
			if exception != nil {
				err = exception
			}
			return err
		}),
		chromedp.Sleep(3 * time.Second),
		// 控股企业
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					_ = target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
				}
			}()
			// 页面必须包含: 控股企业 模块
			result, exception, err := runtime.Evaluate(`document.querySelector('#basic-hold > h3')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName != "HTMLHeadingElement" {
				return
			}
			// 控股企业
			var auto = `
				if (typeof eval('sleep') !== "function") {
					const sleep = async (ms) => {
					  return new Promise(resolve => setTimeout(resolve, ms));
					}
				}
				function getNext_Control() {
				   var tbody = document.querySelectorAll('#basic-hold > div > div > ul > li');
				   if (tbody.length == 0) {
				       return false
				   }
				   if (tbody[tbody.length - 1].className.indexOf('disabled') == -1) {
				       document.querySelector('#basic-hold > div > div > ul > li.ivu-page-next > a').click();
				       return true
				   }
				   return false
				}
		
				const control = async () => {
				   while (true) {
				       var tbody = document.querySelectorAll('#basic-hold > table > tbody > tr');
				       for (var i = 0; i < tbody.length; i++) {
				           var path = tbody[i].querySelectorAll('td:nth-child(4) > div');
				           for (var j = 0; j < path.length; j++) {
				               var info = path[j].querySelectorAll('ul > li');
				               var value = "";
				               for (var k = 0; k < info.length; k++) {
				                   var flag = false;
				                   if (info[k].querySelector("span") !== null) {
				                       flag = true;
				                       value += info[k].querySelector("span").innerText + " *-* ";
				                   }
				                   if (info[k].querySelector("a") !== null) {
				                       flag = true;
				                       value += info[k].querySelector("a").innerText + " *-* ";
				                   }
				                   if (!flag) {
				                       value += info[k].innerText + " *-* ";
				                   }
				               }
				               console.log("bufsnake control " + value.trim());
				           }
				       }
				       if (!getNext_Control()) {
				           break
				       }
				       await sleep(2000);
				   }
				}
				control()
				`
			_, exception, err = runtime.Evaluate(auto).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在多页则等待
			result, exception, err = runtime.Evaluate(`document.querySelector('#basic-hold > div > div > ul > li')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName == "HTMLLIElement" {
				// 下一页为display: none导致WaitVisible不可用
				// REF: https://github.com/chromedp/chromedp/issues/262
				return chromedp.WaitNotVisible("#basic-hold > div > div > ul > li.ivu-page-next").Do(ctx)
			}
			return
		}),
		chromedp.Sleep(2 * time.Second),
		// 对外投资
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					_ = target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
				}
			}()
			result, exception, err := runtime.Evaluate(`document.querySelector('#basic-invest > h3')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName != "HTMLHeadingElement" {
				return
			}

			var auto = `
		if (typeof eval('sleep') !== "function") {
			const sleep = async (ms) => {
			  return new Promise(resolve => setTimeout(resolve, ms));
			}
		}
		
		function getNext_invest() {
		  var tbody = document.querySelectorAll('#basic-invest > div > div > ul > li');
		  if (tbody.length == 0) {
		      return false
		  }
		  if (tbody[tbody.length - 1].className.indexOf('disabled') == -1) {
		      document.querySelector('#basic-invest > div > div > ul > li.ivu-page-next > a').click();
		      return true
		  }
		  return false
		}
		
		const invest = async () => {
		  while (true) {
		      var tbody = document.querySelectorAll('#basic-invest > table > tbody > tr');
		      for (var i = 0; i < tbody.length; i++) {
		          var path = tbody[i].querySelectorAll('td');
		          console.log("bufsnake invest "+path[1].querySelector('div > div.title.portrait-text > a.ellipsis-line-2').innerText +
		              " " + path[4].innerText + " " + path[6].innerText);
		      }
		      if (!getNext_invest()) {
                  console.log("对外投资获取完成")
		          break
		      }
		      await sleep(2000);
		  }
		}
		invest()
		`
			_, exception, err = runtime.Evaluate(auto).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在多页则等待
			result, exception, err = runtime.Evaluate(`document.querySelector('#basic-invest > div.aqc-table-list-pager')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName == "HTMLDivElement" {
				return chromedp.WaitNotVisible("#basic-invest > div.aqc-table-list-pager > div > ul > li.ivu-page-next").Do(ctx)
			}
			return
		}),
		chromedp.Sleep(2 * time.Second),
		//工商注册
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					_ = target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
				}
			}()
			result, exception, err := runtime.Evaluate(`document.querySelector('#basic-business > h3')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName != "HTMLHeadingElement" {
				return
			}
			var auto = `
var output = {};
var tbody = document.querySelectorAll('#basic-business > table > tbody > tr');
for (var i = 0; i < tbody.length; i++) {
   var path = tbody[i].querySelectorAll('td');
   for (var j = 0;j < path.length; j++) {
       if (j % 2 == 1) {
            output[path[j-1].innerText] = path[j].innerText.replace(/\n/g," ");
       }
   }
}
console.log("bufsnake business "+JSON.stringify(output))
`
			_, exception, err = runtime.Evaluate(auto).Do(ctx)
			if exception != nil {
				err = exception
			}
			return err
		}),
		chromedp.Sleep(2 * time.Second),
		// 股东信息
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					_ = target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
				}
			}()
			result, exception, err := runtime.Evaluate(`document.querySelector('#basic-shareholders > h3')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName != "HTMLHeadingElement" {
				return
			}

			var auto = `
if (typeof eval('sleep') !== "function") {
    const sleep = async (ms) => {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

function getNext_shareholders() {
    var tbody = document.querySelectorAll('#basic-shareholders > div > div > ul > li');
    if (tbody.length == 0) {
        return false
    }
    if (tbody[tbody.length - 1].className.indexOf('disabled') == -1) {
        document.querySelector('#basic-shareholders > div > div > ul > li.ivu-page-next > a').click();
        return true
    }
    return false
}

const shareholders = async () => {
    while (true) {
        var tbody = document.querySelectorAll('#basic-shareholders > table > tbody > tr');
        for (var i = 0; i < tbody.length; i++) {
            var path = tbody[i].querySelectorAll('td');
            console.log("bufsnake shareholders " + path[1].innerText.replaceAll('\n', " ").replaceAll("股权结构",
                " ").replaceAll(">", " ").trim().replaceAll(" ", "-") + " " + path[2].innerText);
        }
        if (!getNext_shareholders()) {
            break
        }
        await sleep(2000);
    }
}
shareholders()`
			_, exception, err = runtime.Evaluate(auto).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在多页则等待
			result, exception, err = runtime.Evaluate(`document.querySelector('#basic-shareholders > div.aqc-table-list-pager > div > ul')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName == "HTMLDivElement" {
				return chromedp.WaitNotVisible("#basic-shareholders > div.aqc-table-list-pager > div > ul > li.ivu-page-next").Do(ctx)
			}
			return
		}),
		chromedp.Sleep(2 * time.Second),
		chromedp.Click("body > div.base.page-detail.has-search-tab > div.aqc-content-wrapper.has-footer > div > div.tab-wrapper > div > div > div:nth-child(3) > a"),
		chromedp.Sleep(2 * time.Second),
		// 网站备案
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					_ = target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
				}
			}()
			result, exception, err := runtime.Evaluate(`document.querySelector('#certRecord-webRecord > h3')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName != "HTMLHeadingElement" {
				return
			}
			var auto = `
if (typeof eval('sleep') !== "function") {
	const sleep = async (ms) => {
	  return new Promise(resolve => setTimeout(resolve, ms));
	}
}

function getNext_webRecord() {
   var tbody = document.querySelectorAll('#certRecord-webRecord > div > div > ul > li');
   if (tbody.length == 0) {
	   return false
   }
   if (tbody[tbody.length - 1].className.indexOf('disabled') == -1) {
       document.querySelector('#certRecord-webRecord > div > div > ul > li.ivu-page-next > a').click();
       return true
   }
   return false
}

const webRecord = async () => {
   while (true) {
       var tbody = document.querySelectorAll('#certRecord-webRecord > table > tbody > tr');
       for (var i = 0; i < tbody.length; i++) {
           var path = tbody[i].querySelectorAll('td');
           console.log("bufsnake webRecord "+path[1].innerText.replaceAll("\n",";") + " " + path[2].innerText.replaceAll("\n",";") + " " + path[4].innerText.replaceAll("\n",";"));
       }
       if (!getNext_webRecord()) {
           break
       }
       await sleep(2000);
   }
}
webRecord()`
			_, exception, err = runtime.Evaluate(auto).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在多页则等待
			result, exception, err = runtime.Evaluate(`document.querySelector('#certRecord-webRecord > div > div > ul')`).Do(ctx)
			if exception != nil {
				err = exception
			}
			if err != nil {
				return err
			}
			// 存在: HTMLHeadingElement
			if result.ClassName == "HTMLDivElement" {
				return chromedp.WaitNotVisible("#certRecord-webRecord > div > div > ul > li.ivu-page-next").Do(ctx)
			}
			return
		}),
		chromedp.Sleep(10 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return target.CloseTarget(chromedp.FromContext(ctx).Target.TargetID).Do(cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Browser))
		}),
	}
}
