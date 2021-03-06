// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorStrings(t *testing.T) {
	err := newTransportError(ConnectionClose, fmt.Errorf("foo"))
	assert.Equal(t, "connection close: foo", err.Error())

	err = newTransportError(DialError, fmt.Errorf("foo"))
	assert.Equal(t, "dial error: foo", err.Error())

	err = newTransportError(LaunchError, fmt.Errorf("foo"))
	assert.Equal(t, "launch error: foo", err.Error())

	err = newTransportError(EncodeError, fmt.Errorf("foo"))
	assert.Equal(t, "encode error: foo", err.Error())

	err = newTransportError(DecodeError, fmt.Errorf("foo"))
	assert.Equal(t, "decode error: foo", err.Error())

	err = newTransportError(DetectionError, fmt.Errorf("foo"))
	assert.Equal(t, "detection error: foo", err.Error())

	err = newTransportError(NetworkError, fmt.Errorf("foo"))
	assert.Equal(t, "network error: foo", err.Error())

	err = newTransportError(0, fmt.Errorf("foo"))
	assert.Equal(t, "unknown error: foo", err.Error())
}
