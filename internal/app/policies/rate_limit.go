package policies

import (
    "time"
)

type RateLimiter struct{
    ch chan struct{}
}

func NewRateLimiter(qps int) *RateLimiter {
    if qps <= 0 { qps = 1 }
    rl := &RateLimiter{ch: make(chan struct{}, qps)}
    go func(){
        t := time.NewTicker(time.Second / time.Duration(qps))
        for range t.C { select { case rl.ch <- struct{}{}: default: } }
    }()
    return rl
}

func (r *RateLimiter) Acquire() { <-r.ch }

