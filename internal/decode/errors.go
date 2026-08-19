package decode

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	CodeKGGDatabaseRequired     ErrorCode = "KGG_DATABASE_REQUIRED"
	CodeKGGDatabaseProvider     ErrorCode = "KGG_DATABASE_PROVIDER_FAILED"
	CodeKGGDatabaseInvalid      ErrorCode = "KGG_DATABASE_INVALID"
	CodeKGGDatabaseEntryMissing ErrorCode = "KGG_DATABASE_ENTRY_NOT_FOUND"
	CodeQMCExternalKeyRequired  ErrorCode = "QMC_EXTERNAL_KEY_REQUIRED"
	CodeQMCExternalKeyProvider  ErrorCode = "QMC_EXTERNAL_KEY_PROVIDER_FAILED"
	CodeUnsupportedEncrypted    ErrorCode = "UNSUPPORTED_ENCRYPTED_FORMAT"
	CodeUnknownFormat           ErrorCode = "UNKNOWN_FORMAT"
	CodeDecodedAudioUnknown     ErrorCode = "DECODED_AUDIO_FORMAT_UNKNOWN"
	CodeOutputExists            ErrorCode = "OUTPUT_EXISTS"
	CodeOutputConflictsSource   ErrorCode = "OUTPUT_CONFLICTS_SOURCE"
	CodeInvalidInput            ErrorCode = "INVALID_INPUT"
)

type Error struct {
	Code ErrorCode
	Err  error
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %v", e.Code, e.Err) }
func (e *Error) Unwrap() error { return e.Err }

func coded(code ErrorCode, err error) error { return &Error{Code: code, Err: err} }

func CodeOf(err error) ErrorCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}
