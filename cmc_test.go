package cmc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTickers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		body    string
		tickers []*Ticker
	}{
		{
			body: `
[
    {
        "id": "bitcoin",
        "name": "Bitcoin",
        "symbol": "BTC",
        "rank": "1",
        "price_usd": "573.137",
        "price_btc": "1.0",
        "24h_volume_usd": "72855700.0",
        "market_cap_usd": "9080883500.0",
        "available_supply": "15844176.0",
        "total_supply": "15844176.0",
        "percent_change_1h": "0.04",
        "percent_change_24h": "-0.3",
        "percent_change_7d": "-0.57",
        "last_updated": "1472762067"
    },
    {
        "id": "ethereum",
        "name": "Ethereum",
        "symbol": "ETH",
        "rank": "2",
        "price_usd": "12.1844",
        "price_btc": "0.021262",
        "24h_volume_usd": "24085900.0",
        "market_cap_usd": "1018098455.0",
        "available_supply": "83557537.0",
        "total_supply": "83557537.0",
        "percent_change_1h": "-0.58",
        "percent_change_24h": "6.34",
        "percent_change_7d": "8.59",
        "last_updated": "1472762062"
    }
]`,
			tickers: []*Ticker{
				{
					ID:               "bitcoin",
					Name:             "Bitcoin",
					Symbol:           "BTC",
					Rank:             1,
					PriceUSD:         "573.137",
					PriceBTC:         "1.0",
					VolumeUSD24hr:    "72855700.0",
					MarketCapUSD:     "9080883500.0",
					AvailableSupply:  "15844176.0",
					TotalSupply:      "15844176.0",
					PercentChange1h:  "0.04",
					PercentChange24h: "-0.3",
					PercentChange7d:  "-0.57",
					LastUpdated:      "1472762067",
				},
				{
					ID:               "ethereum",
					Name:             "Ethereum",
					Symbol:           "ETH",
					Rank:             2,
					PriceUSD:         "12.1844",
					PriceBTC:         "0.021262",
					VolumeUSD24hr:    "24085900.0",
					MarketCapUSD:     "1018098455.0",
					AvailableSupply:  "83557537.0",
					TotalSupply:      "83557537.0",
					PercentChange1h:  "-0.58",
					PercentChange24h: "6.34",
					PercentChange7d:  "8.59",
					LastUpdated:      "1472762062",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					_, _ = rw.Write([]byte(test.body))
				}),
			)

			client := NewWebClient(WithAPIAddress(server.URL))

			tickers, err := client.Tickers(nil)
			require.NoError(t, err, "failed to get tickers")
			require.ElementsMatch(t, test.tickers, tickers, "unexpected response")
		})
	}
}

func TestIntegration(t *testing.T) {
	client := NewWebClient()

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "\t")

	fmt.Println("getting Tickers:")
	tickers, err := client.Tickers(nil)
	require.NoError(t, err)
	err = encoder.Encode(tickers)
	require.NoError(t, err)

	fmt.Println("getting Ticker:")
	ticker, err := client.Ticker("bitcoin")
	require.NoError(t, err)
	err = encoder.Encode(ticker)
	require.NoError(t, err)
}
