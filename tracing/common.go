package tracing

import (
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func LogError(span trace.Span, msg string) string {
	return log(span, msg, true)
}

func LogInfo(span trace.Span, msg string) string {
	return log(span, msg, false)
}

func log(span trace.Span, msg string, isError bool) string {
	if span.SpanContext().IsValid() {
		if isError {
			span.SetStatus(codes.Error, msg)
		} else {
			span.AddEvent(msg)
		}
		return fmt.Sprintf("[traceID: %s, spanID: %s] %s", span.SpanContext().TraceID(), span.SpanContext().SpanID(), msg)
	}
	return msg
}

func Name() string {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	lastPeriodIndex := strings.LastIndex(funcName, ".")
	s, _ := strings.CutPrefix(funcName[lastPeriodIndex+1:], "New")
	return s
}
