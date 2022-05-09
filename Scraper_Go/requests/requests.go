package requests

import "net/http"

var hosts = [...]string{"dacardworld", "steelcitycollectibles", "tcgplayer", "trollandtoad", "collectorstore"}
var Clients map[string]RestClient

func InitClients(){
	for _, host := range hosts{
		Clients[host] = RestClient{HTTPClient: http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		}}
	}
}

func newClient(url string) string{

}