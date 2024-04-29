package utils

type Configuration struct {
	Key     string `json:"key"`
	Debug   bool   `json:"debug"`
	KeyUsed bool   `json:"key_used"`
}

func (w *Configuration) SetDebug(value bool) {
	w.Debug = value
}

func (w *Configuration) SetKey(value string) {
	w.Key = value
}

func (w *Configuration) UseKey() {
	w.KeyUsed = true
}
