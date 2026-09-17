package main

import "strings"

type PaymentEvent struct {
	ID       string            `json:"id"`
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata"`
}

func riskAction(metadata map[string]string) string {
	if strings.EqualFold(metadata["risk"], "high") || strings.EqualFold(metadata["status"], "reversed") {
		return "review"
	}
	return "allow"
}
