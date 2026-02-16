package state

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	defaultLogLevel  = "info"
	defaultLogFormat = "json"
)

// sanitizeLogFormat converts format name so it matches other components
func sanitizeLogFormat(format string) string {
	switch format {
	case "console":
		return "text"
	default:
		return format
	}
}

func sFnLoggingConfiguration(_ context.Context, _ *reconciler, s *systemState) (stateFn, *ctrl.Result, error) {
	logLevel := defaultLogLevel
	logFormat := defaultLogFormat
	accessLogDisabled := false

	if s.instance.Spec.Logging != nil {
		if s.instance.Spec.Logging.Level != "" {
			logLevel = s.instance.Spec.Logging.Level
		}
		if s.instance.Spec.Logging.Format != "" {
			logFormat = sanitizeLogFormat(s.instance.Spec.Logging.Format)
		}
		accessLogDisabled = s.instance.Spec.Logging.AccessLogDisabled
	}
	s.flagsBuilder.WithLogging(logLevel, logFormat, accessLogDisabled)

	return nextState(sFnStorageConfiguration)
}
