package cmc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"astuart.co/goq"
)

const (
	apiAddr     = "https://api.coinmarketcap.com/v1"
	websiteAddr = "https://coinmarketcap.com/currencies"
)

// Client is the interface to the CoinMarketCap API.
type Client interface {
	// Tickers returns information for coin tickers.
	Tickers(*TickersOptions) ([]*Ticker, error)
	// Ticker returns information about a given coin specified by its ID.
	Ticker(id string) (*Ticker, error)

	// Prices returns historical price data.
	Prices(id string, _ *PricesOptions) ([]*Price, error)
	// Markets returns information about how much volume is happening on
	// each exchange.
	Markets(id string) ([]*Market, error)
}

// Ticker contains all the information available about a coin.
type Ticker struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Symbol           string `json:"symbol"`
	Rank             int    `json:"rank,string"`
	PriceUSD         string `json:"price_usd"`
	PriceBTC         string `json:"price_btc"`
	VolumeUSD24hr    string `json:"24h_volume_usd"`
	MarketCapUSD     string `json:"market_cap_usd"`
	SupplyAvailable  string `json:"available_supply"`
	SupplyTotal      string `json:"total_supply"`
	PercentChange1h  string `json:"percent_change_1h"`
	PercentChange24h string `json:"percent_change_24h"`
	PercentChange7d  string `json:"percent_change_7d"`
	LastUpdated      string `json:"last_updated"`
}

// TickersOptions allows for customizing the response from the Tickers call.
type TickersOptions struct {
	Start int
	Limit int
}

// Price contains information about the price of a cryptocurrency.
type Price struct {
	Date          string `json:"date"`
	OpenUSD       string `json:"open"`
	HighUSD       string `json:"high_usd"`
	LowUSD        string `json:"low_usd"`
	CloseUSD      string `json:"close_usd"`
	VolumeUSD24hr string `json:"24hr_volume_usd"`
	MarketCapUSD  string `json:"market_cap_usd"`
}

// PricesOptions allows you to specify start and end date for historical price
// data.
type PricesOptions struct {
	Start string
	End   string
}

// Market conotains information about a market a cryptocurrency is sold on.
type Market struct {
	Source string `json:"source"`
	Pair   string `json:"pair"`
	// VolumeUSD24hr    string `json:"24hr_volume_usd,omitempty"`
	// PriceUSD         string `json:"price_usd,omitempty"`
	VolumeUSD24hr    string `json:"-"`
	PriceUSD         string `json:"-"`
	VolumePercentage string `json:"volume_percent"`
}

// WebClient implements Client by calling an HTTP server that is the CoinMarketCap API.
type WebClient struct {
	apiAddr     string
	websiteAddr string
	httpClient  *http.Client
}

var _ Client = (*WebClient)(nil)

// NewWebClient instantiates a new CoinMarketCap API client.
func NewWebClient(opts ...Option) *WebClient {
	cli := &WebClient{
		apiAddr:     apiAddr,
		websiteAddr: websiteAddr,
		httpClient:  http.DefaultClient,
	}

	for _, opt := range opts {
		opt(cli)
	}

	return cli
}

// Option allows for the customization of a WebClient.
type Option func(*WebClient)

// WithHTTPClient allows the specification of an HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(cli *WebClient) {
		cli.httpClient = httpClient
	}
}

// WithAPIAddress allows for the specification of a custom API address.
// This is useful for testing.
func WithAPIAddress(apiAddr string) Option {
	return func(cli *WebClient) {
		cli.apiAddr = apiAddr
	}
}

// Tickers returns available information about coin tickers. Leaving opts nil
// returns the top 100 results.
//
// start - return results from rank [start] and above
// limit - return a maximum of [limit] results (default is 100, use 0 to return all results)
func (cli *WebClient) Tickers(opts *TickersOptions) ([]*Ticker, error) {
	// build URL
	u := fmt.Sprintf("%s/ticker", cli.apiAddr)
	if opts != nil {
		q := url.Values{}
		q.Add("start", strconv.Itoa(opts.Start))
		q.Add("limit", strconv.Itoa(opts.Limit))
		u = fmt.Sprintf("%s?%s", u, q.Encode())
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	res, err := cli.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	tickers := []*Ticker{}
	decoder := json.NewDecoder(res.Body)
	if err = decoder.Decode(&tickers); err != nil {
		return nil, err
	}

	return tickers, nil
}

// Ticker returns information about one coin.
func (cli *WebClient) Ticker(id string) (*Ticker, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/ticker/%s", cli.apiAddr, id), nil)
	if err != nil {
		return nil, err
	}

	res, err := cli.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	tickers := []*Ticker{}
	decoder := json.NewDecoder(res.Body)
	if err = decoder.Decode(&tickers); err != nil {
		return nil, err
	}
	if len(tickers) != 1 {
		return nil, errors.New("unexpected error retreiving ticker")
	}

	return tickers[0], nil
}

// pricesQ contains the goquery queries for parsing out the historical price
// data from the HTML.
type pricesQ struct {
	Prices []struct {
		Date string    `goquery:"td.text-left"`
		Vals [6]string `goquery:"[data-format-value]"`
	} `goquery:"tbody tr"`
}

// Prices returns historical price data. This is not part of the
// standard API, so it involves scraping and is brittle.
func (cli *WebClient) Prices(id string, opts *PricesOptions) ([]*Price, error) {
	u := fmt.Sprintf("%s/%s/historical-data", cli.websiteAddr, id)
	if opts != nil {
		q := url.Values{}
		q.Add("start", opts.Start)
		q.Add("end", opts.End)
		u = fmt.Sprintf("%s?%s", u, q.Encode())
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	res, err := cli.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var _prices pricesQ
	err = goq.NewDecoder(res.Body).Decode(&_prices)
	if err != nil {
		return nil, err
	}

	clean := func(s string) string {
		return strings.Replace(s, ",", "", -1)
	}

	prices := make([]*Price, 0, 30)
	for _, p := range _prices.Prices {
		price := &Price{
			Date:          p.Date,
			OpenUSD:       clean(p.Vals[0]),
			HighUSD:       clean(p.Vals[1]),
			LowUSD:        clean(p.Vals[2]),
			CloseUSD:      clean(p.Vals[3]),
			VolumeUSD24hr: clean(p.Vals[4]),
			MarketCapUSD:  clean(p.Vals[5]),
		}

		prices = append(prices, price)
	}

	return prices, nil
}

type marketsQ struct {
	Markets []struct {
		SourcePair [2]string `goquery:"td a"`
		Vals       [3]string `goquery:"td span"`
	} `goquery:"tbody tr"`
}

// Markets returns information about how much volume is happening on
// each exchange. This involves scraping.
func (cli *WebClient) Markets(id string) ([]*Market, error) {
	req, err := http.NewRequest(
		"GET", fmt.Sprintf("%s/%s#markets", cli.websiteAddr, id), nil,
	)
	if err != nil {
		return nil, err
	}

	res, err := cli.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var _markets marketsQ
	err = goq.NewDecoder(res.Body).Decode(&_markets)
	if err != nil {
		return nil, err
	}

	markets := make([]*Market, 0, 5)
	for _, m := range _markets.Markets {
		market := &Market{
			Source:           m.SourcePair[0],
			Pair:             m.SourcePair[1],
			VolumeUSD24hr:    m.Vals[0],
			PriceUSD:         m.Vals[1],
			VolumePercentage: m.Vals[2],
		}

		markets = append(markets, market)
	}

	return markets, nil
}
