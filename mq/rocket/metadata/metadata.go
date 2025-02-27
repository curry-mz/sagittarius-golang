package metadata

const (
	_uberCtxServiceKey = "_uber_ctx_service_key"
)

type MetaMap struct {
	data map[string]string
}

func NewMetaMap() *MetaMap {
	return &MetaMap{
		data: make(map[string]string),
	}
}

func NewMetaMapWithData(data map[string]string) *MetaMap {
	return &MetaMap{
		data: data,
	}
}

func (mm *MetaMap) SetUberMeta(sk string) {
	mm.data[_uberCtxServiceKey] = sk
}

func (mm *MetaMap) GetUberMeta() string {
	return mm.data[_uberCtxServiceKey]
}

func (mm *MetaMap) Set(key, val string) {
	mm.data[key] = val
}

func (mm MetaMap) ForeachKey(handler func(key, val string) error) error {
	for k, v := range mm.data {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (mm *MetaMap) Data() map[string]string {
	return mm.data
}
