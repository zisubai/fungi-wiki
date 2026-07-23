package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type evaluationCase struct {
	ID, Query     string
	ExpectedSlugs []string `json:"expectedSlugs"`
}
type searchResponse struct {
	Items []struct {
		Slug string `json:"slug"`
	} `json:"items"`
	SemanticEnabled bool `json:"semanticEnabled"`
}

func main() {
	api := flag.String("api", "http://localhost:8080", "API base URL")
	dataset := flag.String("dataset", "testdata/search-evaluation.json", "evaluation JSON")
	k := flag.Int("k", 5, "recall cutoff")
	requireSemantic := flag.Bool("require-semantic", false, "fail when semantic search is disabled")
	flag.Parse()
	data, err := os.ReadFile(*dataset)
	must(err)
	var cases []evaluationCase
	must(json.Unmarshal(data, &cases))
	client := http.Client{Timeout: 20 * time.Second}
	hits := 0
	mrr := 0.0
	semanticSeen := false
	for _, test := range cases {
		response, err := client.Get(*api + "/api/search?q=" + url.QueryEscape(test.Query) + fmt.Sprintf("&limit=%d", *k))
		must(err)
		if response.StatusCode != 200 {
			fmt.Fprintf(os.Stderr, "%s: HTTP %s\n", test.ID, response.Status)
			os.Exit(1)
		}
		var result searchResponse
		must(json.NewDecoder(response.Body).Decode(&result))
		response.Body.Close()
		semanticSeen = semanticSeen || result.SemanticEnabled
		rank := 0
		for i, item := range result.Items {
			for _, expected := range test.ExpectedSlugs {
				if item.Slug == expected {
					rank = i + 1
					break
				}
			}
			if rank > 0 {
				break
			}
		}
		if rank > 0 {
			hits++
			mrr += 1 / float64(rank)
			fmt.Printf("PASS %-38s rank=%d\n", test.ID, rank)
		} else {
			fmt.Printf("FAIL %-38s expected=%v\n", test.ID, test.ExpectedSlugs)
		}
	}
	recall := float64(hits) / float64(len(cases))
	mrr /= float64(len(cases))
	fmt.Printf("Recall@%d=%.3f MRR=%.3f semantic=%t cases=%d\n", *k, recall, mrr, semanticSeen, len(cases))
	if *requireSemantic && !semanticSeen {
		fmt.Fprintln(os.Stderr, "semantic search was required but disabled")
		os.Exit(1)
	}
	if recall < .8 {
		os.Exit(1)
	}
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
