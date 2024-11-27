package src

import (
	"encoding/json"
	"fmt"
	toolkit "github.com/cx-luo/go-toolkit"
	"github.com/gin-gonic/gin"
	"go-pubchem/dao"
	"go-pubchem/pkg"
	"go-pubchem/utils"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	_ "net/url"
	"regexp"
	"strconv"
	"time"
)

// API参考文档 ： https://pubchem.ncbi.nlm.nih.gov/docs/pug-rest#section=The-URL-Path
var sema = toolkit.NewSemaphore(40)

//const prettyElement = CompoundProperty{
//	MolecularFormula:         "MolecularFormula",
//	MolecularWeight:          "MolecularWeight",
//	CanonicalSMILES:          "CanonicalSMILES",
//	IsomericSMILES:           "IsomericSMILES",
//	InChI:                    "InChI",
//	InChIKey:                 "InChIKey",
//	IUPACName:                "IUPACName",
//	Title:                    "Title",
//	XLogP:                    "XLogP",
//	ExactMass:                "ExactMass",
//	MonoisotopicMass:         "MonoisotopicMass",
//	TPSA:                     "TPSA",
//	Complexity:               "Complexity",
//	Charge:                   "Charge",
//	HBondDonorCount:          "HBondDonorCount",
//	HBondAcceptorCount:       "HBondAcceptorCount",
//	RotatableBondCount:       "RotatableBondCount",
//	HeavyAtomCount:           "HeavyAtomCount",
//	IsotopeAtomCount:         "IsotopeAtomCount",
//	AtomStereoCount:          "AtomStereoCount",
//	DefinedAtomStereoCount:   "DefinedAtomStereoCount",
//	UndefinedAtomStereoCount: "UndefinedAtomStereoCount",
//	BondStereoCount:          "BondStereoCount",
//	DefinedBondStereoCount:   "DefinedBondStereoCount",
//	UndefinedBondStereoCount: "UndefinedBondStereoCount",
//	CovalentUnitCount:        "CovalentUnitCount",
//	PatentCount:              "PatentCount",
//	PatentFamilyCount:        "PatentFamilyCount",
//	LiteratureCount:          "LiteratureCount",
//	Volume3D:                 "Volume3D",
//	XStericQuadrupole3D:      "XStericQuadrupole3D",
//	YStericQuadrupole3D:      "YStericQuadrupole3D",
//	ZStericQuadrupole3D:      "ZStericQuadrupole3D",
//	FeatureCount3D:           "FeatureCount3D",
//	FeatureAcceptorCount3D:   "FeatureAcceptorCount3D",
//	FeatureDonorCount3D:      "FeatureDonorCount3D",
//	FeatureAnionCount3D:      "FeatureAnionCount3D",
//	FeatureCationCount3D:     "FeatureCationCount3D",
//	FeatureRingCount3D:       "FeatureRingCount3D",
//	FeatureHydrophobeCount3D: "FeatureHydrophobeCount3D",
//	ConformerModelRMSD3D:     "ConformerModelRMSD3D",
//	EffectiveRotorCount3D:    "EffectiveRotorCount3D",
//	ConformerCount3D:         "ConformerCount3D",
//	Fingerprint2D:            "Fingerprint2D",
//}

type CmpdName struct {
	Name string `json:"name"`
}

type CmpdSmiles struct {
	Smiles string `json:"smiles"`
}

type CmpdCid struct {
	Cid int `json:"cid"`
}

// fetchURL 用来设置代理ip的
func fetchURL(parseurl string) (string, error) {
	// 创建一个自定义的 Transport 实例
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://127.0.0.1:1080") // 设置代理
		},
	}

	// 创建一个自定义的 Client 实例
	client := &http.Client{
		Transport: transport,       // 设置 Transport
		Timeout:   time.Second * 3, // 设置超时
	}

	// 发送 GET 请求
	resp, err := client.Get(parseurl)
	if err != nil {
		// 处理错误
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 处理错误
		return "", err
	}

	// 返回响应内容
	return string(body), nil
}

