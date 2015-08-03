package utils

import (
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
)

type MultiExecError struct {
	Query    string
	Original error
}

func (m *MultiExecError) Error() string {
	return "Error in query " + m.Query + " - " + m.Original.Error()
}

func MultiExec(session *r.Session, terms ...r.Term) error {
	for _, term := range terms {
		if err := term.Exec(session); err != nil {
			return &MultiExecError{
				Query:    term.String(),
				Original: err,
			}
		}
	}

	return nil
}
