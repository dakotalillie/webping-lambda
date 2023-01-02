package ping

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

var httpClient = http.Client{Timeout: 10 * time.Second}

func SendRequestToEndpoint(ctx context.Context, endpoint string) QueryRecord {
	log.Println("sending request to", endpoint)
	now := time.Now()
	expiration := now.Add(24 * time.Hour)
	record := QueryRecord{Endpoint: endpoint, ExpirationTime: expiration.Unix(), Timestamp: now.Unix()}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("received error while creating request for %s: %s\n", endpoint, err)
		record.Result = QueryResultFail
		return record
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("received error while sending request to %s: %s\n", endpoint, err)
		record.Result = QueryResultFail
	} else if resp.StatusCode >= 400 {
		log.Printf("received error status code from request to %s: %d\n", endpoint, resp.StatusCode)
		record.Result = QueryResultFail
	} else if err = resp.Body.Close(); err != nil {
		log.Printf("failed to close response body from request to %s: %s\n", endpoint, err)
		record.Result = QueryResultFail
	} else {
		log.Println("received successful response from", endpoint)
		record.Result = QueryResultPass
	}

	return record
}

func SendRequestsToAllEndpoints(ctx context.Context, endpoints []string) []QueryRecord {
	var wg sync.WaitGroup
	wg.Add(len(endpoints))
	results := make([]QueryRecord, len(endpoints))

	for i, endpoint := range endpoints {
		go func(i int, endpoint string) {
			defer wg.Done()
			results[i] = SendRequestToEndpoint(ctx, endpoint)
		}(i, endpoint)
	}
	wg.Wait()

	return results
}
