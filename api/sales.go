package api

import "encoding/json"

type (
	Sale struct {
		Product string
		Payment string
		Qty     int
		Price   int
		Total   int
		Date    string
	}

	Sales struct {
		data []*Sale
	}
)

func (ms *Sales) MarshalJSON() ([]byte, error) {
	return json.Marshal(ms.data)
}

func (ms *Sales) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &ms.data)
}

func (ss *Sales) Add(sale *Sale) {
	ss.data = append(ss.data, sale)
}

func (ss *Sales) Iterator() []*Sale {
	return ss.data
}

func NewSales() *Sales {
	return &Sales{
		data: []*Sale{},
	}
}
