// Package bencode implements encoding and decoding of bencoded data.
// Bencode is a data serialization format used by the BitTorrent file-sharing system.
// For more information, see the BitTorrent specification:
// https://www.bittorrent.org/beps/bep_0003.html
package bencode

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Unmarshal parses bencoded data from a reader and returns the corresponding Go value.
// It supports the following bencode types:
// - integers (i...e) are unmarshaled into int64
// - strings (<length>:<string>) are unmarshaled into string
// - lists (l...e) are unmarshaled into []interface{}
// - dictionaries (d...e) are unmarshaled into map[string]interface{}
//
// The function automatically handles buffering for the provided io.Reader.
func Unmarshal(r io.Reader) (interface{}, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	b, err := br.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case 'd':
		return unmarshalDict(br)
	case 'l':
		return unmarshalList(br)
	case 'i':
		return unmarshalInt(br)
	default:
		err := br.UnreadByte()
		if err != nil {
			return nil, err
		}

		return unmarshalString(br)
	}
}

// unmarshalDict parses a bencoded dictionary from the reader.
// Dictionaries are expected to be in the format 'd<key><value>...e'.
// Keys must be bencoded strings. Values can be any bencode type.
func unmarshalDict(br *bufio.Reader) (map[string]interface{}, error) {
	dict := make(map[string]interface{})
	for {
		b, err := br.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == 'e' {
			return dict, nil
		}
		br.UnreadByte()

		key, err := unmarshalString(br)
		if err != nil {
			return nil, err
		}

		val, err := Unmarshal(br)
		if err != nil {
			return nil, err
		}

		dict[key] = val
	}
}

// unmarshalList parses a bencoded list from the reader.
// Lists are expected to be in the format 'l<value>...e'.
// Values can be any bencode type.
func unmarshalList(br *bufio.Reader) ([]interface{}, error) {
	var list []interface{}
	for {
		b, err := br.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == 'e' {
			return list, nil
		}
		br.UnreadByte()

		val, err := Unmarshal(br)
		if err != nil {
			return nil, err
		}

		list = append(list, val)
	}
}

// unmarshalInt parses a bencoded integer from the reader.
// Integers are expected to be in the format 'i<integer>e'.
func unmarshalInt(br *bufio.Reader) (int64, error) {
	data, err := br.ReadBytes('e')
	if err != nil {
		return 0, err
	}

	// Trim the 'e'
	s := string(data[:len(data)-1])
	if s == "" {
		return 0, fmt.Errorf("bencode: empty integer")
	}
	
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// unmarshalString parses a bencoded string from the reader.
// Strings are expected to be in the format '<length>:<string>'.
func unmarshalString(br *bufio.Reader) (string, error) {
	lenStr, err := br.ReadString(':')
	if err != nil {
		return "", err
	}

	length, err := strconv.Atoi(lenStr[:len(lenStr)-1])
	if err != nil {
		return "", err
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(br, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// Marshal returns the bencode encoding of data.
//
// This implementation is partial and currently only supports marshaling
// map[string]interface{} types, which is sufficient for creating
// extended handshake messages for the BitTorrent client.
// A complete implementation would handle more types (integers, strings, lists).
func Marshal(w io.Writer, data interface{}) error {
	// A full implementation would be needed for a seeder/uploader.
	// For this client, we only need to encode the extended handshake message.
	switch v := data.(type) {
	case map[string]interface{}:
		return marshalDict(w, v)
	default:
		return fmt.Errorf("bencode: unsupported type for marshaling: %T", data)
	}
}

// marshalDict writes a bencoded dictionary to the writer.
// It encodes the map into the 'd<key><value>...e' format.
// It currently supports int and map[string]interface{} as value types.
func marshalDict(w io.Writer, dict map[string]interface{}) error {
	if _, err := w.Write([]byte("d")); err != nil {
		return err
	}
	for k, v := range dict {
		// Marshal key (string)
		if _, err := fmt.Fprintf(w, "%d:%s", len(k), k); err != nil {
			return err
		}
		// Marshal value
		switch val := v.(type) {
		case int:
			if _, err := fmt.Fprintf(w, "i%de", val); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := marshalDict(w, val); err != nil {
				return err
			}
		default:
			return fmt.Errorf("bencode: unsupported value type in dict: %T", val)
		}
	}
	if _, err := w.Write([]byte("e")); err != nil {
		return err
	}
	return nil
}
