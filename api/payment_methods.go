package api

import "encoding/json"

type (
	PaymentMethod struct {
		Id   string
		Name string
	}

	PaymentMethods struct {
		data map[string]*PaymentMethod
	}
)

func (ms *PaymentMethods) MarshalJSON() ([]byte, error) {
	return json.Marshal(ms.data)
}

func (ms *PaymentMethods) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &ms.data)
}

func (ms *PaymentMethods) Put(k string, v *PaymentMethod) {
	ms.data[k] = v
}

func (ms *PaymentMethods) Get(k string) *PaymentMethod {
	return ms.data[k]
}

func (ms *PaymentMethods) Len() int {
	return len(ms.data)
}

func (ms *PaymentMethods) Find(fn func(k string, v *PaymentMethod) bool) *PaymentMethod {
	for k, v := range ms.data {
		if fn(k, v) {
			return v
		}
	}
	return nil
}

func NewPaymentMethods() *PaymentMethods {
	return &PaymentMethods{
		data: map[string]*PaymentMethod{},
	}
}
