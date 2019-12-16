package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Jeffail/gabs"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

// Retrieve most up to date user agent for any browser
const (
	WigleBaseURL          = "https://api.wigle.net/api/v2/network/search/?onlymine=false&freenet=false&paynet=false&ssid=NETWORKNAME"
	WigleAddressLookupURL = "https://api.wigle.net/api/v2/network/geocode?addresscode=TARGETADDRESS"
	WigleWIFIAddressURL   = "https://api.wigle.net/api/v2/network/search/?onlymine=false&freenet=false&paynet=false&latrange1=LATITUDEA&latrange2=LATITUDEB&longrange1=LONGITUDEA&longrange2=LONGITUDEB&resultsPerPage="
	GeoLocateURL          = "https://maps.googleapis.com/maps/api/geocode/json?latlng=LATITUDE,LONGITUDE&sensor=true"
	OwlerSearchURL        = "https://www.owler.com/iaApp/basicSearchCompanySuggestions.htm?searchTerm=QUERY"
	OwlerDetailsURL       = "https://www.owler.com/iaApp/fetchCompanyProfileData.htm"
	OSINTBaseURL          = "https://thatsthem.com/QUERYTYPE/QUERY"
	enrichment            = "https://api.passivetotal.org/v2/enrichment"
	dnsPassive            = "https://api.passivetotal.org/v2/dns/passive"
	whoisURL              = "https://api.passivetotal.org/v2/whois"
	ABuseDBBaseURL        = "https://www.abuseipdb.com/api/v2/check/"
	ShodanBaseURL         = "https://api.shodan.io/shodan/host/search?key=APIKEY&minify=true&query='net:QUERY'"
)

var (
	// UserAgentString - Browser Identity for requests
	UserAgentString = getLatestUserAgent()
	// WigleAPIKey - API Key from wigle.net
	WigleAPIKey = os.Getenv("WIGLEAPIKEY")
	// WigleAPISecret - Secret from wigle.net
	WigleAPISecret = os.Getenv("WIGLEAPISECRET")

	CensysAPIKey = os.Getenv("CENSYSAPIKEY")
	CensysSecret = os.Getenv("CENSYSAPISECRET")

	// userName PassiveTotal account name
	userName = os.Getenv("PTUSER")
	// APIKey PassiveTotal API Key
	APIKey = os.Getenv("PTAPIKEY")
	// Abuse IP DB API key
	AbuseDBKey = os.Getenv("ABUSEDBSECRET")
	// Shodan API key location
	ShodanAPIKey = os.Getenv("SHODANAPIKEY")
	red          = color.New(color.FgRed)
	cyan         = color.New(color.FgCyan)
)

func getLatestUserAgent() string {
	requestURL := "https://www.whatismybrowser.com/guides/the-latest-user-agent/chrome"
	httpResponse, err := http.Get(requestURL)
	parsedHTML, err := goquery.NewDocumentFromReader(httpResponse.Body)
	if err != nil {
		fmt.Println(err)
	}

	latestUserAgent := parsedHTML.Find(".code").First().Text()
	return latestUserAgent
}

func queryShodan(Query string) gabs.Container {

	if ShodanAPIKey != "" {
		httpClient := http.Client{}
		requestURL := strings.Replace(ShodanBaseURL, "APIKEY", ShodanAPIKey, 1)
		requestURL = strings.Replace(requestURL, "QUERY", Query, 1)
		httpRequest, err := http.NewRequest("GET", requestURL, nil)
		httpResponse, err := httpClient.Do(httpRequest)
		responseBytes := httpResponse.Body
		message, err := ioutil.ReadAll(responseBytes)
		prettyPrint, err := gabs.ParseJSON(message)
		openPorts := prettyPrint.Path("matches.*.port").String()
		nodeData := prettyPrint.Path("matches.*.data").String()
		if err != nil {
			log.Fatal("Shodan Error ", string(message), err)
		}
		red.Print("Ports: ")
		fmt.Println(openPorts, "\n")
		red.Print("Banner: ")
		fmt.Println(nodeData, "\n")
		return *prettyPrint
	} else {
		return gabs.Container{}

	}

}

func CheckIPReputation(IPAddress string) {

	httpClient := http.Client{}

	requestURL := ABuseDBBaseURL + "?ipAddress=" + IPAddress
	httpRequest, _ := http.NewRequest("POST", requestURL, nil)
	httpRequest.Header.Add("Key", AbuseDBKey)
	httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	httpRequest.Header.Add("Accept", "application/json")

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.Fatal(err)
	}
	httpResponseBytes, _ := ioutil.ReadAll(httpResponse.Body)
	fmt.Println(string(httpResponseBytes))

}

