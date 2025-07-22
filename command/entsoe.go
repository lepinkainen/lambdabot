package command

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

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

// GetPriceString retrieves the current, lowest, and highest prices in c/kWh from the entso-e API.
//
// It does this by making an HTTP request to the API, parsing the XML response, and calculating the prices.
// The API key is retrieved from the environment variable "ENTSOE_API_KEY".
//
// Return:
//   - A string containing the current, lowest, and highest prices in the format "Current: {currentPrice} c/kWh | Lowest: {lowestPrice} c/kWh | Highest: {highestPrice} c/kWh".
//   - An error if there is an issue making the HTTP request, reading the HTTP response,
//     unmarshaling the XML, or any other error that occurs during the process.
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
	var highestPrice = PricePoint{Price: -999999999.0}
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

			// price at point is lower than lowest, set to lowest
			if point.PriceAmount < lowestPrice.Price {
				lowestPrice = pricePoint
			}

			// price at point is higher than highest, set to highest
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
func Entsoe(_ string) (string, error) {
	return GetPriceString()
}

var ctx = context.Background()

const DBNAME = "entsoe:fi"

// EntsoeRedis retrieves current, lowest, and highest prices from Redis timeseries
//
// # A separate system is ran in cron to update the timeseries data from entso-e
//
// args: A string representing the arguments for the Redis client.
// Returns: A string containing the formatted current, lowest, and highest prices in c/kWh, and an error if any.
func EntsoeRedis(_ string) (string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: "default",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})

	res := rdb.Ping(ctx)
	if res.Err() != nil {
		return "", res.Err()
	}

	var currentPrice = 999999999.0
	var dayLow = 999999999.0
	var dayHigh = -999999999.0

	// This time we want times specifically in the Finnish time zone
	location, _ := time.LoadLocation("Europe/Helsinki")

	// current time to 1hr accuracy
	now := time.Now().In(location).Truncate(time.Hour)
	nowUnix := int(now.UnixMilli()) // redis likes unix milliseconds

	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	endOfDay := startOfDay.Add(23 * time.Hour).Add(59 * time.Minute).Add(59 * time.Second) // just before midnight

	// day max
	// ts.range entsoe:fi 1705356000000 1705442400000 AGGREGATION max 3600000000
	// day min
	// ts.range entsoe:fi 1705356000000 1705442400000 AGGREGATION min 3600000000
	// NOTE: subtract 100000 from end time to get only 24 hours, otherwise you'll get 25 hours and bad data

	// grab value for the current time from redis
	valueSlice := rdb.TSRange(ctx, DBNAME, nowUnix, nowUnix)
	if valueSlice.Err() != nil {
		fmt.Printf("Error getting values from Redis: %+v\n", valueSlice.Err())
		return "", valueSlice.Err()
	}
	tsValue, _ := valueSlice.Result()
	if len(tsValue) > 0 {
		currentPrice = tsValue[0].Value
	}

	// Grab maximum from redis
	valueSlice = rdb.TSRangeWithArgs(ctx, DBNAME, int(startOfDay.UnixMilli()), int(endOfDay.UnixMilli()), &redis.TSRangeOptions{
		Aggregator:     redis.Max,
		BucketDuration: 3600000000,
	})
	if valueSlice.Err() != nil {
		fmt.Printf("Error getting values from Redis: %+v\n", valueSlice.Err())
		return "", valueSlice.Err()
	}
	tsValue, _ = valueSlice.Result()
	if len(tsValue) > 0 {
		dayHigh = tsValue[0].Value
	}

	// Grab minimum from redis
	valueSlice = rdb.TSRangeWithArgs(ctx, DBNAME, int(startOfDay.UnixMilli()), int(endOfDay.UnixMilli()), &redis.TSRangeOptions{
		Aggregator:     redis.Min,
		BucketDuration: 3600000000,
	})
	if valueSlice.Err() != nil {
		fmt.Printf("Error getting values from Redis: %+v\n", valueSlice.Err())
		return "", valueSlice.Err()
	}
	tsValue, _ = valueSlice.Result()
	if len(tsValue) > 0 {
		dayLow = tsValue[0].Value
	}

	// TODO: dayLow and dayHigh are not working properly
	//  Current: 11.86 c/kWh | Lowes @ 08: 10.18 c/kWh | Highest @ 08: 18.60 c/kWh
	//
	// Should be a range like 07-08 and 15-16

	return fmt.Sprintf("Current: %.2f c/kWh | Lowest: %.2f c/kWh | Highest: %.2f c/kWh", currentPrice*1.24, dayLow*1.24, dayHigh*1.24), nil
}

/*
func unixToHourString(unix int64) string {
	// converting milliseconds to seconds
	timestamp := unix / 1000

	// creating a new time from Unix timestamp
	t := time.Unix(int64(timestamp), 0)

	// formatting the time to get only the hour in 24-hour format
	hour := t.Format("15")

	return hour
}
*/

func init() {
	lambda.RegisterHandler("sahko", EntsoeRedis)
	lambda.RegisterHandler("sahko2", Entsoe)
}
