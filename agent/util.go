package agent

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/go-msgpack/codec"
)

// msgpackHandle is a shared handle for encoding/decoding of
// messages
var msgpackHandle = &codec.MsgpackHandle{
	RawToString: true,
	WriteExt:    true,
}

// decodeMsgPack is used to decode a MsgPack encoded object
func decodeMsgPack(buf []byte, out interface{}) error {
	return codec.NewDecoder(bytes.NewReader(buf), msgpackHandle).Decode(out)
}

// encodeMsgPack is used to encode an object with msgpack
func encodeMsgPack(msg interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := codec.NewEncoder(&buf, msgpackHandle).Encode(msg)
	return buf.Bytes(), err
}

// stringHash returns a simple md5sum for a string.
func stringHash(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// checkIDHash returns a simple md5sum for a types.CheckID.
func checkIDHash(checkID types.CheckID) string {
	return stringHash(string(checkID))
}

// FilePermissions is an interface which allows a struct to set
// ownership and permissions easily on a file it describes.
type FilePermissions interface {
	// User returns a user ID or user name
	User() string

	// Group returns a group ID. Group names are not supported.
	Group() string

	// Mode returns a string of file mode bits e.g. "0644"
	Mode() string
}

// setFilePermissions handles configuring ownership and permissions settings
// on a given file. It takes a path and any struct implementing the
// FilePermissions interface. All permission/ownership settings are optional.
// If no user or group is specified, the current user/group will be used. Mode
// is optional, and has no default (the operation is not performed if absent).
// User may be specified by name or ID, but group may only be specified by ID.
func setFilePermissions(path string, p FilePermissions) error {
	var err error
	uid, gid := os.Getuid(), os.Getgid()

	if p.User() != "" {
		if uid, err = strconv.Atoi(p.User()); err == nil {
			goto GROUP
		}

		// Try looking up the user by name
		if u, err := user.Lookup(p.User()); err == nil {
			uid, _ = strconv.Atoi(u.Uid)
			goto GROUP
		}

		return fmt.Errorf("invalid user specified: %v", p.User())
	}

GROUP:
	if p.Group() != "" {
		if gid, err = strconv.Atoi(p.Group()); err != nil {
			return fmt.Errorf("invalid group specified: %v", p.Group())
		}
	}
	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("failed setting ownership to %d:%d on %q: %s",
			uid, gid, path, err)
	}

	if p.Mode() != "" {
		mode, err := strconv.ParseUint(p.Mode(), 8, 32)
		if err != nil {
			return fmt.Errorf("invalid mode specified: %v", p.Mode())
		}
		if err := os.Chmod(path, os.FileMode(mode)); err != nil {
			return fmt.Errorf("failed setting permissions to %d on %q: %s",
				mode, path, err)
		}
	}

	return nil
}
