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

package arm_test

import (
	"testing"

	"github.com/nokia/arm-go"
)

func TestArgumentsValidate(t *testing.T) {
	tests := []struct {
		name    string
		args    arm.Arguments
		wantErr error
	}{
		{"ok", arm.Arguments{}, nil},
		{"minsupport<0", arm.Arguments{MinSupport: -0.1}, arm.ErrMinSupportOutOfRange},
		{"minsupport=0", arm.Arguments{MinSupport: 0.0}, nil},
		{"minsupport=1", arm.Arguments{MinSupport: 1.0}, nil},
		{"minsupport>1", arm.Arguments{MinSupport: 1.1}, arm.ErrMinSupportOutOfRange},
		{"minconfidence<0", arm.Arguments{MinConfidence: -0.1}, arm.ErrMinConfidenceOutOfRange},
		{"minconfidence=0", arm.Arguments{MinConfidence: 0.0}, nil},
		{"minconfidence=1", arm.Arguments{MinConfidence: 1.0}, nil},
		{"minconfidence>1", arm.Arguments{MinConfidence: 1.1}, arm.ErrMinConfidenceOutOfRange},
		{"minlift<0", arm.Arguments{MinLift: 0.1}, arm.ErrMinLiftOutOfRange},
		{"minlift=0", arm.Arguments{MinLift: 0.0}, nil},
		{"minconfidence<1", arm.Arguments{MinLift: 0.9}, arm.ErrMinLiftOutOfRange},
		{"minconfidence=1", arm.Arguments{MinLift: 1.0}, nil},
		{"minconfidence>1", arm.Arguments{MinLift: 1.1}, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.args.Validate(); err != tt.wantErr {
				t.Logf("expected error %s, got error %s", tt.wantErr, err)
				t.Fail()
			}
		})
	}
}
