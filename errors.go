package resilience

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// ErrInvalidComposition identifies an executor policy configuration error.
	ErrInvalidComposition = errors.New("resilience: invalid composition")
	// ErrInvalidMetadata identifies missing or unbounded execution metadata.
	ErrInvalidMetadata = errors.New("resilience: invalid metadata")
	// ErrInvalidAttempt identifies inconsistent physical-attempt metadata.
	ErrInvalidAttempt = errors.New("resilience: invalid attempt")
	// ErrNilOperation identifies an execution without an operation.
	ErrNilOperation = errors.New("resilience: nil operation")
	// ErrLocalRejection identifies local policy denial without downstream work.
	ErrLocalRejection = errors.New("resilience: local rejection")
	// ErrIgnored identifies work deliberately omitted by an owning policy.
	ErrIgnored = errors.New("resilience: ignored")
	// ErrPolicyFailure identifies failure in policy logic rather than the operation.
	ErrPolicyFailure = errors.New("resilience: policy failure")
)

// ConfigurationError identifies a configuration field and reason. Package-created
// values use fixed safe labels; caller-constructed values are not sanitized.
type ConfigurationError struct {
	// Kind is the sentinel category exposed by Unwrap.
	Kind error
	// Field identifies the rejected configuration input.
	Field string
	// Reason explains why that input was rejected.
	Reason string
}

// Error formats Kind, Field and Reason without sanitizing caller-built values.
func (err *ConfigurationError) Error() string {
	return fmt.Sprintf("%v: %s: %s", err.Kind, err.Field, err.Reason)
}

// Unwrap returns the configuration category in Kind.
func (err *ConfigurationError) Unwrap() error { return err.Kind }

func invalid(kind error, field, reason string) error {
	return &ConfigurationError{Kind: kind, Field: field, Reason: reason}
}

// LocalRejectionError reports local denial. LocalRejection bounds Policy and
// Reason, but retains Cause unchanged. Error omits the cause's text; this is not
// general redaction of arbitrary fields or formatting of caller-owned values.
type LocalRejectionError struct {
	// Policy is the denying policy's diagnostic identity.
	Policy PolicyID
	// Reason identifies the local admission rule.
	Reason string
	// Cause is optional caller-owned detail exposed through Unwrap, not Error.
	Cause error
}

// Error formats the local category, Policy and Reason without calling Cause.Error.
func (err *LocalRejectionError) Error() string {
	return fmt.Sprintf("%v: %s: %s", ErrLocalRejection, err.Policy, err.Reason)
}

// Unwrap returns ErrLocalRejection and the optional original Cause.
func (err *LocalRejectionError) Unwrap() []error {
	if err.Cause == nil {
		return []error{ErrLocalRejection}
	}
	return []error{ErrLocalRejection, err.Cause}
}

// IgnoredError describes omitted work; Ignored bounds and sanitizes its Reason.
type IgnoredError struct {
	// Reason describes omitted work; Ignored sanitizes it to bounded printable text.
	Reason string
}

// Error formats ErrIgnored and Reason without additional sanitization.
func (err *IgnoredError) Error() string { return fmt.Sprintf("%v: %s", ErrIgnored, err.Reason) }

// Unwrap returns ErrIgnored.
func (err *IgnoredError) Unwrap() error { return ErrIgnored }

// PolicyExecutionError identifies a failed policy mechanism. PolicyFailure
// sanitizes Policy and Stage, but preserves Cause unchanged and exposes it
// through Unwrap. Error omits cause text, not arbitrary caller-provided fields.
type PolicyExecutionError struct {
	// Policy identifies the failing policy.
	Policy PolicyID
	// Stage identifies the failed policy mechanism.
	Stage string
	// Cause is optional caller-owned detail exposed through Unwrap, not Error.
	Cause error
}

// Error formats the policy category, Policy and Stage without calling Cause.Error.
func (err *PolicyExecutionError) Error() string {
	return fmt.Sprintf("%v: %s: %s", ErrPolicyFailure, err.Policy, err.Stage)
}

// Unwrap returns ErrPolicyFailure and the optional original Cause.
func (err *PolicyExecutionError) Unwrap() []error {
	if err.Cause == nil {
		return []error{ErrPolicyFailure}
	}
	return []error{ErrPolicyFailure, err.Cause}
}

// Keep a valid printable prefix within the byte cap; do not skip an oversized
// rune to backfill later text. TestIgnoredReasonPreservesSanitizedPrefix protects
// this distinction from a character-count or filtering implementation.
func bounded(value string) string {
	candidate := value[:min(len(value), MaxIdentityLength)]
	if utf8.ValidString(candidate) && printable(candidate) {
		return candidate
	}
	var safe strings.Builder
	safe.Grow(len(candidate))
	for len(value) > 0 && safe.Len() < MaxIdentityLength {
		character, size := utf8.DecodeRuneInString(value)
		value = value[size:]
		if character == utf8.RuneError && size == 1 || !unicode.IsPrint(character) {
			character = '_'
		}
		encoded := utf8.RuneLen(character)
		if encoded < 0 || safe.Len()+encoded > MaxIdentityLength {
			break
		}
		safe.WriteRune(character)
	}
	return safe.String()
}

func printable(value string) bool {
	for _, character := range value {
		if !unicode.IsPrint(character) {
			return false
		}
	}
	return true
}
