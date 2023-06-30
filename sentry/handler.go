package sentry

import "github.com/getsentry/sentry-go"

func SentryHandler(err error) {
	sentry.CaptureException(err)
}
