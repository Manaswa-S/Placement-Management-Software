package errs

const (
	MissingRequiredField = "MISSING_REQUIRED_FIELD"
	Internal = "INTERNAL"
	IncompleteAction = "INCOMPLETE_ACTION"
	PreconditionFailed = "PRECONDITION_FAILED"
	InvalidState = "INVALID_STATE"
	ObjectExists = "OBJECT_EXISTS"
	Unauthorized = "UNAUTHORIZED"
	NotFound = "NOT_FOUND"
	InvalidFormat = "INVALID_FORMAT"
	IncompleteForm = "INCOMPLETE_FORM"

	// Postgres error codes (SQLSTATE)
	UniqueViolation = "23505"
	ForeignKeyViolation = "23503"
	CheckViolation = "23514"
	NoRowsMatch = "no rows in result set"
)

type Error struct {
	Type string // error type, used from errs.{Type}
	Message string // the actual error message
	ToRespondWith bool // send the error message directly to user if true
}
