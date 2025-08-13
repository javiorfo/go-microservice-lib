package tracing

import (
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

func Log(span trace.Span) string {
	if span.SpanContext().HasTraceID() && span.SpanContext().HasSpanID() {
		return fmt.Sprintf("[traceID: %s, spanID: %s] ", span.SpanContext().TraceID(), span.SpanContext().SpanID())
	}
	return ""
}

func Name() string {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	lastPeriodIndex := strings.LastIndex(funcName, ".")
	s, _ := strings.CutPrefix(funcName[lastPeriodIndex+1:], "New")
	return s
}
