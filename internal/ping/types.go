package ping

type QueryResult string

const (
	QueryResultPass QueryResult = "PASS"
	QueryResultFail QueryResult = "FAIL"
)

type QueryRecord struct {
	Endpoint       string      `dynamodbav:"Endpoint" json:"endpoint,omitempty"`
	ExpirationTime int64       `dynamodbav:"ExpirationTime" json:"expirationTime,omitempty"`
	Result         QueryResult `dynamodbav:"Result" json:"result,omitempty"`
	Timestamp      int64       `dynamodbav:"Timestamp" json:"timestamp,omitempty"`
}

type Request struct {
	Endpoints []string `json:"endpoints"`
}
