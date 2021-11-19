/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package ex

// As is a helper method that returns an error as an ex.
func As(err interface{}) *Ex {
	if typed, typedOk := err.(Ex); typedOk {
		return &typed
	}
	if typed, typedOk := err.(*Ex); typedOk {
		return typed
	}
	return nil
}
