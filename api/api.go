package api

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/satori/go.uuid"
)

const (
	VERSION = "0.1.1"
)

var (
	Non200Err       = errors.New("Non 200 Status Code")
	NotLoginErr     = errors.New("Not login yet")
	LoginFailedErr  = errors.New("Login failed")
	QtyPriceRe      = regexp.MustCompile(`^([^ ]+)\s+Pcs\s+Rp\.(\d+)`)
	BalanceRe       = regexp.MustCompile(`Hutang Anda : Rp.\s+(\d+)`)
	UrlToFilenameRe = regexp.MustCompile(`[/\:. _]`)
)

type (
	Api struct {
		baseUrl      *url.URL
		dir          string
		data         *Data
		commonClient *http.Client
		clients      map[string]*http.Client
	}
)

func (api *Api) Info(token string) map[string]string {
	infoMap := map[string]string{
		"api.version": VERSION,
		"api.url":     api.baseUrl.String(),
		"api.dir":     api.dir,
	}
	return infoMap
}

func (api *Api) get(client *http.Client, uri string) (*goquery.Document, error) {
	//log.Println("get", uri)
	res, err := client.Get(uri)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, Non200Err
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
	if token != "" {
		if api.clients[token] != nil {
			return api.clients[token]
		}

		profile := api.data.Profile(token)
		if profile != nil {
			client := api.createClient()

			cookie := &http.Cookie{
				Name:   "BSESS",
				Value:  profile.RemoteToken,
				Path:   "/",
				Domain: api.baseUrl.Host,
			}
			client.Jar.SetCookies(api.baseUrl, []*http.Cookie{cookie})

			if _, err := api.get(client, api.baseUrl.String()+"/sales"); err == nil {
				api.clients[token] = client
				return client
			}

			if _, err := api.Login(profile); err == nil {
				if api.clients[token] != nil {
					return api.clients[token]
				}
			}

		}
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

func (api *Api) Login(profile *Profile) (string, error) {
	client := api.createClient()

	form := url.Values{
		"username": {profile.Username},
		"password": {profile.Password},
	}

	res, err := client.PostForm(api.baseUrl.String()+"/login", form)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", LoginFailedErr
	}

	cookie := client.Jar.Cookies(api.baseUrl)
	remoteToken := cookie[0].Value

	if profile.Token == "" {
		profile.Token = uuid.NewV1().String()
	}
	profile.RemoteToken = remoteToken

	api.clients[profile.Token] = client
	//profile := &Profile{
	//	Username:    username,
	//	Password:    password,
	//	Token:       token,
	//	RemoteToken: token,
	//}
	api.data.SetProfile(profile.Token, profile)

	_, err = api.Profile(profile.Token, true)
	return profile.Token, err
}

func (api *Api) Buy(token string, product *Product, qty int) error {
	client := api.getClient(token)
	if !api.isUserClient(client) {
		return NotLoginErr
	}

	var (
		method *PaymentMethod
		err    error
		qtyStr string
	)

	if method, err = api.DefaultPaymentMethod(token); err != nil {
		return err
	}

	qtyStr = strconv.Itoa(qty)

	form := url.Values{
		"item":     {product.Id},
		"payment":  {method.Id},
		"quantity": {qtyStr},
	}

	_, err = client.PostForm(api.baseUrl.String()+"/sales/null/create", form)
	return err
}

func (api *Api) Profile(token string, force bool) (*Profile, error) {
	profile := api.data.Profile(token)
	//if profile == nil {
	//	return nil, NotLoginErr
	//} else if !force {
	//	return profile, nil
	//}

	client := api.getClient(token)
	if !api.isUserClient(client) {
		return nil, NotLoginErr
	}

	doc, err := api.get(client, api.baseUrl.String()+"/sales")
	if err != nil {
		return profile, err
	}

	debt := doc.Find(".hutang").Text()
	result := BalanceRe.FindAllStringSubmatch(debt, -1)
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

func (api *Api) Products(token string, filter string) (map[string]*Product, error) {
	if err := api.syncCommonData(token); err != nil {
		return nil, err
	}

	products := api.data.ProductsByFilter(filter)
	return products, nil
}

func (api *Api) DefaultPaymentMethod(token string) (*PaymentMethod, error) {
	if err := api.syncCommonData(token); err != nil {
		return nil, err
	}

	paymentMethod := api.data.DefaultPaymentMethod()
	return paymentMethod, nil
}

func (api *Api) PaymentMethods(token string) (map[string]*PaymentMethod, error) {
	if err := api.syncCommonData(token); err != nil {
		return nil, err
	}

	paymentMethods := api.data.PaymentMethods()
	return paymentMethods, nil
}

func (api *Api) syncCommonData(token string) error {
	client := api.getClient(token)

	doc, err := api.get(client, api.baseUrl.String()+"/")
	if err != nil {
		return err
	}

	products := map[string]*Product{}

	doc.Find(".imgList").Each(func(i int, s *goquery.Selection) {
		name := s.Find("strong").Text()

		id := ""
		qty := 0
		price := 0

		href, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}

		id = strings.Split(href, "item=")[1]

		sub := s.Find(".subheader").Text()
		res := QtyPriceRe.FindAllStringSubmatch(sub, -1)

		qty, _ = strconv.Atoi(res[0][1])
		price, _ = strconv.Atoi(res[0][2])

		product := &Product{
			Id:    id,
			Name:  name,
			Qty:   qty,
			Price: price,
		}

		products[id] = product
	})

	asUser := api.isUserClient(client)

	if asUser {
		doc, err := api.get(client, api.baseUrl.String()+"/sales/null/create")
		if err == nil {
			doc.Find("select[name=item] option").Each(func(i int, s *goquery.Selection) {
				value, _ := s.Attr("value")
				if value == "" {
					return
				}
				label := strings.Trim(s.Text(), "\n\r\t ")
				product := products[value]
				if product != nil {
					product.Name = label
				}
			})

			methods := map[string]*PaymentMethod{}

			doc.Find("select[name=payment] option").Each(func(i int, s *goquery.Selection) {
				value, _ := s.Attr("value")
				if value == "" {
					return
				}

				label := strings.Trim(s.Text(), "\n\r\t ")

				method := &PaymentMethod{
					Id:   value,
					Name: label,
				}

				methods[value] = method
			})

			api.data.SetPaymentMethods(methods, asUser)
		}
	}

	api.data.SetProducts(products, asUser)

	return nil
}

func New(baseUrl string, dataDir string, dataFile string) (*Api, error) {
	baseUrl = strings.Trim(baseUrl, "\n\r\t ")
	dataDir = strings.Trim(dataDir, "\n\r\t ")
	dataFile = strings.Trim(dataFile, "\n\r\t ")

	if baseUrl == "" {
		return nil, errors.New("Base URL must be url of sikopat")
	}

	if dataFile == "" {
		dataFile = UrlToFilenameRe.ReplaceAllString(baseUrl, "-")
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	a := &Api{
		baseUrl: u,
		dir:     dataDir,
		data:    NewData(filepath.Join(dataDir, dataFile)),
		clients: make(map[string]*http.Client),
	}

	return a, nil
}