func queryPTAll(Query string) {

	URLS := []string{dnsPassive, enrichment, whoisURL}
	waitGroup := sync.WaitGroup{}

	// Check account quotas before executing
	// queryAccountQuotas()

	for _, CurrentURL := range URLS {
		waitGroup.Add(1)
		go queryPassiveTotal(CurrentURL, &waitGroup, Query)
	}
	waitGroup.Wait()
}

func queryPassiveTotal(endpoint string, waitgroup *sync.WaitGroup, Query string) {
	httpClient := http.Client{}
	httpRequest, err := http.NewRequest("GET", endpoint+"?query="+Query, nil)
	httpRequest.SetBasicAuth(userName, APIKey)
	httpResponse, err := httpClient.Do(httpRequest)
	responseBytes := httpResponse.Body
	message, err := ioutil.ReadAll(responseBytes)
	prettyPrint, err := gabs.ParseJSON(message)
	if err != nil {
		log.Fatal("Error ", err)
	}
	fmt.Println(string(prettyPrint.String()))
	waitgroup.Done()
}

func queryAccountQuotas() {
	httpClient := http.Client{}
	httpRequest, err := http.NewRequest("GET", "https://api.passivetotal.org/v2/account/quota", nil)
	httpRequest.SetBasicAuth(userName, APIKey)
	httpResponse, err := httpClient.Do(httpRequest)
	responseBytes := httpResponse.Body
	message, err := ioutil.ReadAll(responseBytes)
	prettyPrint, err := gabs.ParseJSON(message)
	if err != nil {
		log.Fatal("Error ", string(message), err)
	}
	fmt.Println(string(prettyPrint.String()))
}

func getPerson(queryType string, query string, state string) []byte {
	// queryType possiblities are name,email,phone,ipaddress, and address
	searchResults := gabs.New()
	requestURL := strings.Replace(OSINTBaseURL, "QUERYTYPE", queryType, 1)
	requestURL = strings.Replace(requestURL, "QUERY", query, 1)
	requestURL = strings.Replace(requestURL, "STATE", state, 1)

	httpResponse, err := http.Get(requestURL)
	parsedHTML, err := goquery.NewDocumentFromReader(httpResponse.Body)
	if err != nil {
		fmt.Println(err)
	}

	searchResults.Array("results")
	parsedHTML.Find(".ThatsThem-record").Each(func(i int, s *goquery.Selection) {
		targetName := s.Find("[itemprop=name]").Text()
		targetStreetAddress := s.Find("[itemprop=streetAddress]").Text()
		targetCity := s.Find("[itemprop=addressLocality]").Text()
		targetState := s.Find("[itemprop=addressRegion]").Text()
		targetPhone := s.Find("[itemprop=telephone]").Text()

		targetResult := gabs.New()
		// cleanup names
		targetName = strings.Split(targetName, "\n")[0]

		targetResult.SetP(targetName, "Name")
		targetResult.SetP(targetStreetAddress, "Address")
		targetResult.SetP(targetCity, "City")
		targetResult.SetP(targetState, "State")
		targetResult.SetP(targetPhone, "Phone")

		searchResults.ArrayAppendP(targetResult.String(), "results")
	})
	return searchResults.Bytes()
}
func getWIFINetworksBySSID(SSID string) []byte {
	// Initialize the client
	httpClient := http.Client{}

	// Replace ESSID
	requestURL := strings.Replace(WigleBaseURL, "NETWORKNAME", SSID, 1)

	// Prepare the request, setting auth aand accept headers
	httpRequest, err := http.NewRequest("GET", requestURL, nil)

	// Authenticate
	httpRequest.SetBasicAuth(WigleAPIKey, WigleAPISecret)
	httpRequest.Header.Set("Accept", "application/json")

	// Send request
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.Fatal("", err)
	}

	// Process response
	bytesCollection, errResponse := ioutil.ReadAll(httpResponse.Body)
	if errResponse != nil {
		log.Fatal("", errResponse)
	}
	return bytesCollection
}

