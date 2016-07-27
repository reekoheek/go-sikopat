package api

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	NotLoginErr    = errors.New("Not login yet")
	LoginFailedErr = errors.New("Login failed")
)

type (
	Api struct {
		baseUrl      *url.URL
		data         *Data
		commonClient *http.Client
		clients      map[string]*http.Client
	}
)

func (api *Api) get(client *http.Client, uri string) (*goquery.Document, error) {
	//log.Println("get", uri)
	res, err := client.Get(uri)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (api *Api) createClient() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Error on create new cookie jar")
	}

	return &http.Client{
		Jar: jar,
	}
}

func (api *Api) getClient(token string) *http.Client {
	if token != "" && api.clients[token] != nil {
		return api.clients[token]
	}

	if api.commonClient == nil {
		api.commonClient = api.createClient()
	}

	return api.commonClient
}

func (api *Api) isUserClient(client *http.Client) bool {
	return len(client.Jar.Cookies(api.baseUrl)) != 0
}

func (api *Api) Logout(token string) error {
	client := api.getClient(token)
	if api.isUserClient(client) {
		api.get(client, api.baseUrl.String()+"/logout")
	}

	delete(api.clients, token)
	api.data.UnsetProfile(token)

	return nil
}

func (api *Api) Login(username string, password string) (string, error) {
	client := api.createClient()

	form := url.Values{
		"username": {username},
		"password": {password},
	}

	res, err := client.PostForm(api.baseUrl.String()+"/login", form)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", LoginFailedErr
	}

	cookie := client.Jar.Cookies(api.baseUrl)
	token := cookie[0].Value
	api.clients[token] = client

	profile := &Profile{
		Username:    username,
		Password:    password,
		Token:       token,
		RemoteToken: token,
	}
	api.data.SetProfile(token, profile)

	_, err = api.Profile(token, true)
	return token, err
}

func (api *Api) Profile(token string, force bool) (*Profile, error) {
	profile := api.data.Profile(token)
	if profile == nil {
		return nil, NotLoginErr
	} else if !force {
		return profile, nil
	}

	client := api.getClient(token)
	if !api.isUserClient(client) {
		return nil, NotLoginErr
	}

	doc, err := api.get(client, api.baseUrl.String()+"/sales")
	if err != nil {
		return profile, err
	}

	debt := doc.Find(".hutang").Text()
	result := regexp.MustCompile("Hutang Anda : Rp.\\s+(\\d+)").FindAllStringSubmatch(debt, -1)
	balance, _ := strconv.Atoi(result[0][1])
	profile.Balance = balance

	sales := make([]*Sale, 0)
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		qty, _ := strconv.Atoi(strings.Trim(s.Find("td").Eq(3).Text(), "\r\n\t "))
		price, _ := strconv.Atoi(strings.Trim(s.Find("td").Eq(4).Text(), "\r\n\t "))
		total, _ := strconv.Atoi(strings.Trim(s.Find("td").Eq(5).Text(), "\r\n\t "))
		sale := &Sale{
			Product: strings.Trim(s.Find("td").Eq(1).Text(), "\r\n\t "),
			Payment: strings.Trim(s.Find("td").Eq(2).Text(), "\r\n\t "),
			Qty:     qty,
			Price:   price,
			Total:   total,
			Date:    strings.Trim(s.Find("td").Eq(6).Text(), "\r\n\t "),
		}
		sales = append(sales, sale)
	})
	profile.Sales = sales

	api.data.SetProfile(token, profile)
	return profile, nil
}

func (api *Api) Products(token string) ([]*Product, error) {
	client := api.getClient(token)

	doc, err := api.get(client, api.baseUrl.String()+"/")
	if err != nil {
		return nil, err
	}

	products := []*Product{}

	doc.Find(".imgList").Each(func(i int, s *goquery.Selection) {
		name := s.Find("strong").Text()

		id := ""
		qty := 0
		price := 0

		href, exists := s.Find("a").Attr("href")
		if exists {
			id = strings.Split(href, "item=")[1]

			sub := s.Find(".subheader").Text()
			re := regexp.MustCompile(`^([^ ]+)\s+Pcs\s+Rp\.(\d+)`)
			res := re.FindAllStringSubmatch(sub, -1)

			qty, _ = strconv.Atoi(res[0][1])
			price, _ = strconv.Atoi(res[0][2])
		}

		product := &Product{
			Id:    id,
			Name:  name,
			Qty:   qty,
			Price: price,
		}

		products = append(products, product)
	})

	//doc.Find("select[name=item] option").Each(func(i int, s *goquery.Selection) {
	//	id, _ := s.Attr("value")
	//	name := strings.Trim(s.Text(), "\n\r\t ")
	//	if name != "---" {
	//		product := &Product{
	//			id:   id,
	//			name: name,
	//		}
	//		products = append(products, product)
	//	}
	//})

	return products, nil
}

//func (a *Api) Connected() bool {
//	res, err := a.client.Get(a.base + "/sales")
//	if err != nil {
//		return false
//	}
//
//	if res.StatusCode == 200 {
//		return true
//	}
//
//	return false
//}
//
//
//func (a *Api) Balance() (int, error) {
//	res, err := a.client.Get(a.base + "/sales")
//	if err != nil {
//		return 0, err
//	}
//	body, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		return 0, err
//	}
//
//	x := regexp.MustCompile("Hutang Anda : Rp.\\s+(\\d+)<").FindAllStringSubmatch(string(body), -1)
//
//	balance, _ := strconv.Atoi(x[0][1])
//	return balance, nil
//}
//
//func (a *Api) Write() error {
//	data := a.username + "|" + a.password + "|" + a.token + "\n"
//	ioutil.WriteFile("sikopat.cfg", []byte(data), 0644)
//	return nil
//}
//

//func (a *Api) Sales(token stringV) ([]*Sale, error) {
//	res, err := a.client.Get(a.base + "/sales")
//	if err != nil {
//		return nil, err
//	}
//
//	doc, err := goquery.NewDocumentFromResponse(res)
//	if err != nil {
//		return nil, err
//	}
//
//sales := []*Sale{}
//
//	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
//		qty, _ := strconv.Atoi(strings.Trim(s.Find("td").Eq(3).Text(), "\n\r\t "))
//		price, _ := strconv.Atoi(strings.Trim(s.Find("td").Eq(4).Text(), "\n\r\t "))
//		total, _ := strconv.Atoi(strings.Trim(s.Find("td").Eq(5).Text(), "\n\r\t "))
//		sale := &Sale{
//			Product: strings.Trim(s.Find("td").Eq(1).Text(), "\n\r\t "),
//			Payment: strings.ToUpper(strings.Trim(s.Find("td").Eq(2).Text(), "\n\r\t "))[0:2],
//			Qty:     qty,
//			Price:   price,
//			Total:   total,
//			Date:    strings.Trim(s.Find("td").Eq(6).Text(), "\n\r\t "),
//		}
//		sales = append(sales, sale)
//	})
//
//return sales, nil
//}

func New(baseUrl string, dataFile string) (*Api, error) {
	baseUrl = strings.Trim(baseUrl, "\n\r\t ")
	dataFile = strings.Trim(dataFile, "\n\r\t ")

	if baseUrl == "" {
		return nil, errors.New("Base URL must be url of sikopat")
	}

	if dataFile == "" {
		dataFile = regexp.MustCompile(`[/\:. _]`).ReplaceAllString(baseUrl, "-")
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	a := &Api{
		baseUrl: u,
		data:    NewData(dataFile),
		clients: make(map[string]*http.Client),
	}

	return a, nil
}
