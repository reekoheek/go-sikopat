package api

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type (
	Data struct {
		Profiles map[string]*Profile

		file string
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
)

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
