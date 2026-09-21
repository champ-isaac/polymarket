// Command polymarket scans Polymarket's public Gamma API for multi-outcome
// event groups (e.g. "who will win X") where the group's Yes-leg ask prices
// sum to less than $1. Exactly one leg resolves Yes, so buying every Yes leg
// guarantees a $1 payout — a sum below $1 is a same-platform arbitrage edge,
// before fees and slippage. Read-only: it only reports numbers for you to
// review, it does not place orders or identify traders.
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

const gammaEventsURL = "https://gamma-api.polymarket.com/events"

type market struct {
	Question     string  `json:"question"`
	Outcomes     string  `json:"outcomes"`
	GroupItem    string  `json:"groupItemTitle"`
	NegRisk      bool    `json:"negRisk"`
	AcceptOrders bool    `json:"acceptingOrders"`
	BestAsk      float64 `json:"bestAsk"`
	BestBid      float64 `json:"bestBid"`
}

type event struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Slug    string   `json:"slug"`
	Markets []market `json:"markets"`
}

type opportunity struct {
	EventTitle string
	EventSlug  string
	Legs       int
	AskSum     float64
	Edge       float64 // 1 - AskSum, the gross arbitrage edge
}

func main() {
	limit := flag.Int("limit", 200, "number of events to fetch")
	minEdge := flag.Float64("min-edge", 0.02, "minimum gross edge (e.g. 0.02 = 2 cents on the dollar) to report")
	outCSV := flag.String("csv", "output/results.csv", "path to write results as CSV (empty to skip)")
	flag.Parse()

	events, err := fetchEvents(*limit)
	if err != nil {
		log.Fatal(err)
	}

	opps := findArbitrage(events, *minEdge)
	sort.Slice(opps, func(i, j int) bool { return opps[i].Edge > opps[j].Edge })

	printTable(opps)

	if *outCSV != "" {
		if err := writeCSV(*outCSV, opps); err != nil {
			log.Fatal(err)
		}
		fmt.Println("wrote", *outCSV)
	}
}

func fetchEvents(limit int) ([]event, error) {
	url := fmt.Sprintf("%s?limit=%d&active=true&closed=false&order=volume24hr&ascending=false", gammaEventsURL, limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gamma API returned status %d", resp.StatusCode)
	}
	var events []event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}
	return events, nil
}

// findArbitrage groups each event's negRisk Yes/No markets (mutually
// exclusive candidates in the same event) and flags groups whose Yes-ask
// prices sum to less than 1-minEdge.
func findArbitrage(events []event, minEdge float64) []opportunity {
	var opps []opportunity
	for _, e := range events {
		var legs []market
		for _, m := range e.Markets {
			if m.NegRisk && m.AcceptOrders && isYesNo(m.Outcomes) {
				legs = append(legs, m)
			}
		}
		if len(legs) < 2 {
			continue
		}
		sum := 0.0
		for _, m := range legs {
			sum += m.BestAsk
		}
		edge := 1 - sum
		if edge >= minEdge {
			opps = append(opps, opportunity{
				EventTitle: e.Title,
				EventSlug:  e.Slug,
				Legs:       len(legs),
				AskSum:     sum,
				Edge:       edge,
			})
		}
	}
	return opps
}

func isYesNo(outcomesJSON string) bool {
	var outcomes []string
	if err := json.Unmarshal([]byte(outcomesJSON), &outcomes); err != nil || len(outcomes) != 2 {
		return false
	}
	return outcomes[0] == "Yes" && outcomes[1] == "No"
}

func printTable(opps []opportunity) {
	if len(opps) == 0 {
		fmt.Println("no groups found above the edge threshold")
		return
	}
	fmt.Printf("%-60s %5s %8s %8s\n", "event", "legs", "ask-sum", "edge")
	for _, o := range opps {
		title := o.EventTitle
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		fmt.Printf("%-60s %5d %8.4f %8.4f\n", title, o.Legs, o.AskSum, o.Edge)
	}
}

func writeCSV(path string, opps []opportunity) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"event_title", "event_slug", "legs", "ask_sum", "edge"}); err != nil {
		return err
	}
	for _, o := range opps {
		if err := w.Write([]string{
			o.EventTitle,
			o.EventSlug,
			fmt.Sprint(o.Legs),
			fmt.Sprintf("%.4f", o.AskSum),
			fmt.Sprintf("%.4f", o.Edge),
		}); err != nil {
			return err
		}
	}
	return nil
}
