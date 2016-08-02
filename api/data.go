package api

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"time"
)

type (
	Data struct {
		Profiles      *Profiles
		Products      *Products
		Methods       *PaymentMethods
		DefaultMethod string
		Updated       time.Time
		file          string
	}
)

func (d *Data) PaymentMethods() *PaymentMethods {
	return d.Methods
}

func (d *Data) DefaultPaymentMethod() *PaymentMethod {
	if d.DefaultMethod == "" {
		return nil
	}
	return d.Methods.Get(d.DefaultMethod)
}

func (d *Data) SetPaymentMethods(methods *PaymentMethods, asUser bool) error {
	if (d.Methods != nil && d.Methods.Len() > 0) && !asUser {
		return nil
	}

	if foundMethod := methods.Find(func(k string, method *PaymentMethod) bool {
		return method.Name == "Credit"
	}); foundMethod != nil {
		d.DefaultMethod = foundMethod.Id
	}

	d.Methods = methods
	d.Updated = time.Now()

	return d.persist()
}

func (d *Data) ProductsByFilter(filter string) *Products {
	if len(filter) <= 0 {
		return d.Products
	}

	reStr := "(?i)"
	for _, v := range filter {
		reStr = reStr + string(v) + ".*"
	}
	reStr = reStr[0 : len(reStr)-2]
	re := regexp.MustCompile(reStr)

	products := NewProducts()

	for id, product := range d.Products.Iterator() {
		if re.MatchString(product.Name) {
			products.Put(id, product)
		}
	}

	return products
}

func (d *Data) SetProducts(products *Products, asUser bool) error {
	if (d.Products != nil && d.Products.Len() > 0) && !asUser {
		return nil
	}

	d.Products = products
	d.Updated = time.Now()

	return d.persist()
}

func (d *Data) UnsetProfile(token string) error {
	d.Profiles.Delete(token)

	return d.persist()
}

func (d *Data) SetProfile(token string, profile *Profile) error {
	profile.Updated = time.Now()

	d.Profiles.Put(token, profile)

	return d.persist()
}

func (d *Data) Profile(token string) *Profile {
	return d.Profiles.Get(token)
}

func (d *Data) persist() error {
	contentBytes, _ := json.Marshal(d)

	return ioutil.WriteFile(d.file, contentBytes, 0644)
}

func NewData(file string) *Data {
	data := &Data{
		Profiles: NewProfiles(),
		file:     file,
	}

	//log.Printf("Read data from %s", file)
	content, err := ioutil.ReadFile(file)
	if err == nil {
		json.Unmarshal(content, data)
	}

	return data
}
