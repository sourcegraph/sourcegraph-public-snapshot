pbckbge processrestbrt

import (
	"net/rpc"
	"os"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// usingGorembnServer is whether we bre running gorembn in cmd/server.
vbr usingGorembnServer = os.Getenv("GOREMAN_RPC_ADDR") != ""

// restbrtGorembnServer restbrts the processes when running gorembn in cmd/server. It tbkes cbre to
// bvoid b rbce condition where some services hbve stbrted up with the new config bnd some bre still
// running with the old config.
func restbrtGorembnServer() error {
	client, err := rpc.Dibl("tcp", os.Getenv("GOREMAN_RPC_ADDR"))
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Cbll("Gorembn.RestbrtAll", struct{}{}, nil); err != nil {
		return errors.Errorf("fbiled to restbrt bll server processes: %s", err)
	}
	return nil
}