func urlGet(epUrl string) []byte {
	// 创建一个自定义的 Transport 实例
	transport := &http.Transport{
		//Proxy: func(req *http.Request) (*url.URL, error) {
		//	return url.Parse("http://27.79.147.195:4005") // 设置代理
		//},
		MaxIdleConnsPerHost: 5,  // 每个主机最大空闲连接数
		MaxIdleConns:        20, // 最大空闲连接数
	}

	// 创建一个自定义的 Client 实例
	client := &http.Client{
		Transport: transport,        // 设置 Transport
		Timeout:   time.Second * 10, // 设置超时
	}

	req, err := http.NewRequest("GET", epUrl, nil)
	if err != nil {
		pkg.Logger.Critical(err)
	}

	response, err := client.Do(req)
	if err != nil {
		pkg.Logger.Critical("%v", err)
		return nil
	}

	if response.StatusCode != 200 {
		pkg.Logger.Error("Error status : %s, url : %s ", response.Status, epUrl)
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	bodyText, err := io.ReadAll(response.Body)
	if err != nil {
		pkg.Logger.Critical(err)
		return nil
	}
	return bodyText
}

// GetCidFromSmiles
// @Summary GetCidFromSmiles 从smiles查询cid
// @Description 从smiles查询cid，在返回结果前会把结果写入数据库
// @Tags pug
// @Accept json
// @Param smiles body CmpdSmiles true "smiles"
// @Success 200 {string} string "{"msg": "hello wy"}"
// @Failure 400 {string} string "{"msg": "who are you"}"
// @Router /pug/getCidFromSmiles [post]
func GetCidFromSmiles(c *gin.Context) {
	var s CmpdSmiles
	err := c.ShouldBind(&s)
	if err != nil {
		utils.BadRequestErr(c, err)
		return
	}
	curl := fmt.Sprintf("https://pubchem.ncbi.nlm.nih.gov/rest/pug/compound/smiles/%s/cids/json", s.Smiles)
	res := urlGet(curl)
	type IdentifierList struct {
		IdentifierList struct {
			Cid []int `json:"CID"`
		} `json:"IdentifierList"`
	}
	var idenCids IdentifierList
	if res == nil {
		utils.OkRequestWithData(c, "", gin.H{"total": 0, "list": nil})
		return
	}
	err = json.Unmarshal(res, &idenCids)
	if err != nil {
		utils.InternalRequestErr(c, err)
		return
	}
	if len(idenCids.IdentifierList.Cid) >= 1 && idenCids.IdentifierList.Cid[0] != 0 {
		utils.OkRequestWithData(c, "", gin.H{"total": len(idenCids.IdentifierList.Cid), "list": idenCids.IdentifierList.Cid})
		return
	}
	utils.OkRequest(c, "not found")
	return
}

func GetCIDFromInChi(inChi string) int {
	return 0
}

func GetCIDFromInChiKey(inChiKey string) int {
	return 0
}

// GetCidFromName
// @Summary GetCidFromName 从name查询cid
// @Description insert results to db . If the cid is unique, return the cid, else return nil
// @Tags pug
// @Accept json
// @Param name body CmpdName true "化合物名称"
// @Success 200 {string} string "{"msg": "hello wy"}"
// @Failure 400 {string} string "{"msg": "who are you"}"
// @Router /pug/getCidFromName [post]
func GetCidFromName(c *gin.Context) {
	var s CmpdName
	err := c.ShouldBind(&s)
	if err != nil {
		utils.BadRequestErr(c, err)
		return
	}

	type CIDs struct {
		ConceptsAndCIDs struct {
			CID []int `json:"CID" gorm:"column:CID"`
		} `json:"ConceptsAndCIDs" gorm:"column:ConceptsAndCIDs"`
	}
	curl := "https://pubchem.ncbi.nlm.nih.gov/rest/pug/concepts/name/JSON?"

	params := url.Values{}
	params.Set("name", s.Name)
	curl = curl + params.Encode()

	var cIds CIDs
	res := urlGet(curl)
	if res == nil {
		utils.OkRequestWithData(c, "", gin.H{"total": 0, "list": nil})
		return
	}

	err = json.Unmarshal(res, &cIds)
	if err != nil {
		utils.InternalRequestErr(c, err)
		return
	}

	// 把查询到的结果写到数据库里
	for _, cid := range cIds.ConceptsAndCIDs.CID {
		sqdSet := GetSDQOutputSetFromCid(cid, 10, 1).SDQOutputSet
		err = InsertSDQToDB(&sqdSet)
		if err != nil {
			utils.InternalRequestErr(c, err)
			return
		}
	}

	if len(cIds.ConceptsAndCIDs.CID) == 1 {
		utils.OkRequestWithData(c, "", gin.H{"total": 1, "list": cIds.ConceptsAndCIDs.CID})
		return
	}

	utils.OkRequestWithData(c, "", gin.H{"total": 0, "list": nil})
	return
}

// InsertToDbByCid
// @Summary InsertToDbByCid 把对应Cid的数据写入数据库
// @Description insert compound info to db by cid
// @Tags db
// @Accept json
// @Param cid body CmpdCid true "Cid"
// @Success 200 {string} string "{"msg": "hello wy"}"
// @Failure 400 {string} string "{"msg": "who are you"}"
// @Router /db/insertToDbByCid [post]
func InsertToDbByCid(c *gin.Context) {
	var s CmpdCid
	err := c.ShouldBind(&s)
	if err != nil {
		utils.BadRequestErr(c, err)
		return
	}
	sqdSet := GetSDQOutputSetFromCid(s.Cid, 1, 1).SDQOutputSet
	err = InsertSDQToDB(&sqdSet)
	if err != nil {
		utils.InternalRequestErr(c, err)
		return
	}

	utils.OkRequest(c, "Success")
	return
}

func getCasByRegexp(s string) []string {
	casRegex := regexp.MustCompile(`\b\d{2,7}-\d{2}-\d\b`)
	casNumbers := casRegex.FindAllString(s, -1)
	var validCasNumbers []string
	for _, casNumber := range casNumbers {
		if calculateChecksum(casNumber) {
			validCasNumbers = append(validCasNumbers, casNumber)
		}
	}
	return validCasNumbers
}

type usedProps struct {
	Cid              int     `json:"cid"`
	Cmpdname         string  `json:"cmpdname"`
	Mf               string  `json:"mf"`
	Mw               float64 `json:"mw"`
	Isosmiles        string  `json:"isosmiles"`
	Exactmass        float64 `json:"exactmass"`
	Monoisotopicmass float64 `json:"monoisotopicmass"`
	Inchi            string  `json:"inchi"`
	Inchikey         string  `json:"inchikey"`
	Iupacname        string  `json:"iupacname"`
	Canonicalsmiles  string  `json:"canonicalsmiles"`
}

type usedRows struct {
	Compound usedProps `json:"compound"`
	Cas      []string  `json:"cas"`
}
type usedCmpd struct {
	TotalCount int        `json:"totalCount"`
	Rows       []usedRows `json:"rows"`
}

// GetCmpdWithCasFromCid
// @Summary GetCmpdWithCasFromCid 从cid获取化合物信息
// @Description 从cid获取化合物信息，返回列表
// @Tags query
// @Accept json
// @Param cid body CmpdCid true "Cid"
// @Success 200 {string} string "{"total": 0, "list": []}"
// @Failure 400 {string} string "{"msg": "who are you"}"
// @Router /query/getCmpdWithCasFromCid [post]
func GetCmpdWithCasFromCid(c *gin.Context) {
	var s CmpdCid
	err := c.ShouldBind(&s)
	if err != nil {
		utils.BadRequestErr(c, err)
		return
	}
	var compounds []usedRows
	sqdSet := GetSDQOutputSetFromCid(s.Cid, 10, 1).SDQOutputSet
	for _, row := range sqdSet[0].Rows {
		cas := getCasByRegexp(row.Cmpdsynonym)
		var u = usedRows{
			Compound: usedProps{
				Cid:              row.Cid,
				Mf:               row.Mf,
				Mw:               row.Mw,
				Exactmass:        row.Exactmass,
				Monoisotopicmass: row.Monoisotopicmass,
				Cmpdname:         row.Cmpdname,
				Inchi:            row.Inchi,
				Inchikey:         row.Inchikey,
				Isosmiles:        row.Isosmiles,
				Iupacname:        row.Iupacname,
				Canonicalsmiles:  row.Canonicalsmiles,
			},
			Cas: cas,
		}
		compounds = append(compounds, u)
	}
	utils.OkRequestWithData(c, "", gin.H{"total": len(compounds), "list": compounds})
	return
}

/*
// 还有另外一种方法，使用cache来查询，通过获取cachekey，可以得到更为方便处理的结果，json结构为 Compound
*/

/*
GetCacheKeyAndHitCountFromFormula

GetCacheKeyFromFormula
*/
func GetCacheKeyAndHitCountFromFormula(molecularFormula string, queryType string) (string, int) {
	currUrl := "https://pubchem.ncbi.nlm.nih.gov/unified_search/structure_search.cgi?"
	var queryBlob QueryBlob
	queryBlob.Query.Type = queryType
	var parameters *Parameter
	//params.Set("queryblob", `{"query":{"type":"formula","parameter":[{"name":"FormulaQuery","string":"C9H8O4"},{"name":"UseCache","bool":true},{"name":"SearchTimeMsec","num":5000},{"name":"SearchMaxRecords","num":100000},{"name":"allowotherelements","bool":false}]}}`)
	parameters = new(Parameter)
	parameters.Name = "FormulaQuery"
	parameters.String = molecularFormula
	queryBlob.Query.Parameter = append(queryBlob.Query.Parameter, *parameters)

	parameters = new(Parameter)
	parameters.Name = "UseCache"
	parameters.Bool = true
	queryBlob.Query.Parameter = append(queryBlob.Query.Parameter, *parameters)

	params := url.Values{}
	params.Set("format", "json")
	params.Set("queryblob", queryBlob.toString())
	currUrl = currUrl + params.Encode()
	var pubchemCache PubchemCache

	bodyText := urlGet(currUrl)
	err := json.Unmarshal(bodyText, &pubchemCache)
	if err != nil {
		pkg.Logger.Error(err)
	}
	if pubchemCache.Response.Status != 0 {
		pkg.Logger.Error(pubchemCache.Response.Message)
	}
	return pubchemCache.Response.Cachekey, pubchemCache.Response.Hitcount
}

func GetCacheKeyAndHitCountFromSmiles(smiles string) string {
	currentUrl := "https://pubchem.ncbi.nlm.nih.gov/unified_search/structure_search.cgi?"
	var queryBlob QueryBlob
	queryBlob.Query.Type = "identity"
	var parameters *Parameter
	//{"query":{"type":"identity","parameter":[{"name":"SMILES","string":"COCCO"},{"name":"UseCache","bool":true},{"name":"identity_type","string":"same_stereo_isotope"}]}}
	parameters = new(Parameter)
	parameters.Name = "SMILES"
	parameters.String = smiles
	queryBlob.Query.Parameter = append(queryBlob.Query.Parameter, *parameters)
	parameters = new(Parameter)
	parameters.Name = "UseCache"
	parameters.Bool = true
	queryBlob.Query.Parameter = append(queryBlob.Query.Parameter, *parameters)
	parameters = new(Parameter)
	parameters.Name = "identity_type"
	parameters.String = "same_stereo_isotope"
	queryBlob.Query.Parameter = append(queryBlob.Query.Parameter, *parameters)
	params := url.Values{}
	params.Set("format", "json")
	params.Set("queryblob", queryBlob.toString())
	currentUrl = currentUrl + params.Encode()

	var pubChemCache PubchemCache

	bodyText := urlGet(currentUrl)
	err := json.Unmarshal(bodyText, &pubChemCache)
	if err != nil {
		pkg.Logger.Error(err)
	}
	if pubChemCache.Response.Status != 0 {
		pkg.Logger.Error(pubChemCache.Response.Message)
	}
	fmt.Println(pubChemCache.Response)
	return pubChemCache.Response.Cachekey
}

func GetSDQOutputSetFromCacheKey(netCacheKey string, limit int, start int, orderType string) SDQOutputSet {
	jsData := fmt.Sprintf(`{
		"select":"*",
		"collection":"compound",
		"where":{
		"ands":[
	{
	"input":{
	"type":"netcachekey",
	"idtype":"cid",
	"key":"%s"
	}
	}
	]
	},
	"order":[
	"%s"
	],
	"start":%d,
	"limit":%d,
	"width":1000000,
	"listids":0
	}`, netCacheKey, orderType, start, limit)
	var netCacheKeyPayload NetCacheKeyPayload
	err := json.Unmarshal([]byte(jsData), &netCacheKeyPayload)
	if err != nil {
		pkg.Logger.Error(err)
	}
	currUrl := "https://pubchem.ncbi.nlm.nih.gov/sdq/sdqagent.cgi?"
	params := url.Values{}
	params.Set("infmt", "json")
	params.Set("outfmt", "json")
	params.Set("query", jsData)
	currUrl = currUrl + params.Encode()

	var sdq SDQOutputSet
	bodyText := urlGet(currUrl)
	err = json.Unmarshal(bodyText, &sdq)
	return sdq
}

/*
获取SDQ形式的结果
*/

/*
GetSDQOutputSetFromQuery

"order":["relevancescore,desc"]

In order of relevance, the first one is closer to the result.

按照相关性排序，排在第一位的，更接近结果.
*/
func GetSDQOutputSetFromQuery(cName string, limit int, start int) SDQOutputSet {
	jsData := fmt.Sprintf(`{"select":"*","collection":"compound","where":{"ands":[{"*":"%s"}]},"order":["relevancescore,desc"],"start":%d,"limit":%d,"width":1000000,"listids":0}
`, cName, start, limit)
	currUrl := "https://pubchem.ncbi.nlm.nih.gov/sdq/sdqagent.cgi?"
	params := url.Values{}
	params.Set("infmt", "json")
	params.Set("outfmt", "json")
	params.Set("query", jsData)
	currUrl = currUrl + params.Encode()

	var sdq SDQOutputSet
	bodyText := urlGet(currUrl)
	if bodyText == nil {
		return SDQOutputSet{SDQOutputSet: nil}
	}
	err := json.Unmarshal(bodyText, &sdq)
	if err != nil {
		pkg.Logger.Error(err)
		return SDQOutputSet{SDQOutputSet: nil}
	}
	return sdq
}

/*
GetSDQOutputSetFromCid 通过cid获取SDQOutputSet
*/
func GetSDQOutputSetFromCid(cid int, limit int, start int) SDQOutputSet {
	jsData := fmt.Sprintf(`{"select":"*","collection":"compound","where":{"ands":[{"cid":"%d"}]},"order":["cid,asc"],"start":%d,"limit":%d,"width":1000000,"listids":0}`, cid, start, limit)
	currUrl := "https://pubchem.ncbi.nlm.nih.gov/sdq/sdqagent.cgi?"
	params := url.Values{}
	params.Set("infmt", "json")
	params.Set("outfmt", "json")
	params.Set("query", jsData)
	currUrl = currUrl + params.Encode()
	pkg.Logger.Info(currUrl)

	var sdq SDQOutputSet
	bodyText := urlGet(currUrl)
	err := json.Unmarshal(bodyText, &sdq)
	if err != nil {
		pkg.Logger.Error(err)
		return SDQOutputSet{SDQOutputSet: nil}
	}
	if sdq.SDQOutputSet[0].Status.Code != 0 {
		pkg.Logger.Error("GetSDQOutputSetFromCid : %d, %d, %s", cid, sdq.SDQOutputSet[0].Status.Code, sdq.SDQOutputSet[0].Status.Error)
		return SDQOutputSet{SDQOutputSet: nil}
	}
	return sdq
}

/*
	InsertCompoundsToDB insert the conmpound info to table.

-- compound_from_pubchem definition

CREATE TABLE `compound_from_pubchem` (

	`cid` int NOT NULL,
	`mw` float DEFAULT NULL,
	`polararea` int DEFAULT NULL,
	`complexity` float DEFAULT NULL,
	`xlogp` float DEFAULT NULL,
	`exactmass` float DEFAULT NULL,
	`monoisotopicmass` float DEFAULT NULL,
	`heavycnt` int DEFAULT NULL,
	`hbonddonor` int DEFAULT NULL,
	`hbondacc` int DEFAULT NULL,
	`rotbonds` int DEFAULT NULL,
	`annothitcnt` int DEFAULT NULL,
	`charge` int DEFAULT NULL,
	`covalentunitcnt` int DEFAULT NULL,
	`isotopeatomcnt` int DEFAULT NULL,
	`totalatomstereocnt` int DEFAULT NULL,
	`definedatomstereocnt` int DEFAULT NULL,
	`undefinedatomstereocnt` int DEFAULT NULL,
	`totalbondstereocnt` int DEFAULT NULL,
	`definedbondstereocnt` int DEFAULT NULL,
	`undefinedbondstereocnt` int DEFAULT NULL,
	`pclidcnt` int DEFAULT NULL,
	`gpidcnt` int DEFAULT NULL,
	`gpfamilycnt` int DEFAULT NULL,
	`aids` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`cmpdname` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`cmpdsynonym` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`inchi` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`inchikey` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`isosmiles` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`iupacname` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`mf` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`sidsrcname` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
	`cidcdate` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`depcatg` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`annothits` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`neighbortype` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	`canonicalsmiles` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
	PRIMARY KEY (`cid`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
*/
func (s *SDQSet) InsertCompoundsToDB() error {
	insertSql := `replace INTO ai_repo.compound_from_pubchem(
		cid, mw, polararea, complexity, xlogp, exactmass,
		monoisotopicmass, heavycnt, hbonddonor, hbondacc, 
		rotbonds, annothitcnt, charge, covalentunitcnt, 
		isotopeatomcnt, totalatomstereocnt, definedatomstereocnt, 
		undefinedatomstereocnt,
		totalbondstereocnt,
		definedbondstereocnt,
		undefinedbondstereocnt,
		pclidcnt,
		gpidcnt,
		gpfamilycnt,
		aids,
		cmpdname,
        cmpdsynonym,
		inchi,
		inchikey,
		isosmiles,
		iupacname,
		mf,
		sidsrcname,
		cidcdate,
		depcatg,
		annothits,
		neighbortype,
		canonicalsmiles)
		VALUES (:cid,
		:mw,
		:polararea,
		:complexity,
		:xlogp,
		:exactmass,
		:monoisotopicmass,
		:heavycnt,
		:hbonddonor,
		:hbondacc,
		:rotbonds,
		:annothitcnt,
		:charge,
		:covalentunitcnt,
		:isotopeatomcnt,
		:totalatomstereocnt,
		:definedatomstereocnt,
		:undefinedatomstereocnt,
		:totalbondstereocnt,
		:definedbondstereocnt,
		:undefinedbondstereocnt,
		:pclidcnt,
		:gpidcnt,
		:gpfamilycnt,
		:aids,
		:cmpdname,
		:cmpdsynonym,
		:inchi,
		:inchikey,
		:isosmiles,
		:iupacname,
		:mf,
		:sidsrcname,
		:cidcdate,
		:depcatg,
		:annothits,
		:neighbortype,
		:canonicalsmiles)`

	for i := 0; i < len(s.Rows); i++ {
		row := s.Rows[i]
		sema.Acquire(1)
		go func() {
			defer sema.Release()
			err := func(c Compound) error {
				_, err := dao.MysqlCursor.NamedExec(insertSql, c)
				if err != nil {
					return err
				}

				return nil
			}(row)
			if err != nil {
				pkg.Logger.Error(err)
			}
		}()
	}
	sema.Wait()
	return nil
}

func InsertSDQToDB(s *[]SDQSet) error {
	for i := 0; i < len(*s); i++ {
		err := (*s)[i].InsertCompoundsToDB()
		if err != nil {
			return err
		}
	}
	return nil
}

func updateTableBySql(cid int, cName string, sqlStr string) error {
	pkg.Logger.Info("Begin to update : %s, %d", cName, cid)
	//updateSql := `update enotess.moc_condition_molecule_std set cid = ?
	//                                      where standardized_name = ?`
	affect, err := dao.MysqlCursor.Exec(sqlStr, cid, cName)
	if err != nil {
		return err
	}
	rowsAffected, _ := affect.RowsAffected()
	pkg.Logger.Info("Affected rows : %d, %s, %d, %s", rowsAffected, sqlStr, cid, cName)
	return nil
}

//func tagProcessed(cName string) error {
//	updateSql := `update test.all_condition_first set processed = 1 where name = ? or standardized_name = ?`
//	_, err := dao.MysqlCursor.Exec(updateSql, cName, cName)
//	if err != nil {
//		return err
//	}
//	//rowsAffected, _ := affect.RowsAffected()
//	//dao.Logger.Info("Affected rows : %d, %s,  %s", rowsAffected, updateSql, cName)
//	return nil
//}

/*
	GetCompoundInfo

1. 先通过名字获取Cid，如果能获取到，则通过cid去查询compound.
2. 如果查询不到，则通过名称去查
3. TODO 如果也查询不到，换一种方式
*/

// GetCmpdFromQueryLimit
// @Summary GetCmpdFromQueryLimit 从query获取化合物信息
// @Description 获取不那么准的信息，并写入表里，返回前10个查询结果
// @Tags query
// @Accept json
// @Param name body CmpdName true "化合物名称"
// @Success 200 {string} string "{"statusCode":200,"msg":"","data":{"list":[168478138],"total":1}}"
// @Failure 400 {string} string "{"statusCode":400,"msg":"error!"}"
// @Router /query/getCmpdFromQueryLimit [post]
func GetCmpdFromQueryLimit(c *gin.Context) {
	var s CmpdName
	err := c.ShouldBind(&s)
	if err != nil {
		utils.BadRequestErr(c, err)
		return
	}
	//err := tagProcessed(cName)
	//if err != nil {
	//	return err
	//}
	//nameResolved := strings.SplitN(cName, "|", 2)
	//fmt.Println(nameResolved)
	//cid := GetCidFromName(c) // 在这一步判断cid是否唯一，不唯一返回的是0
	//if cid != 0 {
	//	fmt.Println(cName, cid)
	// 通常情况下，一个cid只对应一个化合物，所以只获取第一个.
	// 不过为了严谨，还是判断一下长度，循环处理

	//return nil
	//}

	// 如果cid为0，说明可能没查出来。换一种方式，可以通过GetSDQOutputSetFromQuery去查询
	// 先获取一千条，一般来说，不会大于1000条，pubchem允许获取单次最大10000.
	sqdSet := GetSDQOutputSetFromQuery(s.Name, 1000, 1).SDQOutputSet
	if sqdSet == nil {
		utils.OkRequestWithData(c, "", gin.H{"total": 0, "list": nil})
		return
	}
	totalCount := sqdSet[0].TotalCount
	switch {
	case totalCount == 0:
		pkg.Logger.Info("Don't find,compound name : %s.", s.Name)
		utils.OkRequestWithData(c, "", gin.H{"total": 0, "list": nil})
		return

	case totalCount > 1000:
		// 先把当前的一千条写入
		err := InsertSDQToDB(&sqdSet)
		if err != nil {
			utils.InternalRequestErr(c, err)
			return
		}

		utils.OkRequestWithData(c, "", gin.H{"total": totalCount, "list": sqdSet[:10]})
		// 获取一千条之后的，再次写入
		cnt := math.Ceil(float64(totalCount) / 1000)
		for i := 1; i < int(cnt); i++ {
			sqdSet := GetSDQOutputSetFromQuery(s.Name, 1000, i*1000+1).SDQOutputSet
			err := InsertSDQToDB(&sqdSet)
			if err != nil {
				utils.InternalRequestErr(c, err)
				return
			}
		}

	default:
		err := InsertSDQToDB(&sqdSet)
		if err != nil {
			utils.InternalRequestErr(c, err)
			return
		}
		utils.OkRequestWithData(c, "", gin.H{"total": totalCount, "list": sqdSet[0].Rows[:10]})
		return
	}
	return
}

func randomInt(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())

	if min >= max {
		return min
	}

	rangeFloat := float64(max-min) + 1
	randomFloat := rand.Float64() * rangeFloat

	return int64(randomFloat) + min
}

