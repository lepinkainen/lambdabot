package command

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/lepinkainen/lambdabot/lambda"
)

type PublicationMarketDocument struct {
	XMLName                   xml.Name          `xml:"Publication_MarketDocument"`
	MRID                      string            `xml:"mRID"`
	RevisionNumber            int               `xml:"revisionNumber"`
	Type                      string            `xml:"type"`
	SenderMarketParticipant   MarketParticipant `xml:"sender_MarketParticipant"`
	ReceiverMarketParticipant MarketParticipant `xml:"receiver_MarketParticipant"`
	CreatedDateTime           time.Time         `xml:"createdDateTime"`
	PeriodTimeInterval        TimeInterval      `xml:"period.timeInterval"`
	TimeSeries                []TimeSeries      `xml:"TimeSeries"`
}

type MarketParticipant struct {
	MRID         string `xml:"mRID,attr"`
	CodingScheme string `xml:"codingScheme,attr"`
	MarketRole   struct {
		Type string `xml:"type"`
	} `xml:"marketRole"`
}

type TimeInterval struct {
	Start string `xml:"start"`
	End   string `xml:"end"`
}

type TimeSeries struct {
	MRID         string `xml:"mRID"`
	BusinessType string `xml:"businessType"`
	InDomainMRID struct {
		CodingScheme string `xml:"codingScheme,attr"`
		MRID         string `xml:",chardata"`
	} `xml:"in_Domain.mRID"`
	OutDomainMRID struct {
		CodingScheme string `xml:"codingScheme,attr"`
		MRID         string `xml:",chardata"`
	} `xml:"out_Domain.mRID"`
	CurrencyUnitName     string `xml:"currency_Unit.name"`
	PriceMeasureUnitName string `xml:"price_Measure_Unit.name"`
	CurveType            string `xml:"curveType"`
	Period               struct {
		TimeInterval TimeInterval `xml:"timeInterval"`
		Resolution   string       `xml:"resolution"`
		Points       []Point      `xml:"Point"`
	} `xml:"Period"`
}

type Point struct {
	Position    int     `xml:"position"`
	PriceAmount float64 `xml:"price.amount"`
}

type PricePoint struct {
	Time  string
	Price float64
}

func GetPriceString() (string, error) {

	now := time.Now()

	apikey := os.Getenv("ENTSOE_API_KEY")

	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Format("200601020000")
	end := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Format("200601020000")

	url := fmt.Sprintf("https://web-api.tp.entsoe.eu/api?securityToken=%s&documentType=A44&out_Domain=10YFI-1--------U&in_Domain=10YFI-1--------U&periodStart=%s&periodEnd=%s", apikey, start, end)

	resp, err := http.Get(url)
	if err != nil {
		return "Error making HTTP request", err
	}
	defer resp.Body.Close()

	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Error reading HTTP response", err
	}

	var doc PublicationMarketDocument
	err = xml.Unmarshal(xmlData, &doc)
	if err != nil {
		return "Error unmarshaling XML", err
	}

	var lowestPrice = PricePoint{Price: 999999999.0}
	var highestPrice = PricePoint{Price: 0.0}
	var currentPrice PricePoint

	layout := "2006-01-02T15:04Z"

	for _, timeserie := range doc.TimeSeries {
		// start of this timeseries
		startStr := timeserie.Period.TimeInterval.Start
		start, _ := time.Parse(layout, startStr)

		for _, point := range timeserie.Period.Points {

			// the actual time of the point
			pointTime := start.Add(time.Duration(point.Position-1) * time.Hour)

			// timestamp and actual c/kWh price (VAT included)
			pricePoint := PricePoint{
				Time:  pointTime.Format(layout),
				Price: point.PriceAmount / 10 * 1.24,
			}

			if point.PriceAmount < lowestPrice.Price {
				lowestPrice = pricePoint

			}
			if point.PriceAmount > highestPrice.Price {
				highestPrice = pricePoint
			}

			// Check if the pointTime matches the current time, without the minute and second
			now := time.Now().Truncate(time.Hour)
			if pointTime.Equal(now) {
				currentPrice = pricePoint
			}
		}
	}

	return fmt.Sprintf("Current: %.2f c/kWh | Lowest: %.2f c/kWh | Highest: %.2f c/kWh", currentPrice.Price, lowestPrice.Price, highestPrice.Price), nil
}

// Echo echoes the arguments back
func Entsoe(args string) (string, error) {
	return GetPriceString()
}

func init() {
	lambda.RegisterHandler("sahko", Entsoe)
}
