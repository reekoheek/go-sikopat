package api

type (
	Info struct {
		data map[string]string
	}
)

func (info *Info) Put(k, v string) {
	info.data[k] = v
}

func (info *Info) Get(k string) string {
	return info.data[k]
}

func (info *Info) ForEach(fn func(k, v string)) {
	for k, v := range info.data {
		fn(k, v)
	}
}

func NewInfo() *Info {
	return &Info{
		data: map[string]string{},
	}
}
