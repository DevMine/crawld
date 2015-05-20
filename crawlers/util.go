// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crawlers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

// genInsQuery generates a query string for an insertion in the database.
func genInsQuery(tableName string, fields ...string) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("INSERT INTO %s(%s)\n",
		tableName, strings.Join(fields, ",")))
	buf.WriteString("VALUES(")

	for ind := range fields {
		if ind > 0 {
			buf.WriteString(",")
		}

		buf.WriteString(fmt.Sprintf("$%d", ind+1))
	}

	buf.WriteString(")\n")

	return buf.String()
}

// genUpdateQuery generates a query string for an update of fields in the
// database.
func genUpdateQuery(tableName string, id int, fields ...string) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("UPDATE %s\n", tableName))
	buf.WriteString("SET ")

	for ind, field := range fields {
		if ind > 0 {
			buf.WriteString(",")
		}

		buf.WriteString(fmt.Sprintf("%s=$%d", field, ind+1))
	}

	buf.WriteString(fmt.Sprintf("WHERE id=%d\n", id))

	return buf.String()

}

// formatTimestamp formats a github.Timestamp to a string suitable to use
// as a timestamp with timezone PostgreSQL data type
func formatTimestamp(timeStamp *github.Timestamp) string {
	timeFormat := time.RFC3339
	if timeStamp == nil {
		glog.Error("'timeStamp' arg given is nil")
		t := time.Time{}
		return t.Format(timeFormat)
	}
	return timeStamp.Format(timeFormat)
}

// isLanguageWanted checks if language(s) is in the list of wanted
// languages.
func isLanguageWanted(suppLangs []string, prjLangs interface{}) (bool, error) {
	if prjLangs == nil {
		return false, nil
	}

	switch prjLangs.(type) {
	case map[string]int:
		langs := prjLangs.(map[string]int)
		for k := range langs {
			for _, v := range suppLangs {
				if strings.EqualFold(k, v) {
					return true, nil
				}
			}
		}
	case *string:
		lang := prjLangs.(*string)
		if lang == nil {
			return false, nil
		}

		for _, sl := range suppLangs {
			if sl == *lang {
				return true, nil
			}
		}
	default:
		return false, errors.New("isLanguageSupported: invalid prjLangs type")
	}

	return false, nil
}

// verifyRepo checks all essential fields of a Repository structure for nil
// values. An error is returned if one of the essential field is nil.
func verifyRepo(repo *github.Repository) error {
	if repo == nil {
		return newInvalidStructError("verifyRepo: repo is nil")
	}

	var err *invalidStructError
	if repo.ID == nil {
		err = newInvalidStructError("verifyRepo: contains nil fields:").AddField("ID")
	} else {
		err = newInvalidStructError(fmt.Sprintf("verifyRepo: repo #%d contains nil fields: ", *repo.ID))
	}

	if repo.Name == nil {
		err.AddField("Name")
	}

	if repo.Language == nil {
		err.AddField("Language")
	}

	if repo.CloneURL == nil {
		err.AddField("CloneURL")
	}

	if repo.Owner == nil {
		err.AddField("Owner")
	} else {
		if repo.Owner.Login == nil {
			err.AddField("Owner.Login")
		}
	}

	if repo.Fork == nil {
		err.AddField("Fork")
	}

	if err.FieldsLen() > 0 {
		return err
	}

	return nil
}
