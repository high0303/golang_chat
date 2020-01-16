package lib

import (
	"encoding/binary"
	"errors"
	//"log"
	"sechat-server/scutil"
	//"crypto/x509"
)

const HEADER_SIZE = 12

type FrameHolder struct {
	JSON             []byte
	JSONLength       uint32
	BinaryData       []byte
	BinaryDataLength uint32
}

func CreateFrame(frame_data []byte) []byte {
	frame_data_length := uint32(len(frame_data))

	scutil.FramingLog("Frame len: %d", frame_data_length)
	// Our header is 12 bytes long
	data := make([]byte, HEADER_SIZE+frame_data_length)
	data[0] = 'Y'
	data[1] = 'O'
	data[2] = 'L'
	data[3] = 'O'
	binary.BigEndian.PutUint32(data[4:8], frame_data_length)
	binary.BigEndian.PutUint32(data[8:12], 0)
	if copy(data[HEADER_SIZE:], frame_data) != len(frame_data) {
		scutil.ErrorLog("server: createFrame: failed to copy frame data!")
		// todo error
	}
	return data
}

func CreateFrameWithData(frame_json []byte, frame_data []byte) []byte {
	frame_json_length := uint32(len(frame_json))
	frame_data_length := uint32(len(frame_data))

	scutil.FramingLog("Frame len: %d", frame_json_length+frame_data_length)
	// Our header is 12 bytes long
	data := make([]byte, HEADER_SIZE+frame_json_length+frame_data_length)
	data[0] = 'Y'
	data[1] = 'O'
	data[2] = 'L'
	data[3] = 'O'
	binary.BigEndian.PutUint32(data[4:8], frame_json_length)
	binary.BigEndian.PutUint32(data[8:12], frame_data_length)
	if copy(data[HEADER_SIZE:], frame_json) != len(frame_json) {
		panic("server: createFrame: failed to copy frame json!")
	}
	if copy(data[HEADER_SIZE+frame_json_length:], frame_data) != len(frame_data) {
		panic("server: createFrame: failed to copy frame data!")
	}
	return data
}

func ReadFrameHeader(data []byte) (*FrameHolder, error) {
	frameHolder := FrameHolder{}
	delimiter := string(data[0:4])

	if delimiter == "YOLO" {
		scutil.FramingLog("server: readFrameHeader: Valid header found.")
		// todo check for maxlength to avoid dos
		frameHolder.JSONLength = binary.BigEndian.Uint32(data[4:8])
		frameHolder.JSON = make([]byte, frameHolder.JSONLength)
		frameHolder.BinaryDataLength = binary.BigEndian.Uint32(data[8:12])
		frameHolder.BinaryData = make([]byte, frameHolder.BinaryDataLength)
		scutil.FramingLog("server: readFrameHeader: JSON Length: %d", frameHolder.JSONLength)
		scutil.FramingLog("server: readFrameHeader: DATA Length: %d", frameHolder.BinaryDataLength)
	} else {
		scutil.ErrorLog("Invalid header!")
		return nil, errors.New("ReadFrameHeader: Invalid header.")
	}

	return &frameHolder, nil
}
