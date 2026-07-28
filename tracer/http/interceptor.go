package http

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func ExtractResponseSpan(ctx context.Context, r *http.Response) context.Context {
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

	return ctx
}

func InjectRequestSpan(ctx context.Context, r *http.Request) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))
}

func ExtractRequestSpan(r *http.Request) *http.Request {
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	r = r.WithContext(ctx)

	return r
}
