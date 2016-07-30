package api

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"time"
)

type (
	Data struct {
		Profiles      map[string]*Profile
		Products      map[string]*Product
		Methods       map[string]*PaymentMethod
		DefaultMethod string
		Updated       time.Time
		file          string
	}

	PaymentMethod struct {
		Id   string
		Name string
	}

	Product struct {
		Id    string
		Name  string
		Qty   int
		Price int
	}

	Profile struct {
		Token       string
		Username    string
		Password    string
		RemoteToken string
		Balance     int
		Sales       []*Sale
		Updated     time.Time
	}

	Sale struct {
		Product string
		Payment string
		Qty     int
		Price   int
		Total   int
		Date    string
	}
)

func (d *Data) PaymentMethods() map[string]*PaymentMethod {
	return d.Methods
}

func (d *Data) DefaultPaymentMethod() *PaymentMethod {
	if d.DefaultMethod == "" {
		return nil
	}
	return d.Methods[d.DefaultMethod]
}

func (d *Data) SetPaymentMethods(methods map[string]*PaymentMethod, asUser bool) error {
	if len(d.Methods) > 0 && !asUser {
		return nil
	}

	for _, method := range methods {
		if method.Name == "Credit" {
			d.DefaultMethod = method.Id
			break
		}
	}

	d.Methods = methods
	d.Updated = time.Now()

	return d.persist()
}

func (d *Data) ProductsByFilter(filter string) map[string]*Product {
	if len(filter) <= 0 {
		return d.Products
	}

	reStr := "(?i)"
	for _, v := range filter {
		reStr = reStr + string(v) + ".*"
	}
	reStr = reStr[0 : len(reStr)-2]
	re := regexp.MustCompile(reStr)

	products := make(map[string]*Product)

	for id, product := range d.Products {
		if re.MatchString(product.Name) {
			products[id] = product
		}
	}

	return products
}

func (d *Data) SetProducts(products map[string]*Product, asUser bool) error {
	if len(d.Products) > 0 && !asUser {
		return nil
	}

	d.Products = products
	d.Updated = time.Now()

	return d.persist()
}

func (d *Data) UnsetProfile(token string) error {
	delete(d.Profiles, token)

	return d.persist()
}

func (d *Data) SetProfile(token string, profile *Profile) error {
	if profile.Sales == nil {
		profile.Sales = make([]*Sale, 0)
	}
	profile.Updated = time.Now()

	d.Profiles[token] = profile

	return d.persist()
}

func (d *Data) Profile(token string) *Profile {
	return d.Profiles[token]
}

func (d *Data) persist() error {
	contentBytes, _ := json.Marshal(d)

	return ioutil.WriteFile(d.file, contentBytes, 0644)
}

func NewData(file string) *Data {
	data := &Data{
		Profiles: make(map[string]*Profile),
		file:     file,
	}

	content, err := ioutil.ReadFile(file)
	if err == nil {
		json.Unmarshal(content, data)
	}

	return data
	//
	//	tokens := strings.Split(strings.Trim(string(data), "\n\r\t "), "|")
	//	a.username = tokens[0]
	//	a.password = tokens[1]
	//	a.token = tokens[2]
	//
	//	u, err := url.Parse(a.base)
	//	if err != nil {
	//		return err
	//	}
	//
	//	cookie := &http.Cookie{
	//		Name:   "BSESS",
	//		Value:  a.token,
	//		Path:   "/",
	//		Domain: "sikopat.xinix.co.id",
	//	}
	//	cookies := []*http.Cookie{cookie}
	//
	//	a.client.Jar.SetCookies(u, cookies)
	//
	return nil
}