func InsertCompoundInfo() {
	var cids []int
	//var cids = []int{
	//	24096961, 3036746, 6604605, 443311, 23618032, 439530, 67914943, 163285965, 6035, 236184, 6137, 876, 84815, 6101, 521998, 2771871, 2723706, 66601, 167158, 118701110, 135408751, 7200, 135465081, 6101, 521998, 5324279, 7863, 5324280, 1030, 259994, 439846, 82916, 8135, 10937607, 11026, 110612, 25199637, 25200171, 11026, 110612, 25199637, 25200171, 11026, 110612, 25199637, 25200171, 11026, 110612, 25199637, 25200171, 11026, 110612, 25199637, 25200171, 11026, 110612, 25199637, 25200171, 9811564, 6505921, 9897285, 92796, 13955575, 2735131, 24190657, 25203940, 64774, 13989026, 15975413, 131875416, 11383651, 134715655, 156026226, 18462, 5771688, 131278, 5771688, 218580, 5284468, 124040768, 15978235, 53384597, 76030824, 21875429, 131674917, 152122, 11123095, 21674, 122130469, 423677, 86657289, 24193707, 152122, 11123095, 21674, 122130469, 423677, 86657289, 24193707, 152122, 11123095, 21674, 122130469, 423677, 86657289, 24193707, 152122, 11123095, 21674, 122130469, 423677, 86657289, 24193707, 74236, 169536, 13598032, 84284, 10874691, 10352, 638051, 28112, 11062293, 617, 5951, 5806, 443400, 6034084, 97268, 90454, 2757, 2724372, 6957610, 6916142, 688243, 14822549, 5702153, 6419964, 101744, 16394527, 449293, 11859618, 16218784, 5462814, 31269, 123025, 2735065, 8759, 13020083, 9902403, 6049, 9839867, 44120005, 18401221, 57339238, 2181, 205586, 641266, 10975445, 79028, 10855416, 12313473, 133671389, 157653234, 9846230, 11025, 18668094, 50931538, 176, 9846230, 11025, 18668094, 50931538, 176, 11143, 45051571, 3776, 11143, 45051571, 3776, 7067560, 11456593, 219661, 2723618, 101782, 11639, 521874, 654, 5166299, 779, 12647, 2724170, 135460230, 12764, 135465065, 23666342, 875, 9836981, 21902459, 517096, 54607388, 2733415, 2733288, 70429, 2724356, 102560, 5885, 5886, 15938972, 16219760, 225710, 88180, 24417, 737158, 99777, 517326, 6597, 25975, 5284468, 124040768, 15978235, 53384597, 76030824, 21875429, 131674917, 92043608, 138991687, 6433264, 3032389, 91884664, 62034, 16759156, 637516, 7409, 108092, 175534, 139057631, 5376733, 15590, 10868176, 134689737, 2723716, 6419961, 11957673, 5363146, 75555, 56776520, 6560260, 98232, 260378, 642608, 98052286, 637794, 13223, 75977, 98507, 2723766, 688303, 102891, 18171, 5284342, 13710713, 18954, 6435836, 19966535, 76524, 2724286, 25203935, 76524, 2724286, 25203935, 11821519, 16217092, 92134854, 90489750, 42602778, 50911993, 4128835, 5486770, 22836554, 24939731, 21909834, 122403013, 136863048, 9834912, 10879725, 5112550, 44630395, 76030819, 50919599, 53384569, 5359853, 25199796, 516935, 45051745, 516935, 45051745, 12135446, 2724268, 86601334, 446094, 19001, 439196, 446094, 19001, 439196, 6400537, 6399078, 6399475, 5360053, 21471, 5374066, 5459377, 26197, 521701, 21801, 25200196, 25203927, 23689980, 23666330, 24181102, 22323430, 22465, 2723721, 637922, 20024, 643789, 9357, 9855836, 44120005, 636371, 71312661, 13020083, 6049, 9839867, 16211050, 21977868, 57339259, 80811, 517696, 25199807, 131675103, 70681, 16212273, 2724287, 6322, 5287702, 28782, 66250, 444360, 1549073, 643756, 95308, 637551, 23670855, 4463282, 24201352, 18374552, 10484, 23670855, 4463282, 24201352, 18374552, 10484, 25248, 6913588, 25248, 6913588, 11010647, 71310736, 145705882, 6436379, 82920, 6436379, 82920, 25798, 25799, 6433572, 2723615, 11149677, 25949, 5749883, 44135491, 49868025, 53384494, 25199798, 134688993, 5483663, 11110968, 84159, 54669729, 53384555, 4651115, 134688994, 5483663, 11110968, 84159, 54669729, 53384555, 4651115, 134688994, 5483663, 11110968, 84159, 54669729, 53384555, 4651115, 134688994, 5483664, 10967121, 16212262, 3645904, 5377486, 22836398, 25199829, 53384493, 134688992, 156592367, 643918, 84253, 642240, 498315, 84306, 6332751, 9576005, 6816376, 6399471, 84538, 10865532, 6392637, 156588341, 54587876, 156619680, 162394253, 84538, 10865532, 6392637, 156588341, 54587876, 156619680, 162394253, 9871585, 2734542, 22831854, 74763990, 16213138, 44658185, 1550213, 86498, 9566465, 2731007, 6074309, 5487681, 30926, 5360664, 9581576, 609329, 5375730, 90200, 16211298, 21116599, 94358, 5366669, 9601896, 94358, 5366669, 9601896, 9566069, 2724994, 11032497, 2724995, 9566069, 2724994, 11032497, 2724995, 12040442, 11309787, 10887092, 22250742, 76030739, 71463884, 132285030, 73877899, 154825255, 155886776, 11984827, 53384301, 78076231, 169536, 9881569, 2794833, 9582835, 5712089, 2794833, 9582835, 5712089, 53384630, 21879930, 74765431, 154825015, 24942131, 131718033, 11636836, 71751692, 2724201, 165340051, 16211607, 10908382, 11000435, 11514661, 91884784, 11000435, 11514661, 91884784, 9582826, 2794666, 5712079, 9811564, 6505921, 9897285, 5365066, 5868400, 93865, 101051, 5366448, 6436851, 135445743, 136259513, 854020, 6916014, 145454, 658384, 2724986, 6542022, 6097878, 6927074, 67752744, 11329479, 86278429, 15606283, 126842319, 50909485, 2725003, 3267515, 139211162, 15659126, 49758013, 57558961, 118704813, 145926263, 156619417, 11329479, 126842319, 145926263, 15606283, 2725003, 7128360, 572659, 16212266, 86278371, 11075194, 13619565, 7128359, 13619563, 13619568, 122410204, 16212283, 572659, 10977265, 98044902, 86278372, 11075194, 18531129, 44630207, 132988852, 13619565, 122410204, 132935775, 9838490, 551090, 9923113, 13515837, 4076588, 22831428, 44630215, 91873398, 13515831, 133124869, 134693906, 2733286, 9601231, 6508711, 11355477, 3382465, 11143, 45051571, 3776, 11143, 45051571, 3776, 11026, 110612, 25199637, 25200171, 11542188, 21976333, 53297465, 101043619, 5702657, 73042, 51003674, 10963812, 73554456, 6415376, 121514148, 124040763, 16212514, 92131501, 155897334, 90908, 640091, 11026, 110612, 25199637, 25200171, 13347, 5142, 167845, 4144353, 50912043, 11198150, 23629008, 644020, 3966, 7014, 168213, 15939859, 442965, 12304487, 11744782, 12538348, 70700858, 77896185, 133636837, 444041, 320761, 24238, 71306822, 90384115, 6518168, 101136808, 124079507, 6102075, 84124, 420332, 6332635, 51003665, 131664180, 21894687, 4418607, 54716542, 24720980, 137319740, 16211827, 54691831, 498840, 10271322, 425002, 53380994, 76029261, 12482041, 86278388, 11011589, 10989779, 136663, 707065, 707067, 10912181, 11219097, 2723635, 2830832, 9859210, 45933886, 6857375, 24139, 11861101, 3368226, 10954510, 5462715, 3035388, 5486778, 44630150, 50897067, 53384484, 517722, 50930385, 10942126, 131858450, 427911, 10909282, 24196496, 129652858, 11066282, 14059101, 73906282, 146025878, 11388194, 45358635, 54725666, 11026, 110612, 25199637, 25200171, 52945042, 5884, 2733511, 15983949, 91654154, 91886371, 145807040, 11388194, 59117881, 10908382, 84353, 44140593, 53216975, 74788062, 76001521, 101136483, 124040745, 145712247, 14924461, 118796993, 121514139, 5892, 5893, 15938971, 10897651, 73415790, 18462, 65617, 218580, 131278, 6433572, 2723615, 11149677, 25949, 5749883, 44135491, 49868025, 53384494, 25199798, 134688993, 5702662, 2734867, 74787731, 551090, 2734713, 72750008, 16218143, 21782969, 131885550, 11204981, 129827494, 145712438, 152122, 11123095, 21674, 122130469, 423677, 86657289, 24193707, 152122, 11123095, 21674, 122130469, 423677, 86657289, 24193707, 11272965, 16212311, 637090, 19499, 5369162, 121015, 5369175, 14229, 5324720, 70699, 5325263, 99777, 737158, 23272, 44134826, 643756, 95308, 637551, 5487681, 30926, 5360664, 15104, 207306, 2733309, 9580359, 11863663, 7057953, 150502, 180596, 89553, 149229, 188583, 89553, 639675, 94297, 117065403, 162344841, 166594976, 166642308, 155882747, 162342433, 10974168, 15971881, 91972075, 124219572, 134715651, 162343995, 11017574, 15972311, 25000063, 163196553, 73553235, 84678, 71463915, 138991738, 72376320, 131675868,
	//}
	err := dao.MysqlCursor.Select(&cids, `select a.cid from enotess.buchwald_reactants_all a left join ai_repo.compound_from_pubchem b on a.cid  = b.cid where a.cid != 0 and b.cid is null group by cid order by cid `)
	//err := dao.MysqlCursor.Select(&stNames, `SELECT name from test.all_condition_first acf WHERE processed = 0;`)
	if err != nil {
		panic(err)
	}
	for _, cid := range cids {
		sema.Acquire(1)
		go func(n int) {
			defer sema.Release()
			sqdSet := GetSDQOutputSetFromCid(n, 10, 1).SDQOutputSet
			if len(sqdSet) == 0 {
				return
			}
			pkg.Logger.Info("%v\n", sqdSet)
			err := InsertSDQToDB(&sqdSet)
			if err != nil {
				pkg.Logger.Error(err)
			}
			//rInt := randomInt(1000, 2000)
			//time.Sleep(time.Millisecond * time.Duration(rInt))
			return
		}(cid)
	}
	sema.Wait()
	time.Sleep(2 * time.Second)
}

