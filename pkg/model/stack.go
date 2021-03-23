package model

type Stack struct {
	Current string `json:"current"`
	Name    string `json:"name"`
	Server  string `json:"server"`
	Kafka   string `json:"kafka"`
	KvStore string `json:"kvstore"`
}
