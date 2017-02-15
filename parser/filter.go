// Copyright 2017 wgliang. All rights reserved.
// Use of this source code is governed by Apache
// license that can be found in the LICENSE file.

// Package parser provides filtering rules if you need.
package parser

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
)

// Callback function from proxy to postgresql for rewrite
// request or sql.
type Callback func(get []byte) bool

// Extracte sql statement from string
func Extracte(str []byte) string {
	return string(str)[5:]
}

// ReWrite SQL test
func ReWriteSQL(str []byte) []byte {
	return append(str[0:5], []byte(strings.Replace(Extracte(str), "20", "10", -1))...)
}

// GetQueryModificada calllback
func GetQueryModificada(queryOriginal string) string {
	if queryOriginal[:5] != "power" {

		return queryOriginal
	}
	return "select * from clientes limit 1;"
}

func Filter(str []byte) bool {
	sql := Extracte(str)
	tree, err := Parse(sql)
	if err != nil {
		glog.Errorln(err)
		return false
	}

	switch tree.(type) {
	case *Select:
		return ParseSelect(tree.(*Select))
	case *Delete:
		return ParseDelete(tree.(*Delete))
	case *Insert:
		return ParseInsert(tree.(*Insert))
	case *Update:
		return ParseUpdate(tree.(*Update))
	}
	return false
}

func Return(str []byte) bool {
	fmt.Println(string(str))
	return true
}

func ParseSelect(sql *Select) bool {
	return !Is_SELECT_ALL(sql) && !Is_ORDER_BY_RAND(sql)
}

func Is_SELECT_ALL(sql *Select) bool {
	buf := NewTrackedBuffer(nil)
	sql.SelectExprs.Format(buf)
	if "*" == buf.String() {
		return true
	}
	return false
}

func Is_ORDER_BY_RAND(sql *Select) bool {
	buf := NewTrackedBuffer(nil)
	sql.OrderBy.Format(buf)
	if "rand()" == strings.ToLower(buf.String()) {
		return true
	}
	return false
}

func ParseDelete(sql *Delete) bool {
	return !Is_BIG_DELETE(sql)
}

func Is_BIG_DELETE(sql *Delete) bool {
	buf := NewTrackedBuffer(nil)
	sql.Limit.Format(buf)
	if "1000" < buf.String() {
		return true
	}
	return false
}

func ParseInsert(sql *Insert) bool {
	return !Is_BIG_INSERT(sql)
}

func Is_BIG_INSERT(sql *Insert) bool {
	buf := NewTrackedBuffer(nil)
	sql.Rows.Format(buf)
	if "1000" < buf.String() {
		return true
	}
	return false
}

func ParseUpdate(sql *Update) bool {
	return true
}