func calculateChecksum(cas string) bool {
	casRegex := regexp.MustCompile(`(\d{2,7})-(\d{2})-(\d)`)
	match := casRegex.FindStringSubmatch(cas)
	// 将part1和part2拼接起来
	number := []rune(match[1] + match[2])

	// 初始化校验码计算的和
	sum := 0

	// 从最低位开始计算，即从字符串的末尾开始
	for i, r := range number {
		digit, err := strconv.Atoi(string(r))
		if err != nil {
			pkg.Logger.Error(err)
			return false
		}
		sum += (len(number) - i) * digit
	}

	// 计算最终的校验码
	checksum := sum % 10
	atoi, err := strconv.Atoi(match[3])
	if err != nil {
		pkg.Logger.Error(err)
		return false
	}

	return checksum == atoi
}

// PubChemURLBuilder 结构用于构建PubChem API的URL
type PubChemURLBuilder struct {
	Prefix     string
	InputSpec  string
	Operation  string
	OutputSpec string
	Options    url.Values
}

// NewPubChemURLBuilder 初始化并返回一个新的PubChemURLBuilder实例
func NewPubChemURLBuilder() *PubChemURLBuilder {
	return &PubChemURLBuilder{
		Prefix:  "https://pubchem.ncbi.nlm.nih.gov/rest/pug/",
		Options: url.Values{},
	}
}

