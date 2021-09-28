package log

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
)

var name_ string
var data_file *excelize.File
var business []business_struct
var shareholders []shareholders_struct
var invest []invest_struct
var control [][]string
var webRecord []webRecord_struct

type business_struct struct {
	LegalRepresentative          string `json:"法定代表人"`
	OperatingStatus              string `json:"经营状态"`
	RegisteredCapital            string `json:"注册资本"`
	PaidCapital                  string `json:"实缴资本"`
	UsedName                     string `json:"曾用名"`
	Industry                     string `json:"所属行业"`
	UnifiedSocialCreditCode      string `json:"统一社会信用代码"`
	TaxpayerIdentificationNumber string `json:"纳税人识别号"`
	BusinessRegistrationNumber   string `json:"工商注册号"`
	OrganizationCode             string `json:"组织机构代码"`
	RegistrationAuthority        string `json:"登记机关"`
	EstablishedDate              string `json:"成立日期"`
	EnterpriseType               string `json:"企业类型"`
	BusinessPeriod               string `json:"营业期限"`
	AdministrativeDivisions      string `json:"行政区划"`
	ApprovedDate                 string `json:"核准日期"`
	NumberOfParticipants         string `json:"参保人数"`
	RegisteredAddress            string `json:"注册地址"`
	BusinessScope                string `json:"经营范围"`
}
type shareholders_struct struct {
	currEnt string
	entName string // 公司名
	regRate string // 占比
}
type invest_struct struct {
	currEnt    string
	entName    string // 公司名
	regRate    string // 占比
	openStatus string // 状态
}
type webRecord_struct struct {
	currEnt  string
	site     string
	siteName string
	icp      string
}

func SetOutFile(name string) {
	name_ = name
	data_file = excelize.NewFile()
	// 创建一个工作表
	business_ := data_file.NewSheet("工商注册")
	_ = data_file.NewSheet("股东信息")
	_ = data_file.NewSheet("对外投资")
	_ = data_file.NewSheet("控股企业")
	_ = data_file.NewSheet("网站备案")
	// 设置单元格的值
	_ = data_file.SetCellValue("工商注册", "A1", "法定代表人")
	_ = data_file.SetCellValue("工商注册", "B1", "经营状态")
	_ = data_file.SetCellValue("工商注册", "C1", "注册资本")
	_ = data_file.SetCellValue("工商注册", "D1", "实缴资本")
	_ = data_file.SetCellValue("工商注册", "E1", "曾用名")
	_ = data_file.SetCellValue("工商注册", "F1", "所属行业")
	_ = data_file.SetCellValue("工商注册", "G1", "统一社会信用代码")
	_ = data_file.SetCellValue("工商注册", "H1", "纳税人识别号")
	_ = data_file.SetCellValue("工商注册", "I1", "工商注册号")
	_ = data_file.SetCellValue("工商注册", "J1", "组织机构代码")
	_ = data_file.SetCellValue("工商注册", "K1", "登记机关")
	_ = data_file.SetCellValue("工商注册", "L1", "成立日期")
	_ = data_file.SetCellValue("工商注册", "M1", "企业类型")
	_ = data_file.SetCellValue("工商注册", "N1", "营业期限")
	_ = data_file.SetCellValue("工商注册", "O1", "行政区划")
	_ = data_file.SetCellValue("工商注册", "P1", "核准日期")
	_ = data_file.SetCellValue("工商注册", "Q1", "参保人数")
	_ = data_file.SetCellValue("工商注册", "R1", "注册地址")
	_ = data_file.SetCellValue("工商注册", "S1", "经营范围")

	_ = data_file.SetCellValue("股东信息", "A1", "当前企业")
	_ = data_file.SetCellValue("股东信息", "B1", "发起人/股东")
	_ = data_file.SetCellValue("股东信息", "C1", "持股比例")

	_ = data_file.SetCellValue("对外投资", "A1", "当前企业")
	_ = data_file.SetCellValue("对外投资", "B1", "被投资企业")
	_ = data_file.SetCellValue("对外投资", "C1", "投资占比")
	_ = data_file.SetCellValue("对外投资", "D1", "状态")

	// 根据控股链最长的定义
	//_ = data_file.SetCellValue("控股企业", "A1", "企业名")
	//_ = data_file.SetCellValue("控股企业", "B1", "占比")

	_ = data_file.SetCellValue("网站备案", "A1", "当前企业")
	_ = data_file.SetCellValue("网站备案", "B1", "首页地址")
	_ = data_file.SetCellValue("网站备案", "C1", "网站名称")
	_ = data_file.SetCellValue("网站备案", "D1", "备案号")

	// 设置工作簿的默认工作表
	data_file.SetActiveSheet(business_)
	data_file.DeleteSheet("Sheet1")
}

