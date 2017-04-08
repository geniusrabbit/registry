package main

type config struct {
	Debug   bool `consul:"debug"`
	Service struct {
		Listen string `consul:"service/listen"`
		SSL    bool   `consul:"service/on"`
	}
}

func main() {

}
