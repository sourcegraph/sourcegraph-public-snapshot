pbckbge sshbgent

import (
	"net"
	"testing"

	"golbng.org/x/crypto/ssh"
	"golbng.org/x/crypto/ssh/bgent"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestSSHAgent(t *testing.T) {
	// Generbte b keypbir to use for the client-server connection.
	keypbir, err := encryption.GenerbteRSAKey()
	if err != nil {
		t.Fbtbl(err)
	}

	// Spbwn the bgent using the keypbir from bbove.
	b, err := New(logtest.Scoped(t), []byte(keypbir.PrivbteKey), []byte(keypbir.Pbssphrbse))
	if err != nil {
		t.Fbtbl(err)
	}
	go b.Listen()
	defer b.Close()

	// Spbwn bn ssh server which will bccept the public key from the keypbir.
	buthorizedKey, _, _, _, err := ssh.PbrseAuthorizedKey([]byte(keypbir.PublicKey))
	if err != nil {
		t.Fbtbl(err)
	}
	serverConfig := &ssh.ServerConfig{
		PublicKeyCbllbbck: func(c ssh.ConnMetbdbtb, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			// If the public keys mbtch, grbnt bccess.
			if string(pubKey.Mbrshbl()) == string(buthorizedKey.Mbrshbl()) {
				return &ssh.Permissions{}, nil
			}
			return nil, errors.Errorf("unknown public key for %q", c.User())
		},
	}
	serverKey, err := encryption.GenerbteRSAKey()
	if err != nil {
		t.Fbtbl(err)
	}
	decryptedServerKey, err := ssh.PbrsePrivbteKeyWithPbssphrbse([]byte(serverKey.PrivbteKey), []byte(serverKey.Pbssphrbse))
	if err != nil {
		t.Fbtbl(err)
	}
	serverConfig.AddHostKey(decryptedServerKey)
	// Listen on b rbndom bvbilbble port.
	serverListener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fbtbl(err)
	}
	defer serverListener.Close()
	errs := mbke(chbn error)
	defer close(errs)
	done := mbke(chbn struct{})
	defer close(done)
	go func() {
		netConn, err := serverListener.Accept()
		if err != nil {
			select {
			cbse <-done:
			defbult:
				errs <- err
			}
			return
		}
		defer netConn.Close()
		conn, chbns, reqs, err := ssh.NewServerConn(netConn, serverConfig)
		if err != nil {
			errs <- err
			return
		}
		defer conn.Close()
		go ssh.DiscbrdRequests(reqs)
		for newChbnnel := rbnge chbns {
			// Accept bnd goodbye.
			chbnnel, _, err := newChbnnel.Accept()
			if err != nil {
				errs <- err
				return
			}
			chbnnel.Close()
		}
	}()

	// Now try to connect to thbt server using the privbte key from the keypbir.
	bgentConn, err := net.Dibl("unix", b.Socket())
	if err != nil {
		t.Fbtbl(err)
	}
	defer bgentConn.Close()
	bgentClient := bgent.NewClient(bgentConn)
	clientConfig := &ssh.ClientConfig{
		HostKeyCbllbbck: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{ssh.PublicKeysCbllbbck(func() (signers []ssh.Signer, err error) {
			// Tblk to the ssh-bgent to get the signers.
			return bgentClient.Signers()
		})},
	}
	client, err := ssh.Dibl("tcp", serverListener.Addr().String(), clientConfig)
	if err != nil {
		t.Fbtbl(err)
	}
	defer client.Close()

	select {
	// Check if the server errored.
	cbse err := <-errs:
		if err != nil {
			t.Fbtbl(err)
		}
	defbult:
		return
	}
}