var chars = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

func SaveData() {
	// 控股企业 - 保存数据
	for i := 0; i < len(control); i++ {
		for j := 0; j < len(control[i]); j++ {
			cell := ""
			if j/26 < 1 {
				cell = chars[j]
			} else {
				cell = chars[(j/26)-1]
				cell += chars[j%26]
			}
			if j%2 == 0 {
				_ = data_file.SetCellValue("控股企业", cell+"1", "企业名")
			} else {
				_ = data_file.SetCellValue("控股企业", cell+"1", "占比")
			}
			_ = data_file.SetCellValue("控股企业", cell+strconv.Itoa(i+2), control[i][j])
		}
	}
	for i := 0; i < len(invest); i++ {
		_ = data_file.SetCellValue("对外投资", "A"+strconv.Itoa(i+2), invest[i].currEnt)
		_ = data_file.SetCellValue("对外投资", "B"+strconv.Itoa(i+2), invest[i].entName)
		_ = data_file.SetCellValue("对外投资", "C"+strconv.Itoa(i+2), invest[i].regRate)
		_ = data_file.SetCellValue("对外投资", "D"+strconv.Itoa(i+2), invest[i].openStatus)
	}
	for i := 0; i < len(shareholders); i++ {
		_ = data_file.SetCellValue("股东信息", "A"+strconv.Itoa(i+2), shareholders[i].currEnt)
		_ = data_file.SetCellValue("股东信息", "B"+strconv.Itoa(i+2), shareholders[i].entName)
		_ = data_file.SetCellValue("股东信息", "C"+strconv.Itoa(i+2), shareholders[i].regRate)
	}
	for i := 0; i < len(webRecord); i++ {
		_ = data_file.SetCellValue("网站备案", "A"+strconv.Itoa(i+2), webRecord[i].currEnt)
		_ = data_file.SetCellValue("网站备案", "B"+strconv.Itoa(i+2), webRecord[i].site)
		_ = data_file.SetCellValue("网站备案", "C"+strconv.Itoa(i+2), webRecord[i].siteName)
		_ = data_file.SetCellValue("网站备案", "D"+strconv.Itoa(i+2), webRecord[i].icp)
	}
	for i := 0; i < len(business); i++ {
		_ = data_file.SetCellValue("工商注册", "A"+strconv.Itoa(i+2), business[i].LegalRepresentative)
		_ = data_file.SetCellValue("工商注册", "B"+strconv.Itoa(i+2), business[i].OperatingStatus)
		_ = data_file.SetCellValue("工商注册", "C"+strconv.Itoa(i+2), business[i].RegisteredCapital)
		_ = data_file.SetCellValue("工商注册", "D"+strconv.Itoa(i+2), business[i].PaidCapital)
		_ = data_file.SetCellValue("工商注册", "E"+strconv.Itoa(i+2), business[i].UsedName)
		_ = data_file.SetCellValue("工商注册", "F"+strconv.Itoa(i+2), business[i].Industry)
		_ = data_file.SetCellValue("工商注册", "G"+strconv.Itoa(i+2), business[i].UnifiedSocialCreditCode)
		_ = data_file.SetCellValue("工商注册", "H"+strconv.Itoa(i+2), business[i].TaxpayerIdentificationNumber)
		_ = data_file.SetCellValue("工商注册", "I"+strconv.Itoa(i+2), business[i].BusinessRegistrationNumber)
		_ = data_file.SetCellValue("工商注册", "J"+strconv.Itoa(i+2), business[i].OrganizationCode)
		_ = data_file.SetCellValue("工商注册", "K"+strconv.Itoa(i+2), business[i].RegistrationAuthority)
		_ = data_file.SetCellValue("工商注册", "L"+strconv.Itoa(i+2), business[i].EstablishedDate)
		_ = data_file.SetCellValue("工商注册", "M"+strconv.Itoa(i+2), business[i].EnterpriseType)
		_ = data_file.SetCellValue("工商注册", "N"+strconv.Itoa(i+2), business[i].BusinessPeriod)
		_ = data_file.SetCellValue("工商注册", "O"+strconv.Itoa(i+2), business[i].AdministrativeDivisions)
		_ = data_file.SetCellValue("工商注册", "P"+strconv.Itoa(i+2), business[i].ApprovedDate)
		_ = data_file.SetCellValue("工商注册", "Q"+strconv.Itoa(i+2), business[i].NumberOfParticipants)
		_ = data_file.SetCellValue("工商注册", "R"+strconv.Itoa(i+2), business[i].RegisteredAddress)
		_ = data_file.SetCellValue("工商注册", "S"+strconv.Itoa(i+2), business[i].BusinessScope)
	}
	// 根据指定路径保存文件
	if err := data_file.SaveAs(name_ + ".xlsx"); err != nil {
		fmt.Println(err)
	}
}

