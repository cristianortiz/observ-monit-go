package tracing

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Business domain attributes
var (
	// User attributes
	UserIDKey    = attribute.Key("user.id")
	UserEmailKey = attribute.Key("user.email")
	UserNameKey  = attribute.Key("user.name")

	// Operation attributes
	OperationKey     = attribute.Key("operation")
	OperationTypeKey = attribute.Key("operation.type")
	EntityKey        = attribute.Key("entity")

	// Database attributes (additional to otelpgx)
	DBOperationKey    = attribute.Key("db.operation")
	DBRowsAffectedKey = attribute.Key("db.rows_affected")
	DBQueryTimeKey    = attribute.Key("db.query_time_ms")
)

// HTTP attributes (beyond standard otelfiber)
var (
	EndpointKey     = attribute.Key("endpoint")
	RequestSizeKey  = attribute.Key("http.request.size")
	ResponseSizeKey = attribute.Key("http.response.size")
	UserAgentKey    = attribute.Key("http.user_agent")
)

// Business operation types
const (
	OperationTypeCreate = "create"
	OperationTypeRead   = "read"
	OperationTypeUpdate = "update"
	OperationTypeDelete = "delete"
	OperationTypeList   = "list"
	OperationTypeCount  = "count"
)

// Entity types
const (
	EntityUser    = "user"
	EntityProduct = "product"
	EntityOrder   = "order"
)

// SetUserAttributes sets user-related attributes on a span
func SetUserAttributes(span trace.Span, userID, email, name string) {
	if userID != "" {
		span.SetAttributes(UserIDKey.String(userID))
	}
	if email != "" {
		span.SetAttributes(UserEmailKey.String(email))
	}
	if name != "" {
		span.SetAttributes(UserNameKey.String(name))
	}
}

// SetOperationAttributes sets operation-related attributes on a span
func SetOperationAttributes(span trace.Span, operation, operationType, entity string) {
	span.SetAttributes(
		OperationKey.String(operation),
		OperationTypeKey.String(operationType),
		EntityKey.String(entity),
	)
}

// SetDBOperationAttributes sets database operation attributes
func SetDBOperationAttributes(span trace.Span, operation string, rowsAffected int64) {
	span.SetAttributes(
		DBOperationKey.String(operation),
		DBRowsAffectedKey.Int64(rowsAffected),
	)
}

// SetHTTPAttributes sets HTTP-related attributes on a span
func SetHTTPAttributes(span trace.Span, method, path, userAgent string, reqSize, respSize int) {
	span.SetAttributes(
		attribute.String("http.method", method),
		EndpointKey.String(path),
	)
	if userAgent != "" {
		span.SetAttributes(UserAgentKey.String(userAgent))
	}
	if reqSize > 0 {
		span.SetAttributes(RequestSizeKey.Int(reqSize))
	}
	if respSize > 0 {
		span.SetAttributes(ResponseSizeKey.Int(respSize))
	}
}

// RecordError records an error on a span with appropriate status
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// RecordSuccess sets span status to OK
func RecordSuccess(span trace.Span) {
	span.SetStatus(codes.Ok, "")
}
