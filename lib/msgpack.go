package lib

import (
	//"encoding/binary"
	//"errors"
	//"io"
	//"log"
	//"crypto/x509"
	"github.com/ugorji/go/codec"
)

func MpUnmarshal(data []byte, v interface{}) error {
	var mh codec.MsgpackHandle
	mh.RawToString = false
	mh.WriteExt = false
	h := &mh
	dec := codec.NewDecoderBytes(data, h)
	err := dec.Decode(&v)
	if err != nil {
		panic("Unmarshalling failed")
	}
	return err
}

func MpMarshal(v interface{}) []byte {
	var b []byte
	var mh codec.MsgpackHandle
	mh.RawToString = false
	mh.WriteExt = true
	h := &mh
	enc := codec.NewEncoderBytes(&b, h)
	err := enc.Encode(v)
	if err != nil {
		panic("Marshalling failed")
	}
	return b
}
