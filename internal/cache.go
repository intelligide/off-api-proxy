package internal

type Cache struct {
	DataMap map[string]int
	Data []string
	max int
}

func NewCache() *Cache {
	return &Cache {
		DataMap: make(map[string]int),
		max: 1000,
	}
}

func (this *Cache) Clear() {
	this.DataMap = make(map[string]int)
}

func (this *Cache) Add(key string, value string) {

	_, inCache := this.DataMap[key]
	if (!inCache) {
		this.Data = append(this.Data, value)
		this.DataMap[key] = len(this.Data) - 1
	} else {

	}
}
