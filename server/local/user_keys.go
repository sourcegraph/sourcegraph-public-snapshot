package local

import (
	"encoding/base64"
	"log"
	"os"
	"strconv"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	pstorage "src.sourcegraph.com/sourcegraph/platform/storage"
)

var UserKeys sourcegraph.UserKeysServer = &userKeys{}

type userKeys struct{}

var _ sourcegraph.UserKeysServer = (*userKeys)(nil)

func (s *userKeys) AddKey(ctx context.Context, key *sourcegraph.SSHPublicKey) (*pbtypes.Void, error) {
	defer noCache(ctx)
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	keyID := int64(0)
	userKV := pstorage.Namespace(ctx, "core.ssh-keys", "")
	data, err := userKV.Get(strconv.FormatInt(int64(actor.UID), 10), "current_index")
	if err == nil {
		keyID, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	key.Id = uint64(keyID)
	log.Printf("Bucket: %s Key.Id: %d Adding: %#v", strconv.FormatInt(int64(actor.UID), 10), keyID, key)
	err = pstorage.PutJSON(userKV, strconv.FormatInt(int64(actor.UID), 10), strconv.FormatInt(keyID, 10), key)
	if err != nil {
		return nil, err
	}

	// Increment the start index to ensure sequential SSHKey IDs
	err = userKV.Put(strconv.FormatInt(int64(actor.UID), 10), "current_index", []byte(strconv.FormatInt(keyID+1, 10)))
	if err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}

// LookupUser looks up user by key. The returned UserSpec will only have UID field set.
func (s *userKeys) LookupUser(ctx context.Context, key *sourcegraph.SSHPublicKey) (*sourcegraph.UserSpec, error) {
	defer noCache(ctx)

	//TODO: Implement
	userSpec := sourcegraph.UserSpec{}
	return &userSpec, nil
}

func (s *userKeys) ListKeys(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.SSHKeyList, error) {
	defer noCache(ctx)
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	userKV := pstorage.Namespace(ctx, "core.ssh-keys", "")
	keys, err := userKV.List(strconv.FormatInt(int64(actor.UID), 10))
	if err != nil {
		return nil, err
	}

	if len(keys) < 2 {
		return &sourcegraph.SSHKeyList{}, nil
	}

	sshKeyList := make([]sourcegraph.SSHPublicKey, len(keys)-1)
	for x, key := range keys {
		if key == "current_index" {
			continue
		}

		var data = struct {
			Key, Name string
			Id        uint64
		}{}

		if err := pstorage.GetJSON(userKV, strconv.FormatInt(int64(actor.UID), 10), key, &data); err != nil {
			return nil, err
		}

		sshKeyList[x].Name = data.Name
		sshKeyList[x].Id = data.Id
		sshKeyList[x].Key = []byte{}

		keyBytes, err := base64.StdEncoding.DecodeString(data.Key)
		if err != nil {
			log15.Error("Failed to base64 decode user ssh key", err)
			continue
		}

		pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyBytes))
		if err != nil {
			log15.Error("Failed to parse user ssh key", err)
			continue
		}

		tmpKey := ssh.MarshalAuthorizedKey(pubKey)
		sshKeyList[x].Key = tmpKey
	}

	return &sourcegraph.SSHKeyList{SSHKeys: sshKeyList}, nil
}

func (s *userKeys) ClearKeys(ctx context.Context, _ *pbtypes.Void) (*pbtypes.Void, error) {
	defer noCache(ctx)
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	userKV := pstorage.Namespace(ctx, "core.ssh-keys", "")
	err := userKV.Delete(strconv.FormatInt(int64(actor.UID), 10), "")
	if err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}

func (s *userKeys) DeleteKey(ctx context.Context, key *sourcegraph.SSHPublicKey) (*pbtypes.Void, error) {
	defer noCache(ctx)
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	userKV := pstorage.Namespace(ctx, "core.ssh-keys", "")
	err := userKV.Delete(strconv.FormatInt(int64(actor.UID), 10), string(key.Id))
	if err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}
