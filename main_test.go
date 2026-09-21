package main

import "testing"

func TestFindArbitrage(t *testing.T) {
	events := []event{
		{
			Title: "underpriced group",
			Markets: []market{
				{Outcomes: `["Yes","No"]`, NegRisk: true, AcceptOrders: true, BestAsk: 0.30},
				{Outcomes: `["Yes","No"]`, NegRisk: true, AcceptOrders: true, BestAsk: 0.40},
				{Outcomes: `["Yes","No"]`, NegRisk: true, AcceptOrders: true, BestAsk: 0.20}, // sum 0.90 -> edge 0.10
			},
		},
		{
			Title: "efficiently priced group",
			Markets: []market{
				{Outcomes: `["Yes","No"]`, NegRisk: true, AcceptOrders: true, BestAsk: 0.50},
				{Outcomes: `["Yes","No"]`, NegRisk: true, AcceptOrders: true, BestAsk: 0.51}, // sum 1.01 -> no edge
			},
		},
		{
			Title: "not a negRisk group",
			Markets: []market{
				{Outcomes: `["Giants","Rams"]`, NegRisk: false, AcceptOrders: true, BestAsk: 0.30},
				{Outcomes: `["Giants","Rams"]`, NegRisk: false, AcceptOrders: true, BestAsk: 0.30},
			},
		},
		{
			Title: "single leg, no group",
			Markets: []market{
				{Outcomes: `["Yes","No"]`, NegRisk: true, AcceptOrders: true, BestAsk: 0.10},
			},
		},
	}

	opps := findArbitrage(events, 0.02)
	if len(opps) != 1 {
		t.Fatalf("expected 1 opportunity, got %d: %+v", len(opps), opps)
	}
	got := opps[0]
	if got.EventTitle != "underpriced group" {
		t.Errorf("wrong event flagged: %s", got.EventTitle)
	}
	if got.Legs != 3 {
		t.Errorf("expected 3 legs, got %d", got.Legs)
	}
	want := 0.10
	if diff := got.Edge - want; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("edge = %v, want %v", got.Edge, want)
	}
}