// 公司名 - 数据
func Control(name, data string) {
	split := strings.Split(strings.ReplaceAll(data, "bufsnake control ", ""), "*-*")
	structs := make([]string, 0)
	output := ""
	for i := 0; i < len(split); i += 1 {
		if strings.Trim(split[i], " ") == "" {
			continue
		}
		if i%2 == 0 {
			output += strings.Trim(split[i], " ") + " - "
		} else {
			output += strings.Trim(split[i], " ") + " -> "
		}
		structs = append(structs, strings.Trim(split[i], " "))
	}
	fmt.Println("控股企业", strings.Trim(output, " -"))
	control = append(control, structs)
}

func Invest(name, data string) {
	data = strings.ReplaceAll(data, "bufsnake invest ", "")
	split := strings.Split(data, " ")
	if len(split) == 3 {
		invest = append(invest, invest_struct{
			currEnt:    name,
			entName:    split[0],
			regRate:    split[1],
			openStatus: split[2],
		})
		fmt.Println(name, "投资", split[0], split[1], split[2])
	}
}

func Business(name, data string) {
	data = strings.ReplaceAll(data, "bufsnake business ", "")
	businessStruct := business_struct{}
	err := json.Unmarshal([]byte(data), &businessStruct)
	if err != nil {
		return
	}
	business = append(business, businessStruct)
	fmt.Println("工商注册信息", data)
}

func Shareholders(name, data string) {
	data = strings.ReplaceAll(data, "bufsnake shareholders ", "")
	split := strings.Split(data, " ")
	if len(split) == 2 {
		fmt.Println(name, "股东", split[0], split[1])
		shareholders = append(shareholders, shareholders_struct{
			currEnt: name,
			entName: split[0],
			regRate: split[1],
		})
	}
}

func WebRecord(name, data string) {
	data = strings.ReplaceAll(data, "bufsnake webRecord ", "")
	split := strings.Split(data, " ")
	if len(split) == 3 {
		fmt.Println(name, "网站信息", split[0], split[1], split[2])
		webRecord = append(webRecord, webRecord_struct{
			currEnt:  name,
			site:     split[0],
			siteName: split[1],
			icp:      split[2],
		})
	}
}
