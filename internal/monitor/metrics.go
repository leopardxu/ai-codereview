package monitor

import "sync/atomic"

var NodeErrors uint64
var NodeCalls uint64
var ContextCalls uint64
var ContextCacheHits uint64
var ContextCacheMiss uint64
var GraphExecMillisSum uint64
var GraphExecCount uint64
var ContextFuncCount uint64
var ContextClassCount uint64
var ContextDepCount uint64

func IncError() { atomic.AddUint64(&NodeErrors, 1) }
func IncCall()  { atomic.AddUint64(&NodeCalls, 1) }
func IncContextCall() { atomic.AddUint64(&ContextCalls, 1) }
func IncContextHit()  { atomic.AddUint64(&ContextCacheHits, 1) }
func IncContextMiss() { atomic.AddUint64(&ContextCacheMiss, 1) }
func AddGraphExecMillis(ms uint64) { atomic.AddUint64(&GraphExecMillisSum, ms); atomic.AddUint64(&GraphExecCount, 1) }
func IncContextFunc() { atomic.AddUint64(&ContextFuncCount, 1) }
func IncContextClass() { atomic.AddUint64(&ContextClassCount, 1) }
func IncContextDep() { atomic.AddUint64(&ContextDepCount, 1) }