func getWIFINetworksByAddress(Address string) []byte {
	// Initialize the client
	httpClient := http.Client{}

	requestURL := strings.Replace(WigleAddressLookupURL, "TARGETADDRESS", Address, 1)
	// Prepare the request, setting auth aand accept headers
	httpRequest, err := http.NewRequest("GET", requestURL, nil)
	// Authenticate
	httpRequest.SetBasicAuth(WigleAPIKey, WigleAPISecret)
	httpRequest.Header.Set("Accept", "application/json")
	// Send request
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.Fatal("", err)
	}
	// Process response
	bytesCollection, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Fatal(err)
	}

	boundingBox, err := gabs.ParseJSON(bytesCollection)
	geoBox := boundingBox.Path("results.boundingbox").Children()
	if err != nil {
		log.Fatal(err, string(bytesCollection))
	}
	geoBox = geoBox[0].Children()

	requestURL = strings.Replace(WigleWIFIAddressURL, "LATITUDEA", geoBox[0].String(), 1)
	requestURL = strings.Replace(requestURL, "LATITUDEB", geoBox[1].String(), 1)
	requestURL = strings.Replace(requestURL, "LONGITUDEA", geoBox[2].String(), 1)
	requestURL = strings.Replace(requestURL, "LONGITUDEB", geoBox[3].String(), 1)
	// Prepare the request, setting auth aand accept headers
	httpRequest, err = http.NewRequest("GET", requestURL, nil)
	// Authenticate
	httpRequest.SetBasicAuth(WigleAPIKey, WigleAPISecret)
	httpRequest.Header.Set("Accept", "application/json")
	// Send request
	httpResponse, err = httpClient.Do(httpRequest)
	if err != nil {
		log.Fatal("", err)
	}
	// Process response
	bytesCollection, err = ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Fatal("", err)
	}

	return bytesCollection
}
func getOrganizationByName(OrgName string) []byte {
	requestURL := strings.Replace(OwlerSearchURL, "QUERY", OrgName, 1)
	httpResponse, err := http.Get(requestURL)
	if err != nil {
		log.Fatal(err)
	}
	// Process response
	bytesCollection, errResponse := ioutil.ReadAll(httpResponse.Body)
	if errResponse != nil {
		log.Fatal("", errResponse)
	}
	return bytesCollection
}

func getOrganizationDetails(OrgID string) []byte {
	requestURL := OwlerDetailsURL

	httpClient := http.Client{}
	requestBodyJSON := gabs.New()
	requestBodyJSON.SetP("cp", "section")
	requestBodyJSON.SetP(OrgID, "companyId")
	requestBodyJSON.SetP([]string{"company_info", "ceo", "top_competitors", "keystats", "cp"}, "components")

	httpRequest, err := http.NewRequest("POST", requestURL, bytes.NewReader(requestBodyJSON.EncodeJSON()))
	httpRequest.Header.Set("Origin", "https://www.owler.com")
	httpRequest.Header.Set("User-Agent", UserAgentString)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("DNT", "1")
	httpRequest.Header.Set("Accept", "*/*")

	httpResponse, err := httpClient.Do(httpRequest)
	// Process response
	bytesCollection, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Fatal(err)
	}
	return bytesCollection

}

func getSubDomains(domain string) string {
	DNSDUmpsterSearchURL := "http://api.hackertarget.com/hostsearch/?q=QUERY"
	requestURL := strings.Replace(DNSDUmpsterSearchURL, "QUERY", domain, 1)
	httpResponse, err := http.Get(requestURL)
	if err != nil {
		log.Fatal(err)
	}
	// Process response
	bytesCollection, errResponse := ioutil.ReadAll(httpResponse.Body)
	if errResponse != nil {
		log.Fatal("", errResponse)
	}

	subDomains := string(bytesCollection)
	subDomains = strings.Replace(subDomains, ".", " ", 0)
	// fmt.Println(subDomains)
	return subDomains
}

func queryCensys(query string) string {
	httpClient := http.Client{}

	httpRequestBody := `
	{
		"query": "ZZZZZ",
		"page": 1,
		"fields": [
			"80.http.get.title",
			"443.https.get.title",
			"location.registered_country",
			"location.longitude",
			"location.continent",
			"url",
			"ip",
			"location.registered_country_code",
			"location.country_code",
			"location.latitude",
			"protocols"
		]
	}`
	httpRequestBody = strings.Replace(httpRequestBody, "ZZZZZ", query, 1)
	httpRequestData, _ := gabs.ParseJSON([]byte(httpRequestBody))
	requestBodyBytes := httpRequestData.Bytes()
	requestBodyReader := bytes.NewReader(requestBodyBytes)

	httpRequest, err := http.NewRequest("POST", "https://www.censys.io/api/v1/search/ipv4", requestBodyReader)

	httpRequest.SetBasicAuth(CensysAPIKey, CensysSecret)
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Add("User-Agent", UserAgentString)

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.Fatal("Error ", err)
	}

	responseBytes := httpResponse.Body
	message, err := ioutil.ReadAll(responseBytes)
	prettyPrint, err := gabs.ParseJSON(message)
	if err != nil {
		log.Fatal("Error ", string(message), err)
	}
	fmt.Println("Ports:", string(prettyPrint.Path("results.protocols").String()))
	// fmt.Println(string(prettyPrint.Path("results.80.http.get.title").String()))
	return prettyPrint.String()
}

