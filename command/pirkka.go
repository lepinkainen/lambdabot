package command

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/lepinkainen/lambdabot/lambda"
)

type Price struct {
	Price float64 `json:"price"`
}

// Grab Pirkka price from API and respond
func Pirkka(_ string) (string, error) {
	url := "https://juho.tech/api/pirkka_price"

	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var price Price
	err = json.Unmarshal(body, &price)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("Pirkka olut: %.2f €", price.Price), nil
}

func init() {
	lambda.RegisterHandler("pirkka", Pirkka)
}
