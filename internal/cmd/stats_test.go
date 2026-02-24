// Copyright 2025 Erst Users
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"
	"github.com/dotandev/hintents/internal/simulator"
)

// TestBuildContractStats verifies that the aggregation and cost weighting
// logic correctly identifies the most expensive calls.
func TestBuildContractStats(t *testing.T) {
	cidA := "CONTRACT_A_EX"
	cidB := "CONTRACT_B_CH"

	// Simulate a response with specific events
	resp := &simulator.SimulationResponse{
		CategorizedEvents: []simulator.CategorizedEvent{
			// Contract A: 1 write + 1 auth = 5 cost, 2 depth
			{ContractID: &cidA, EventType: "storage_write"},
			{ContractID: &cidA, EventType: "require_auth"},
			// Contract B: 1 default = 1 cost, 1 depth
			{ContractID: &cidB, EventType: "contract_call"},
		},
	}

	stats := buildContractStats(resp)

	if len(stats) != 2 {
		t.Fatalf("expected 2 contract stats, got %d", len(stats))
	}

	// Verify Contract A (Should be index 0 as it is more expensive)
	if stats[0].contractID != cidA {
		t.Errorf("expected %s to be the most expensive, got %s", cidA, stats[0].contractID)
	}
	if stats[0].estimatedCost != 5 {
		t.Errorf("expected cost 5 for A, got %d", stats[0].estimatedCost)
	}
	if stats[0].callDepth != 2 {
		t.Errorf("expected depth 2 for A, got %d", stats[0].callDepth)
	}

	// Verify Contract B
	if stats[1].estimatedCost != 1 {
		t.Errorf("expected cost 1 for B, got %d", stats[1].estimatedCost)
	}
}

// TestBuildContractStats_Empty ensures the function handles empty data gracefully
func TestBuildContractStats_Empty(t *testing.T) {
	resp := &simulator.SimulationResponse{
		CategorizedEvents: []simulator.CategorizedEvent{},
	}

	stats := buildContractStats(resp)

	if len(stats) != 0 {
		t.Errorf("expected 0 stats for empty input, got %d", len(stats))
	}
}