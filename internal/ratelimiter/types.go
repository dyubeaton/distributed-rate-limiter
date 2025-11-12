package ratelimiter

type RateLimitRequest struct {
	/*
		use backtick tags to tell encoding/json how to convert between the Go structs and JSON
		without this it would convert to something like {"Identifier": "user:123", "Tokens": 5} which breaks lowercase JSON convention
	*/
	Identifier string `json:"identifier"` //tag syntax seems to be `json"<name>"`
	Tokens     string `json:"tokens"`     //number of token required in request
	Algorithm  string `json:"algorithm"`  //which rate limiting algo: Supported: "token_bucket", "fixed_window", "sliding_window"
}

type RateLimitResponse struct {
	Allowed   bool `json:"allowed"`   //if requested should be permitted
	Remaining int  `json:"remaining"` //remaining tokens
	Limit     int  `json:"limit"`     //token limit
	//unix timestamps and durations are conventially int64
	ResetAfter int64 `json:"reset_after"` //next time bucket wil be full (seconds)
	RetryAfter int64 `json:"retry_after"` //next time to try if limit exceeded (seconds, only if !Allowed)
}

type Config struct {

	//bucket limit
	Capacity int

	//tokens added per/s
	RefillRate float64

	Algorithm string
}
