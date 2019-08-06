package productsubscription

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	"golang.org/x/crypto/ssh"
)

func TestProductSubscriptions(t *testing.T) {
	ctx := dbtesting.TestContext(t)

	// devPublicKey is the public key used to verify product license keys generated in dev environments.
	//
	// It is hardcoded here intentionally. Keys generated in dev environments can't be used for production
	// instances. The corresponding private key can be found in the https://github.com/sourcegraph/dev-private
	// repository.
	licensing.Mocks.PublicKey = (func() ssh.PublicKey {
		const devPublicKeyData = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCv5OBSBiNblorj5K33ZhVgQqXGQfy64l2pwZ0MviHE9pbzLtPuFI9HAKCGH+zz6Wj8oXs5zBA4gtJ1z5CJL3MsAw7K7BoilrruXEP41TUZEuI8BzSC1yLdxEIvE+ScOr6v0iypR0pzrtQQQpgmZNbX0giRHAPtyqsyhbeFPLs+srJ/HXMCZBj1egd05P7sXxQ7nUwuklXDC/f2Jf3ebrL20nw65vxdPSg512KhjNYlc1R4idn1pZDcTwSloWG6Fz032mphuJJ6pZ3CXZIio/SoPoLVOHxzYWvcD30xgGigFDd6v4SQRmJAWmLLxgFdd/PKnw4KjIANML+G1NvJpYKv`
		var err error
		devPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(devPublicKeyData))
		if err != nil {
			panic("failed to parse public key for license verification: " + err.Error())
		}
		return devPublicKey
	}())

	u1, err := db.Users.Create(ctx, db.NewUser{Username: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	u2, err := db.Users.Create(ctx, db.NewUser{Username: "u2"})
	if err != nil {
		t.Fatal(err)
	}
	u3, err := db.Users.Create(ctx, db.NewUser{Username: "u3"})
	if err != nil {
		t.Fatal(err)
	}
	u4, err := db.Users.Create(ctx, db.NewUser{Username: "u4"})
	if err != nil {
		t.Fatal(err)
	}

	ps1, err := dbSubscriptions{}.Create(ctx, u1.ID)
	if err != nil {
		t.Fatal(err)
	}
	ps2, err := dbSubscriptions{}.Create(ctx, u2.ID)
	if err != nil {
		t.Fatal(err)
	}
	ps3, err := dbSubscriptions{}.Create(ctx, u3.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dbSubscriptions{}.Create(ctx, u4.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Old key
	_, err = dbLicenses{}.Create(ctx, ps1, "eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJLRW54d0N0Mnp0dFFFTDdwN1dLekwvYU94RWMxZEtIRlUzUmRGMGdYbFExU1FjejRXcjNLb2lscm1MRklBTkF4WGFwRGFMOVpoa011KzNGM25VWmtGY0FBZ1Z2SFFoV2FtQ0VOcnhZSnlDSmFqZVE5VGFOQjY2Z1ZtQUNOaW5CZU02cXhjNm1BQjN5OUtEVEwvQkhMczliSmdKdkRrLzg2cXBISGdRK1F5QnpWYzA2UnpxajNFU0d6OEpPYks0d1lFTWNRTGMzSEp6eFZuUVp6ZHc4UmJaRTlvRmNMSldHL1lwRVZZYlJub2tLWVVnM1dvRlhxaTQ5eUd4OVRXcVR3Zkl1VFZMamNmWnN1YVFaTUtRdTZLTStudUh4Y0tGbExCNXJxRVowOTVoTDFNWU44WHpMMDlWTmNZL2I1YU5LZldlVXllcnJ3dGhJZHlFZ1k4VzV6Mmc9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1UUXNPREFzTVRRNUxESXdOaXd5TVRBc01UZzRMREUyT1N3eE1qVmRMQ0owSWpwYkluUnlkV1V0ZFhBaVhTd2lkU0k2TVN3aVpTSTZJakl3TVRrdE1EVXRNVFpVTURBNk1EQTZNREF0TURjNk1EQWlmUT09In0")
	if err != nil {
		t.Fatal(err)
	}
	// Expires at: May 16, 2019 12:00:00 AM PT
	_, err = dbLicenses{}.Create(ctx, ps1, "eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJXbGFoODFWSTkrYVJmK08zY1lUTk5tK3VQNWxxUVY1YlFPaGpIYm1KOFVCeFZvZEllRTIrT0ZscHhrUnNkam9VcWxVd2UxNTRvM2pZMm5yMmVyMkl2S2dWZ2ZENnVWYjUxcGRlcWp5Vk91Z2NRNjIxaTdVb0xhay9rQmRSOENxTEg0VE5qUlVUTmFiYlY2NW5kQkdtb3lQa2tQVGZPWlVnVHQvSkdGUGI5MkVIR1crbm9ka2xSSWUxSVVXVDFQMkx6aU1vaFpwOStMUUYwdzV1RGQ2MjFWeHVtUGFiZXhHYTB5U3BDY3lDNTltZnQraHdJWFJoeDlJdUU0eFZZa0gxbGpKVjhqaFp3UVVTVTV0bmNQM3hWeC9SRHdTaHlFWlhkRkJGM3hRKytVT0FMaDRDTkUxSGV3dlZTZHZ2dFQ5eEpLdWRKejlCWlk3KzVTTzZoMVVvRkE9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk5UVXNNVE01TERJME5Dd3hORE1zTWpBMUxERXpOeXd4Tmprc01qSTNYU3dpZENJNld5SjBjblZsTFhWd0lsMHNJblVpT2pFc0ltVWlPaUl5TURFNUxUQTFMVEUyVkRBd09qQXdPakF3TFRBM09qQXdJbjA9In0")
	if err != nil {
		t.Fatal(err)
	}
	// Expires at: May 26, 2019 12:00:00 AM PT
	_, err = dbLicenses{}.Create(ctx, ps2, "eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJreDkxOHdMeXdLR01CTndCZUdMaERMNUxCTmVBYmMyc20wcEl4SXR3VEMybnQ0VlJBZm9wODBFWXhpYzBlem15S3l6L3AxcFp4Q1NObjNMUjNDWUpBTWpJSnRJNVl2SVpwbG84bkJzZkViaGdTS2ZDaG00T1ZzNDNMN2NaWm4yQ0o3RlNxNGJjSWZRU2Z1Zjh1R2JaU3JNdm5zeDdYRnRqQWZQdSsvMGlObGlTOWFpdWx3ZGgrRTJvVjcvVENRUnJ2cEwxQ2o0cFRSa01wc09xRFdpZ2o0RXQyV25XcjJ5eFFEdUJjdU1yMHZ2U0o2SDZVZmsxT0NvbzQ0OHBxbExQRE1CMWs5dHJsRWJrQWJSQWNCSVM1WnRXWTRrL2ZOaUx0MXBWdWcyYkZrSFhETi9YR3U2NTZwaEV1dWN5ODlCcEdKaW5uN2Y2YVM4L1U5d1dia0djb3c9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1URTRMREkwTlN3NU5TdzFMREUwTERFNE5pdzBPQ3d5TWpGZExDSjBJanBiSW5SeWRXVXRkWEFpTENKMGNtbGhiQ0pkTENKMUlqb3hMQ0psSWpvaU1qQXhPUzB3TlMweU5sUXdNRG93TURvd01DMHdOem93TUNKOSJ9")
	if err != nil {
		t.Fatal(err)
	}
	// Expires at: May 17, 2019 12:00:00 AM PT
	_, err = dbLicenses{}.Create(ctx, ps3, "eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJZUG5EOXYvQXZHTlliZkVwbElrQnRLVGM1dVFUakpCbnlNQ2M2cWpQTTB0OXVSSzNiWktkV0l4Q0hhcHFzdlREQVRhTHl0MnA4Ym1VSGZrbThGNUxZd0NnbWltbUlaTVhZTElPZ2wrN2NlcWgzenJrcUJOVWwvY0c2dXE1eVNLQnlkNnNXSFFXZGhwb2FuUDJEQmtiRnBYaVFndkcvQ0xHZ1FmUXlNM1dDRklySThvOURDVUpUbW9YTUpQMHZzZ1Z0M0dZcXFYcmNka3o2N2p1aXl2WjY0S0diaW9IdmhRWFNCRldGcDNUaDMwSW5XTngzK2FaUnFpR0hzWDhBamRsMTlKWWllTXlmdHl3dXIyRnFXZkhla3BpZ2tKcURaTUEwc3ZuelZPSFRRZWFHSk9qQ2xrbTJqRXNUbzZ3QTFEalR5MjlTS28wN3BSaWNobVlER1duR3c9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1qSTVMRFlzTVRJNExEWXdMRFEzTERFME1pd3lNRElzTVRjM1hTd2lkQ0k2V3lKMGNuVmxMWFZ3SWwwc0luVWlPakVzSW1VaU9pSXlNREU1TFRBMUxURTNWREF3T2pBd09qQXdMVEEzT2pBd0luMD0ifQ")
	if err != nil {
		t.Fatal(err)
	}

	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(orig) // reset

	subs, err := (ProductSubscriptionLicensingResolver{}).ProductSubscriptions(ctx, &graphqlbackend.ProductSubscriptionsArgs{})
	if err != nil {
		t.Fatal(err)
	}

	nodes, err := subs.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}

	var actual []string
	for _, node := range nodes {
		al, err := node.ActiveLicense(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if al == nil {
			actual = append(actual, "")
		} else {
			actual = append(actual, al.LicenseKey())
		}
	}

	expected := []string{
		"eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJXbGFoODFWSTkrYVJmK08zY1lUTk5tK3VQNWxxUVY1YlFPaGpIYm1KOFVCeFZvZEllRTIrT0ZscHhrUnNkam9VcWxVd2UxNTRvM2pZMm5yMmVyMkl2S2dWZ2ZENnVWYjUxcGRlcWp5Vk91Z2NRNjIxaTdVb0xhay9rQmRSOENxTEg0VE5qUlVUTmFiYlY2NW5kQkdtb3lQa2tQVGZPWlVnVHQvSkdGUGI5MkVIR1crbm9ka2xSSWUxSVVXVDFQMkx6aU1vaFpwOStMUUYwdzV1RGQ2MjFWeHVtUGFiZXhHYTB5U3BDY3lDNTltZnQraHdJWFJoeDlJdUU0eFZZa0gxbGpKVjhqaFp3UVVTVTV0bmNQM3hWeC9SRHdTaHlFWlhkRkJGM3hRKytVT0FMaDRDTkUxSGV3dlZTZHZ2dFQ5eEpLdWRKejlCWlk3KzVTTzZoMVVvRkE9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk5UVXNNVE01TERJME5Dd3hORE1zTWpBMUxERXpOeXd4Tmprc01qSTNYU3dpZENJNld5SjBjblZsTFhWd0lsMHNJblVpT2pFc0ltVWlPaUl5TURFNUxUQTFMVEUyVkRBd09qQXdPakF3TFRBM09qQXdJbjA9In0",
		"eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJZUG5EOXYvQXZHTlliZkVwbElrQnRLVGM1dVFUakpCbnlNQ2M2cWpQTTB0OXVSSzNiWktkV0l4Q0hhcHFzdlREQVRhTHl0MnA4Ym1VSGZrbThGNUxZd0NnbWltbUlaTVhZTElPZ2wrN2NlcWgzenJrcUJOVWwvY0c2dXE1eVNLQnlkNnNXSFFXZGhwb2FuUDJEQmtiRnBYaVFndkcvQ0xHZ1FmUXlNM1dDRklySThvOURDVUpUbW9YTUpQMHZzZ1Z0M0dZcXFYcmNka3o2N2p1aXl2WjY0S0diaW9IdmhRWFNCRldGcDNUaDMwSW5XTngzK2FaUnFpR0hzWDhBamRsMTlKWWllTXlmdHl3dXIyRnFXZkhla3BpZ2tKcURaTUEwc3ZuelZPSFRRZWFHSk9qQ2xrbTJqRXNUbzZ3QTFEalR5MjlTS28wN3BSaWNobVlER1duR3c9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1qSTVMRFlzTVRJNExEWXdMRFEzTERFME1pd3lNRElzTVRjM1hTd2lkQ0k2V3lKMGNuVmxMWFZ3SWwwc0luVWlPakVzSW1VaU9pSXlNREU1TFRBMUxURTNWREF3T2pBd09qQXdMVEEzT2pBd0luMD0ifQ",
		"eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJreDkxOHdMeXdLR01CTndCZUdMaERMNUxCTmVBYmMyc20wcEl4SXR3VEMybnQ0VlJBZm9wODBFWXhpYzBlem15S3l6L3AxcFp4Q1NObjNMUjNDWUpBTWpJSnRJNVl2SVpwbG84bkJzZkViaGdTS2ZDaG00T1ZzNDNMN2NaWm4yQ0o3RlNxNGJjSWZRU2Z1Zjh1R2JaU3JNdm5zeDdYRnRqQWZQdSsvMGlObGlTOWFpdWx3ZGgrRTJvVjcvVENRUnJ2cEwxQ2o0cFRSa01wc09xRFdpZ2o0RXQyV25XcjJ5eFFEdUJjdU1yMHZ2U0o2SDZVZmsxT0NvbzQ0OHBxbExQRE1CMWs5dHJsRWJrQWJSQWNCSVM1WnRXWTRrL2ZOaUx0MXBWdWcyYkZrSFhETi9YR3U2NTZwaEV1dWN5ODlCcEdKaW5uN2Y2YVM4L1U5d1dia0djb3c9PSJ9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1URTRMREkwTlN3NU5TdzFMREUwTERFNE5pdzBPQ3d5TWpGZExDSjBJanBiSW5SeWRXVXRkWEFpTENKMGNtbGhiQ0pkTENKMUlqb3hMQ0psSWpvaU1qQXhPUzB3TlMweU5sUXdNRG93TURvd01DMHdOem93TUNKOSJ9",
		"",
	}

	if !reflect.DeepEqual(actual, expected) {
		fmt.Println(actual)
		t.Fatalf("Unexpected license keys:\n%s", cmp.Diff(actual, expected))
	}

}
