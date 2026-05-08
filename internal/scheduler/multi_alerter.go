package scheduler

// MultiAlerter fans out alert notifications to multiple Alerter implementations.
// It collects all errors and returns a combined error if any alerter fails.
type MultiAlerter struct {
	alerters []Alerter
}

// NewMultiAlerter creates a MultiAlerter that dispatches to all provided alerters.
func NewMultiAlerter(alerters ...Alerter) *MultiAlerter {
	return &MultiAlerter{alerters: alerters}
}

// Alert sends the result to every registered alerter.
// It continues even if one fails, collecting all errors.
func (m *MultiAlerter) Alert(result Result) error {
	var errs []string
	for _, a := range m.alerters {
		if err := a.Alert(result); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return &multiAlertError{errs: errs}
	}
	return nil
}

// multiAlertError aggregates errors from multiple alerters.
type multiAlertError struct {
	errs []string
}

func (e *multiAlertError) Error() string {
	if len(e.errs) == 1 {
		return e.errs[0]
	}
	out := "multiple alert errors:"
	for _, s := range e.errs {
		out += " [" + s + "]"
	}
	return out
}
