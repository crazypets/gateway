package tgbot

import (
	"context"
	"github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	"time"
)

type (
	sdk interface {
		Send(ctx context.Context, t time.Time, lvl string, msg string) error
	}
	TracedSDK struct {
		sdk    sdk
		tracer opentracing.Tracer
		opName string
	}
)

func WithTracer(sdk sdk, tracer opentracing.Tracer, operationName string) *TracedSDK {
	return &TracedSDK{sdk, tracer, operationName}
}

func (t *TracedSDK) Send(ctx context.Context, time time.Time, lvl string, msg string) error {
	clientSpan := opentracing.SpanFromContext(ctx)
	if clientSpan == nil {
		clientSpan = t.tracer.StartSpan(t.opName)
	} else {
		clientSpan.SetOperationName(t.opName)
	}
	defer clientSpan.Finish()

	otext.SpanKindRPCClient.Set(clientSpan)
	ctx = opentracing.ContextWithSpan(ctx, clientSpan)

	return t.sdk.Send(ctx, time, lvl, msg)
}
