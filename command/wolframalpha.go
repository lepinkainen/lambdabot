package command

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"

	"github.com/lepinkainen/lambdabot/lambda"

	log "github.com/sirupsen/logrus"
)

// WolframAlpha queries Wolfram Alpha for answers
func WolframAlpha(args string) (string, error) {

	appid := os.Getenv("WOLFRAM_ALPHA_API_KEY")
	query := url.QueryEscape(args)
	apiurl := fmt.Sprintf("http://api.wolframalpha.com/v1/result?appid=%s&units=metric&i=%s", appid, query)

	res, err := http.Get(apiurl)
	if err != nil {
		log.Errorf("Unable to get API response from WolframAlpha: %v", err)
		return "", errors.Wrap(err, "Unable to get API response")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Unable to read response from WolframAlpha: %v", err)
		return "", errors.Wrap(err, "Unable to read response")
	}

	return fmt.Sprintf("%s = %s", args, string(body)), nil
}

func init() {
	lambda.RegisterHandler("wa", WolframAlpha)
}
