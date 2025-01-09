package errs

const (
	MissingRequiredField = "MISSING_REQUIRED_FIELD"
	Internal = "INTERNAL"
	IncompleteAction = "INCOMPLETE_ACTION"
	PreconditionFailed = "PRECONDITION_FAILED"
	InvalidState = "INVALID_STATE"
	ObjectExists = "OBJECT_EXISTS"
	Unauthorized = "UNAUTHORIZED"

	// Postgres error codes
	UniqueViolation = "23505"
	ForeignKeyViolation = "23503"
	CheckViolation = "23514"
	NoRowsMatch = "no rows in result set"
)

type Error struct {
	Type string
	Message string
}
