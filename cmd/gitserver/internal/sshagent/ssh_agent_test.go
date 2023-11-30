package sshagent

import (
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var testKeypair = &encryption.RSAKey{
	// This is a bogus key generated solely for this test and not used by anything.
	// It was created using encryption.GenerateRSAKey().
	PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nProc-Type: 4,ENCRYPTED\nDEK-Info: AES-256-CBC,61eab095a9eb3d8a87dcb013a6daa726\n\nIALaMk9OMa4pobRzvG7pxDt4hmWXWjj5XaM4bx1dskZtErLpbCBryoDTmiA1PQP2\nuUl8ylCNo9jZpagaMZCCAvrLpUJxjGdEzIVDwU6OfVip5dWS3kl40JoILbJ4JGyK\nbTSOrl8Y/xa3V54mUXxfMXKS6n5S3WR5fvPgmCetYKYQ9nQxz7tY1rWZ7Ug9rUUd\nloIF9UQ/PUWkHP44EP5TyzhloMQsJ1TCGt8g+ICNrRYTT59dZy2CQnZoyZJtfHMh\nmU8dyqEy7qrt4z8FtTt1sMUhNrF5s4zxj0U5ZPzHM+3cQwVpm3gtX8isQ30T8yUB\nHJFhOlPipxpvdmTwOkBBDLtI8Zx1rxRDwn3g0ouQke5dRrekwCp3AmYchY11YUeu\nXGYWf7XuRLx2pRimqRC++yEmgaALw7MWwKZm1x0LsP4XeUTHgJeN2c/nEVehJj1+\nU8m3lsM+BOVG1Uinn0WBjr+ARI44M8w3NI5fU5B4b6XvSSPB3fJK8KXIH+OxiqGS\nAv2zFiElJiDto1uWnyMuVh8GtWTlmkAg0T5QubFBCpf5gUjXZXpp1hhCKPUNyPU+\nctcHoCkmdeiFCDUasec+GoB80eH2Ea5L0GCrU686G/wh1QNzV8/zPnlsiXHO+Hfw\nNJHtrTCQpRvfJYxVSG287DDSY6JBOwi+BNE3sm5+YK31qCFviCEuItg3g9ZDc/BG\n6swP/NZynwVMBdatfNLzcadk6UTFGiWfQNVYMy0Qy3Yt51pKYrSF/JvQkHZZDkko\nIaKtv9dJEDLdJF4PfsF93J2jyhrtPfNEeCYWW1Y7RwEKSJDxNQyTSWAs3rKxrbU/\nUeKx7kogZ7EUz0+WH237eWDnyyuKmc6R2hu8tg93272niiRJ7pDbbV6qGPbhcAWW\nhT5x9tisKI48Ac0RSgWO6KkqzPmykcIFRn+nSwUBO3aiLbQrOqXjnTxnmP92yotE\nRYyelm2Eppv9tFo4wENaDjhuwN1R2vkx4aepVLfWIk8n1lQo/3Dj21xHfTW6B8lD\n1vCUw2BaEGeL47ap0PO2+zOWcP2dc0zoS213JR+pENGp32zoA0DMVDfiNBzdYndr\nLouhrVpgFbw1K9cgwkjO1iYVpTclTbaeiT0n/ZEHfjc26rxx53ALv06LktaXVJSA\nY9pb2QpPnX0chaL1M54e3E1pXT/BdXab1eCLHUHPBc7BUR7IvmBaNCcPuJrrSZ66\nY3uqn3DoKcztUuEvyC//xJUqxtsh6pM4Ng54HxzyzX63/GoDVH4ZqxDsaoIcliyu\n2/L43m8YQeH2NpYgB5585Qp1vN+Dr0eBjn9SGGj1vO/0FoqLG2I0Cp/E/QgsUCWQ\nFQREmHQNZnUK2pLPgAKE5fpnPSELeRLz5xEwizisRCTuQ28zHT5XDwj2JrCGVMww\nH0q22MXsQ8uzJdjg/HEWZnyoiX5jeTPu6KWo6e0IaavFY3UOZ7KvFrZ2LPgDDlGO\nvngeys4M0/EbRp/1T0pbJY+XCovo/rokemAWgnRDyNwSHk1Ps/XPimFR9O1nkV9b\nP7U4NSvZkMdu14ZUiXzGGc0qJnky0v9b+vvFJmbB1jHRXLFyF1sQ2T2ISlujGbqZ\nyH1VEbyzOjxnx5Ck6p2m1ZdoCYD230NQwU6Z4kvdN5EBFepTnF0unRZOQQ3gAFde\n0Bw04rot7of6lypfUnUQJrSsY6G2w1yogeSi2eoZtblMq5knj/Xro49G1r/ytJXQ\nhlpdePKM5OhY881qaaDHY2L+IO8w9syPOJEkVjkoJJKz7K3S4j+V1P9VfhY1E4/l\n0bVKL2EIgFZRSPPILknW49qMi/AQ2edOj+e29zmJXMDBxy5mMOC4O9UI7qGZaCDs\nheOogapKcWiPku6EiLGgBT/vSqTSI2u1C7Q7H59ckdXy9BTfKFKyHIJzJ7q1ew7L\nNmBnLV/G56KVz1eUe05L6vWSrWIW1j+At75aKfGWNU0SmTznrL4FB8DSrWw0+KtT\nNmALnNBiLx0ddIUVNpmpESK/aSi0jPzTON09o1jvzH+FGsc5R1XnxtSq+nfL/FyO\nhEgIy/jUOMIU/ERUUjTFWhf/tiRekmjL4wj596p+oeFkHmJbQ2WFtcOwdbx9Xg7k\neSGFQI3/OXa0YsPo1QR0EgF06Wa4VDnHlxGqWEj+dl8tPjnrcYfwVVkQDJiFCqbW\n7jBBKPJGPOEOfvNdeUNSgyzoqU4CPBLYrYd88JiQ+Z9zBxD33Og98DTL3exyPQ8c\nWZ4scvZ+fafcMqOXK4r+wcBQ9QoB8KP8TSE/bh4xZTNiuwM8N9T8anE4Iw1fDyck\nW8R7t/spE1/O9ETcPAMYmv3uIlPxhQtSUbVyqjUHxWWFCCE3pblhOSSLW0psF9OC\ndRLeFUB3yQ1ffui05TgaFih/XLFAOWBLnAd/K3Ps4kU3uxkCnhhJd1RaYIflKHyp\nRF3bSwVqEmMHS+fyaUfcNDVQYCVj0hk1IpMfECxFDrJOsDUwYLKxYIZjmyhBj0zd\nI5w0YmY2EZ2W6gpKCC40LFDUGZSRUx4Zn22odFwB5bzg+rdNb1wCNkfjK34d6ZSD\ndDpWI07mnMJ+/JRkZxPN9Xo7shAgd5QrcvWEgTa4saP8yokcjvXJHPaeTjR8WSOV\nVBQa0xAL9nUxHwv1B1CRjIIomuo/JVoe1el9UygvwlEJn2r0ZOM0XbDeIb6zrHcJ\nWI0uEhjDzswQNxM2U6lJ0bXlsdJXYWTAmAqPTSqI6KNKC4mQhov6/cCQKABSG5GE\niK7OiBtCK0MeEX4uKaWdJ+Sos057F99knxJPbrppxKvNN6cf38r+B/EWj9Tz+UmA\nu5o4OTcuznOUIbZ53usd5VcYeW2BSnt5bSbLgZMFxjAGqBUxHHXc3SOH04OkkS7m\nK0JTvMWNUZvPIPQBgxQUp6IRhkgtSC2LBIncOjr4T3lXBgHjACZrNd4AW/8mN1vI\nm50L2qP0aAxhdDuHJptvarGWByjBhLNbhPgyLum3LzW7gieKBSkVg0tk8KtmUE2+\ny+wGZgcli3q67hxYJIwDtFuBroTcYkHV5HG2eZOGOAzCSLkfthRp0g7MBGs4M24S\niL/8tt2kej7T5VhdaCLg7BHaZoKFsKNXU77TRk1+GPiZkTuHJYrnBfzlTKMivrQe\n-----END RSA PRIVATE KEY-----\n",
	PublicKey:  "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC6PwFpQQjCgFtWf8DTFe5LDNhl8WIEL7zmvmElPQkFhoXWQU/kR8S/TqlwRUm0UbSbVbHbUijNNixIEnZaK4SlYsPtAK1AZWYIogGKkEm+AFpo0HXKQF+dLCjwN1o3bLcmdvxvpF7MHl5bug1t1S6BWSkkOs5Ve7GekhiHZeGMwo4Y7fskQQPQqiuvdj6YO9WICiTqdQ1WsfuiEuJXM4UoYpdBC+lgT2FOfFTWt7n4zhW16tGvk/k8hUDrBoiBIZKG2+oEIMbfmoZemp9ZIWbOJ/mgIHjQOa7ZTixXd4jJTKK08aHpdcyAg/tjLk2aAMhz2Hf5zA+y+XmgrRcZvCGywNKH349d5oCX7TZxgffLm/pYmr/NslEGyaCIgs3PzqrmfUqWotmg/HRUYu6NAiJdcNGwWNg4YLummHKtaVcC8EGlrhhM59D/OJQUPzTw9uzH0quE0FTVdYPAeHbTKe011G8Weh3x93HMCugbocjCI1TDJQO4wBAVp2wBDvFqFWXiA2uECspikLCsSizn6t4TmKUNZXTToXemE1BL1x40UfUJyTmLfss3rMSq4XfIcDo7h0mrxLiyoKxJPFGlEe9GKGGs3/uk1pB27EAxRZoh+N4UMwPOPvF8zUel5DQKlB3oSZK2XC5BNeZgSqJSpqHPtrHr0sbVbQ+PLtudtdsvmw==\n",
	Passphrase: "ab1faed5-e477-41b3-a7e2-f46a15e4fa03",
}

func TestSSHAgent(t *testing.T) {
	// Spawn the agent using the keypair from above.
	a, err := New(logtest.Scoped(t), []byte(testKeypair.PrivateKey), []byte(testKeypair.Passphrase))
	if err != nil {
		t.Fatal(err)
	}
	go a.Listen()
	defer a.Close()

	// Spawn an ssh server which will accept the public key from the keypair.
	authorizedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(testKeypair.PublicKey))
	if err != nil {
		t.Fatal(err)
	}
	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			// If the public keys match, grant access.
			if string(pubKey.Marshal()) == string(authorizedKey.Marshal()) {
				return &ssh.Permissions{}, nil
			}
			return nil, errors.Errorf("unknown public key for %q", c.User())
		},
	}
	serverKey, err := encryption.GenerateRSAKey()
	if err != nil {
		t.Fatal(err)
	}
	decryptedServerKey, err := ssh.ParsePrivateKeyWithPassphrase([]byte(serverKey.PrivateKey), []byte(serverKey.Passphrase))
	if err != nil {
		t.Fatal(err)
	}
	serverConfig.AddHostKey(decryptedServerKey)
	// Listen on a random available port.
	serverListener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer serverListener.Close()
	errs := make(chan error)
	defer close(errs)
	done := make(chan struct{})
	defer close(done)
	go func() {
		netConn, err := serverListener.Accept()
		if err != nil {
			select {
			case <-done:
			default:
				errs <- err
			}
			return
		}
		defer netConn.Close()
		conn, chans, reqs, err := ssh.NewServerConn(netConn, serverConfig)
		if err != nil {
			errs <- err
			return
		}
		defer conn.Close()
		go ssh.DiscardRequests(reqs)
		for newChannel := range chans {
			// Accept and goodbye.
			channel, _, err := newChannel.Accept()
			if err != nil {
				errs <- err
				return
			}
			channel.Close()
		}
	}()

	// Now try to connect to that server using the private key from the keypair.
	agentConn, err := net.Dial("unix", a.Socket())
	if err != nil {
		t.Fatal(err)
	}
	defer agentConn.Close()
	agentClient := agent.NewClient(agentConn)
	clientConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{ssh.PublicKeysCallback(func() (signers []ssh.Signer, err error) {
			// Talk to the ssh-agent to get the signers.
			return agentClient.Signers()
		})},
	}
	client, err := ssh.Dial("tcp", serverListener.Addr().String(), clientConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	select {
	// Check if the server errored.
	case err := <-errs:
		if err != nil {
			t.Fatal(err)
		}
	default:
		return
	}
}
