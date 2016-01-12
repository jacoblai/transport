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
	"testing"

	"github.com/gomqtt/packet"
	"github.com/stretchr/testify/require"
	"github.com/gorilla/websocket"
)

type abstractTestPreparer func (Handler) (Conn, chan struct{})

func abstractConnConnectTest(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		pkt, err := conn1.Receive()
		require.Equal(t, pkt.Type(), packet.CONNECT)
		require.NoError(t, err)

		err = conn1.Send(packet.NewConnackPacket())
		require.NoError(t, err)

		pkt, err = conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, ExpectedClose, toError(err).Code())
	})

	err := conn2.Send(packet.NewConnectPacket())
	require.NoError(t, err)

	pkt, err := conn2.Receive()
	require.Equal(t, pkt.Type(), packet.CONNACK)
	require.NoError(t, err)

	err = conn2.Close()
	require.NoError(t, err)

	<-done
}

func abstractConnCloseTest(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		err := conn1.Close()
		require.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	require.Nil(t, pkt)
	require.Equal(t, ExpectedClose, toError(err).Code())

	<-done
}

func abstractConnEncodeErrorTest(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		pkt := packet.NewConnackPacket()
		pkt.ReturnCode = 11 // <- invalid return code

		err := conn1.Send(pkt)
		require.Equal(t, EncodeError, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	require.Nil(t, pkt)
	require.Equal(t, ExpectedClose, toError(err).Code())

	<-done
}

func abstractConnDecodeError1Test(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		buf := []byte{0x00, 0x00} // <- too small

		if netConn, ok := conn1.(*NetConn); ok {
			netConn.conn.Write(buf)
		} else if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.conn.WriteMessage(websocket.BinaryMessage, buf)
		}

		pkt, err := conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, ExpectedClose, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	require.Nil(t, pkt)
	require.Equal(t, DetectionError, toError(err).Code())

	<-done
}

func abstractConnDecodeError2Test(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		buf := []byte{0x10, 0xff, 0xff, 0xff, 0x80} // <- too long

		if netConn, ok := conn1.(*NetConn); ok {
			netConn.conn.Write(buf)
		} else if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.conn.WriteMessage(websocket.BinaryMessage, buf)
		}

		pkt, err := conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, ExpectedClose, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	require.Nil(t, pkt)
	require.Equal(t, DetectionError, toError(err).Code())

	<-done
}

func abstractConnDecodeError3Test(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		buf := []byte{0x20, 0x02, 0x00, 0x06} // <- invalid packet

		if netConn, ok := conn1.(*NetConn); ok {
			netConn.conn.Write(buf)
		} else if webSocketConn, ok := conn1.(*WebSocketConn); ok {
			webSocketConn.conn.WriteMessage(websocket.BinaryMessage, buf)
		}

		pkt, err := conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, ExpectedClose, toError(err).Code())
	})

	pkt, err := conn2.Receive()
	require.Nil(t, pkt)
	require.Equal(t, DecodeError, toError(err).Code())

	<-done
}

func abstractConnSendAfterCloseTest(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		err := conn1.Close()
		require.NoError(t, err)
	})

	pkt, err := conn2.Receive()
	require.Nil(t, pkt)
	require.Equal(t, ExpectedClose, toError(err).Code())

	err = conn2.Send(packet.NewConnectPacket())
	require.Equal(t, ConnectionError, toError(err).Code())

	<-done
}

func abstractConnCountersTest(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		pkt, _ := conn1.Receive()
		require.Equal(t, int64(pkt.Len()), conn1.BytesRead())

		pkt2 := packet.NewConnackPacket()
		conn1.Send(pkt2)
		require.Equal(t, int64(pkt2.Len()), conn1.BytesWritten())

		conn1.Close()
	})

	pkt := packet.NewConnectPacket()
	conn2.Send(pkt)
	require.Equal(t, int64(pkt.Len()), conn2.BytesWritten())

	pkt2, _ := conn2.Receive()
	require.Equal(t, int64(pkt2.Len()), conn2.BytesRead())

	conn2.Close()
	<-done
}

func abstractConnReadLimitTest(t *testing.T, preparer abstractTestPreparer) {
	conn2, done := preparer(func(conn1 Conn){
		conn1.SetReadLimit(1)

		pkt, err := conn1.Receive()
		require.Nil(t, pkt)
		require.Equal(t, ReadLimitExceeded, toError(err).Code())
	})

	err := conn2.Send(packet.NewConnectPacket())
	require.NoError(t, err)

	pkt, err := conn2.Receive()
	require.Nil(t, pkt)
	require.Equal(t, ExpectedClose, toError(err).Code())

	<-done
}