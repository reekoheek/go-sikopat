package api

import "encoding/json"

type (
	Product struct {
		Id    string
		Name  string
		Qty   int
		Price int
	}

	Products struct {
		data         map[string]*Product
		seriesCursor int
		series       []*Product
	}
)

func (ps *Products) Rewind() {
	ps.seriesCursor = 0
	ps.series = []*Product{}
	for _, v := range ps.data {
		ps.series = append(ps.series, v)
	}
}

func (ps *Products) HasNext() bool {
	return ps.series != nil && len(ps.series) > ps.seriesCursor
}

func (ps *Products) Next() *Product {
	if !ps.HasNext() {
		return nil
	}

	item := ps.series[ps.seriesCursor]
	ps.seriesCursor++
	return item
}

func (ps *Products) MarshalJSON() ([]byte, error) {
	return json.Marshal(ps.data)
}

func (ps *Products) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &ps.data)
}

func (ps *Products) Put(k string, v *Product) {
	ps.data[k] = v
}

func (ps *Products) Get(k string) *Product {
	return ps.data[k]
}

func (ps *Products) Iterator() map[string]*Product {
	return ps.data
}

func (ps *Products) Len() int {
	return len(ps.data)
}

func (ps *Products) Delete(k string) {
	delete(ps.data, k)
}

func NewProducts() *Products {
	return &Products{
		data: map[string]*Product{},
	}
}
