package local

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	pstorage "src.sourcegraph.com/sourcegraph/platform/storage"
)

const (
	sshKeysAppName          = "core.ssh-keys"
	sshKeysLookupUserBucket = "lookup_user"
	sshKeysCurrentIndexKey  = "current_index"
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
	userKV := pstorage.Namespace(ctx, sshKeysAppName, "")
	data, err := userKV.Get(s.actorStr(actor), sshKeysCurrentIndexKey)
	if err == nil {
		keyID, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	key.ID = uint64(keyID)
	err = pstorage.PutJSON(userKV, s.actorStr(actor), strconv.FormatInt(keyID, 10), key)
	if err != nil {
		return nil, err
	}

	// Increment the start index to ensure sequential SSHKey IDs
	err = userKV.Put(s.actorStr(actor), sshKeysCurrentIndexKey, []byte(strconv.FormatInt(keyID+1, 10)))
	if err != nil {
		return nil, err
	}

	// Add the key to the lookup_user index.
	if err := s.addLookupIndex(ctx, key.Key, actor.UID); err != nil {
		return nil, err
	}

	return &pbtypes.Void{}, nil
}

// LookupUser looks up user by key. The returned UserSpec will only have UID field set.
func (s *userKeys) LookupUser(ctx context.Context, key *sourcegraph.SSHPublicKey) (*sourcegraph.UserSpec, error) {
	defer noCache(ctx)

	userKV := pstorage.Namespace(ctx, sshKeysAppName, "")
	var keysToUID map[string]int32
	keyHash := publicKeyToHash(key.Key)
	err := pstorage.GetJSON(userKV, sshKeysLookupUserBucket, keyHash, &keysToUID)
	if err != nil {
		return nil, err
	}
	uid, ok := keysToUID[base64.RawURLEncoding.EncodeToString(key.Key)]
	if !ok {
		return nil, errors.New("no such public key")
	}
	return &sourcegraph.UserSpec{UID: int32(uid)}, nil
}

func (s *userKeys) ListKeys(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.SSHKeyList, error) {
	defer noCache(ctx)
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	userKV := pstorage.Namespace(ctx, sshKeysAppName, "")
	keys, err := userKV.List(s.actorStr(actor))
	if err != nil {
		return nil, err
	}

	if len(keys) < 2 {
		return &sourcegraph.SSHKeyList{}, nil
	}

	sshKeyList := make([]sourcegraph.SSHPublicKey, len(keys)-1)
	for x, key := range keys {
		if key == sshKeysCurrentIndexKey {
			continue
		}

		sshKey, err := s.getSSHKey(userKV, actor, key)
		if err != nil {
			return nil, err
		}
		sshKeyList[x] = *sshKey
	}
	return &sourcegraph.SSHKeyList{SSHKeys: sshKeyList}, nil
}

func (s *userKeys) getSSHKey(userKV pstorage.System, actor authpkg.Actor, key string) (*sourcegraph.SSHPublicKey, error) {
	var data = struct {
		Key, Name string
		ID        uint64
	}{}

	if err := pstorage.GetJSON(userKV, s.actorStr(actor), key, &data); err != nil {
		return nil, err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(data.Key)
	if err != nil {
		return nil, err
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyBytes))
	if err != nil {
		return nil, err
	}

	return &sourcegraph.SSHPublicKey{
		Name: data.Name,
		ID:   data.ID,
		Key:  ssh.MarshalAuthorizedKey(pubKey),
	}, nil
}

func (s *userKeys) DeleteAllKeys(ctx context.Context, _ *pbtypes.Void) (*pbtypes.Void, error) {
	defer noCache(ctx)
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	userKV := pstorage.Namespace(ctx, sshKeysAppName, "")

	// Delete each key.
	keys, err := userKV.List(s.actorStr(actor))
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if key == sshKeysCurrentIndexKey {
			continue
		}
		keyID, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			return nil, err
		}
		_, err = s.DeleteKey(ctx, &sourcegraph.SSHPublicKey{
			ID: keyID,
		})
		if err != nil {
			return nil, err
		}
	}
	return &pbtypes.Void{}, nil
}

func (s *userKeys) DeleteKey(ctx context.Context, key *sourcegraph.SSHPublicKey) (*pbtypes.Void, error) {
	defer noCache(ctx)
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	if key.Name != "" {
		// List the keys to find the ID.
		//
		// TODO(slimsag): implement this more efficiently -- not super important
		// because users are not expected to have many SSH keys.
		list, err := s.ListKeys(ctx, &pbtypes.Void{})
		if err != nil {
			return nil, err
		}
		found := false
		for _, listedKey := range list.SSHKeys {
			if listedKey.Name == key.Name {
				key.ID = listedKey.ID
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("no such key with name %q", key.Name)
		}
	}

	userKV := pstorage.Namespace(ctx, sshKeysAppName, "")
	storageKey := strconv.FormatInt(int64(key.ID), 10)

	// Remove the key from the lookup_user index.
	fullKey, err := s.getSSHKey(userKV, actor, storageKey)
	if err != nil {
		return nil, err
	}
	if err := s.removeLookupIndex(ctx, fullKey.Key); err != nil {
		return nil, err
	}

	err = userKV.Delete(s.actorStr(actor), storageKey)
	return &pbtypes.Void{}, err
}

type keyToUID map[string]int32

func (s *userKeys) addLookupIndex(ctx context.Context, key []byte, uid int) error {
	// Marshal key into network format.
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		return err
	}
	key = pubKey.Marshal()

	userKV := pstorage.Namespace(ctx, sshKeysAppName, "")
	keysToUID := make(map[string]int32)
	keyHash := publicKeyToHash(key)
	err = pstorage.GetJSON(userKV, sshKeysLookupUserBucket, keyHash, &keysToUID)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	keysToUID[base64.RawURLEncoding.EncodeToString(key)] = int32(uid)
	return pstorage.PutJSON(userKV, sshKeysLookupUserBucket, keyHash, keysToUID)
}

func (s *userKeys) removeLookupIndex(ctx context.Context, key []byte) error {
	// Marshal key into network format.
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		return err
	}
	key = pubKey.Marshal()

	userKV := pstorage.Namespace(ctx, sshKeysAppName, "")
	var keysToUID map[string]int32
	keyHash := publicKeyToHash(key)
	err = pstorage.GetJSON(userKV, sshKeysLookupUserBucket, keyHash, &keysToUID)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	delete(keysToUID, base64.RawURLEncoding.EncodeToString(key))
	return pstorage.PutJSON(userKV, sshKeysLookupUserBucket, keyHash, keysToUID)
}

// actorStr returns actor.UID as a string.
func (s *userKeys) actorStr(actor authpkg.Actor) string {
	return strconv.Itoa(int(actor.UID))
}

func publicKeyToHash(key []byte) string {
	sum := sha1.Sum(key)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