// SetInputSpec 设置输入规格
func (b *PubChemURLBuilder) SetInputSpec(domain, namespace, identifiers string) {
	b.InputSpec = fmt.Sprintf("%s/%s/%s", domain, namespace, identifiers)
}

// SetOperation 设置操作规格
func (b *PubChemURLBuilder) SetOperation(operation string) {
	b.Operation = operation
}

// SetOutputSpec 设置输出规格
func (b *PubChemURLBuilder) SetOutputSpec(output string) {
	b.OutputSpec = output
}

// AddOption 添加操作选项
func (b *PubChemURLBuilder) AddOption(key, value string) {
	b.Options.Add(key, value)
}

// BuildURL 构建并返回完整的URL
func (b *PubChemURLBuilder) BuildURL() string {
	urlPath := fmt.Sprintf("%s/%s/%s/%s", b.Prefix, b.InputSpec, b.Operation, b.OutputSpec)
	if b.Options.Encode() != "" {
		urlPath += "?" + b.Options.Encode()
	}
	return urlPath
}

func BuildUrl(c *gin.Context) {
	// 创建一个新的URL构建器实例
	builder := NewPubChemURLBuilder()

	// 设置输入规格，例如compound domain的CID
	builder.SetInputSpec("compound", "cid", "1234,5678")

	// 设置操作规格，例如获取化合物信息
	builder.SetOperation("view")

	// 设置输出规格，例如JSON格式
	builder.SetOutputSpec("json")

	// 添加一些操作选项
	builder.AddOption("response_type", "json")
	builder.AddOption("relaxed_query", "true")

	// 构建并打印最终的URL
	finalURL := builder.BuildURL()
	fmt.Println(finalURL)
}