func main() {
	UserQuery := os.Args[1]
	// UserQuery := "Coke"

	Results := getOrganizationByName(UserQuery)

	parsedResults, err := gabs.ParseJSON(Results)
	if err != nil {
		log.Fatal(err)
	}
	companyDataURL := parsedResults.Path("results.*.attributeForAutoSuggestAsMap").Children()
	companyInfo := companyDataURL[0]

	companyID := companyInfo.Path("id").String()
	companyID = strings.Replace(companyID, "\"", "", 2)
	companyDomain := companyInfo.Path("primary_domain").String()
	companyDomain = strings.Replace(companyDomain, "\"", "", 2)

	companyDetails := getOrganizationDetails(companyID)
	parsedResults, err = gabs.ParseJSON(companyDetails)
	if err != nil {
		log.Fatal(err)
	}

	CEOFirstName := parsedResults.Path("ceo.current_ceo.first_name").String()
	CEOLastName := parsedResults.Path("ceo.current_ceo.last_name").String()
	CEOName := strings.Replace(CEOFirstName, "\"", "", 2) + " " + strings.Replace(CEOLastName, "\"", "", 2)
	industrySector := parsedResults.Path("company_info.company_details.industrySector.sector_name")

	companyFounded := parsedResults.Path("company_info.company_details.founded").String()
	companyFounded = strings.Replace(companyFounded, "\"", "", 4)

	companyAddressCountry := parsedResults.Path("company_info.company_details.hqAddress.country").String()
	companyAddressCountry = strings.Replace(companyAddressCountry, "\"", "", 4)

	companyAddressState := parsedResults.Path("company_info.company_details.hqAddress.state").String()
	companyAddressState = strings.Replace(companyAddressState, "\"", "", 4)

	companyAddressCity := parsedResults.Path("company_info.company_details.hqAddress.city").String()
	companyAddressCity = strings.Replace(companyAddressCity, "\"", "", 4)

	companyAddressStreet1 := parsedResults.Path("company_info.company_details.hqAddress.street1").String()
	companyAddressStreet1 = strings.Replace(companyAddressStreet1, "\"", "", 4)

	companyAddressStreet2 := parsedResults.Path("company_info.company_details.hqAddress.street2").String()
	companyAddressStreet2 = strings.Replace(companyAddressStreet2, "\"", "", 4)
	companyFullAddress := (companyAddressStreet1 + " " + companyAddressStreet2 + " " + companyAddressCity + " " + companyAddressState)

	companyName := parsedResults.Path("company_info.company_details.name").String()
	companyName = strings.Replace(companyName, "\"", "", 4)

	cyan.Println("Company Details\n")

	fmt.Println("Name:", companyName)
	fmt.Println("CEO:", CEOName)
	fmt.Println("Founded:", companyFounded)
	fmt.Println("Company TLD:", companyDomain)
	fmt.Println("Industry Sector:", industrySector)
	fmt.Println("Address:", companyFullAddress)

	cyan.Println("\nNetwork Perimeter from Dig\n")

	CompanyDomainEndpoints := getSubDomains(companyDomain)
	EndpointsList := strings.Split(CompanyDomainEndpoints, "\n")
	if len(EndpointsList) < 2 {
		log.Fatal(EndpointsList)
	}

	// EndpointResultsCollection := gabs.New()
	for endpoint := range EndpointsList {
		cyan.Print("IP: ")
		fmt.Println(strings.Split(EndpointsList[endpoint], ",")[1], "\t\t", "Hostname:", strings.Split(EndpointsList[endpoint], ",")[0])
		// queryCensys(strings.Split(EndpointsList[endpoint], ",")[1])
		queryShodan(strings.Split(EndpointsList[endpoint], ",")[1])
		// CEODoxx := getPerson("name", strings.Replace(CEOName, " ", "-", -1), "XX")
		// fmt.Println(string(CEODoxx))
	}

}
