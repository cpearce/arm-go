// Copyright 2022 Nokia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package arm

import (
	"errors"
	"io"
)

var (
	ErrItemsReaderIsNil = errors.New("ItemsReader may not be nil")
	ErrRulesWriterIsNil = errors.New("RulesWriter may not be nil")
)

type (
	ItemsReader    func() (io.ReadCloser, error)
	RulesWriter    func() (io.WriteCloser, error)
	ItemsetsWriter func() (io.WriteCloser, error)
)

type ArgumentsV2 struct {
	ItemsReader    ItemsReader
	RulesWriter    RulesWriter
	ItemsetsWriter ItemsetsWriter
	MinSupport     float64
	MinConfidence  float64
	MinLift        float64
}

func (args ArgumentsV2) Validate() error {
	if args.ItemsReader == nil {
		return ErrItemsReaderIsNil
	}
	if args.RulesWriter == nil {
		return ErrRulesWriterIsNil
	}
	return Arguments{
		MinSupport:    args.MinSupport,
		MinConfidence: args.MinConfidence,
		MinLift:       args.MinLift,
	}.Validate()
}
