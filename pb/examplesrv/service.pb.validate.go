// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: pb/examplesrv/service.proto

package examplesrv

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/ptypes"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = ptypes.DynamicAny{}
)

// define the regex for a UUID once up-front
var _service_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on Req with the rules defined in the proto
// definition for this message. If any rules are violated, an error is returned.
func (m *Req) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Str

	return nil
}

// ReqValidationError is the validation error returned by Req.Validate if the
// designated constraints aren't met.
type ReqValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ReqValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ReqValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ReqValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ReqValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ReqValidationError) ErrorName() string { return "ReqValidationError" }

// Error satisfies the builtin error interface
func (e ReqValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sReq.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ReqValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ReqValidationError{}

// Validate checks the field values on Resp with the rules defined in the proto
// definition for this message. If any rules are violated, an error is returned.
func (m *Resp) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Str

	return nil
}

// RespValidationError is the validation error returned by Resp.Validate if the
// designated constraints aren't met.
type RespValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RespValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RespValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RespValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RespValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RespValidationError) ErrorName() string { return "RespValidationError" }

// Error satisfies the builtin error interface
func (e RespValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sResp.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RespValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RespValidationError{}
