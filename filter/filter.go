// Copyright 2017 wgliang. All rights reserved.
// Use of this source code is governed by Apache
// license that can be found in the LICENSE file.

// Package filter provides filtering rules if you need.
package filter

import (
	"strings"
)

// Callback function from proxy to postgresql for rewrite
// request or sql.
type Callback func(get string) string

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
