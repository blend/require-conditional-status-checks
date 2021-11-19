// Copyright 2021 Blend Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package actions

import (
	"reflect"
	"unsafe"

	githubactions "github.com/sethvargo/go-githubactions"
)

// actionGetEnv uses reflection to access the unexported `getenv` field in
// an action and invoke it. This is to help make it easier to write unit tests
// for code using actions. By doing this, the environment need only be mocked
// via `githubactions.New(githubactions.WithGetenv(...))` instead of patching
// both the action and the real OS environment.
func actionGetEnv(action *githubactions.Action, k string) string {
	getenvUnexported := reflect.ValueOf(action).Elem().FieldByName("getenv")
	// NOTE: Need a new `reflect.Value{}` since calling `getenvUnexported.Interface()` panics
	//       with `cannot return value obtained from unexported field or method`.
	getenvValue := reflect.NewAt(getenvUnexported.Type(), unsafe.Pointer(getenvUnexported.UnsafeAddr())).Elem()
	// NOTE: This can **technically** panic, but the type of `getenv` is well-known
	//       provided the `NewAt()` pointer manipulation was successful.
	getenv := getenvValue.Interface().(githubactions.GetenvFunc)
	return getenv(k)
}
