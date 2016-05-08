package client

import (
	"github.com/ccding/go-logging/logging"
	"github.com/gtfierro/cs262-project/common"
	uuidlib "github.com/satori/go.uuid"
	"os"
)

var NamespaceUUID = uuidlib.FromStringOrNil("85ce106e-0ccf-11e6-81fc-0cc47a0f7eea")

func init() {
	log, _ = logging.WriterLogger("main", logging.DEBUG, logging.BasicFormat, logging.DefaultTimeFormat, os.Stderr, true)
}

// Creates a deterministic UUID from a given name. Names are easier to remember
// than UUIDs, so this should make writing scripts easier
func UUIDFromName(name string) common.UUID {
	return common.UUID(uuidlib.NewV5(NamespaceUUID, name).String())
}
