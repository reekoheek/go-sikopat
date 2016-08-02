package api

import (
	"encoding/json"
	"time"
)

type (
	Profile struct {
		Token       string
		Username    string
		Password    string
		RemoteToken string
		Balance     int
		Sales       *Sales
		Updated     time.Time
	}

	Profiles struct {
		data map[string]*Profile
	}
)

func (ms *Profiles) MarshalJSON() ([]byte, error) {
	return json.Marshal(ms.data)
}

func (ms *Profiles) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &ms.data)
}

func (ps *Profiles) Put(k string, v *Profile) {
	ps.data[k] = v
}

func (ps *Profiles) Get(k string) *Profile {
	return ps.data[k]
}

func (ps *Profiles) Delete(k string) {
	delete(ps.data, k)
}

func NewProfile(username string, password string) *Profile {
	return &Profile{
		Username: username,
		Password: password,
	}
}

func NewProfiles() *Profiles {
	return &Profiles{
		data: map[string]*Profile{},
	}
}
