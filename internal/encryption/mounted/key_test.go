package mounted

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRoundTrip(t *testing.T) {
	testKeyFile := filepath.Join(t.TempDir(), "testroundtrip_testkey_file")
	secret := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, secret)
	require.NoError(t, err)

	var tcs = []struct {
		name   string
		config schema.MountedEncryptionKey
		setup  func(t *testing.T)
	}{
		{
			name: "env_var",
			config: schema.MountedEncryptionKey{
				Type:       "mounted",
				Keyname:    "testkey/env_var",
				EnvVarName: "testroundtrip_testkey_env_var",
			},
			setup: func(t *testing.T) {
				t.Setenv("testroundtrip_testkey_env_var", string(secret))
			},
		},
		{
			name: "file_name",
			config: schema.MountedEncryptionKey{
				Type:     "mounted",
				Keyname:  "testkey/env_var",
				Filepath: testKeyFile,
			},
			setup: func(t *testing.T) {
				require.NoError(t, os.WriteFile(testKeyFile, secret, 0644))
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t)
			}

			ctx := context.Background()
			k, err := NewKey(ctx, tc.config)
			require.NoError(t, err)

			var plaintext = theScriptOfBeeMovie

			ciphertext, err := k.Encrypt(ctx, []byte(plaintext))
			require.NoError(t, err)

			assert.NotEqual(t, string(ciphertext), plaintext)

			s, err := k.Decrypt(ctx, ciphertext)
			require.NoError(t, err)
			assert.Equal(t, plaintext, s.Secret())

			// Now verify that the old encryption method we used to use still works.
			oldEncryptedValue, err := oldEncrypt(ctx, tc.config.Keyname, secret, []byte(plaintext))
			require.NoError(t, err)

			s, err = k.Decrypt(ctx, oldEncryptedValue)
			require.NoError(t, err)
			assert.Equal(t, plaintext, s.Secret())
		})
	}
}

func oldEncrypt(ctx context.Context, keyname string, secret, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, errors.Wrap(err, "creating AES cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "creating GCM block cipher")
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	out := encryptedValue{
		KeyName:    keyname,
		Ciphertext: ciphertext,
		Checksum:   crc32Sum(plaintext),
	}
	jsonKey, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	buf := base64.StdEncoding.EncodeToString(jsonKey)
	return []byte(buf), err
}

type encryptedValue struct {
	KeyName    string
	Ciphertext []byte
	Checksum   uint32
}

func crc32Sum(data []byte) uint32 {
	t := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(data, t)
}

var theScriptOfBeeMovie = `According to all known laws
	of aviation,
	there is no way a bee
	should be able to fly.
	Its wings are too small to get
	its fat little body off the ground.
	The bee, of course, flies anyway
	because bees don't care
	what humans think is impossible.
	Yellow, black. Yellow, black.
	Yellow, black. Yellow, black.
	Ooh, black and yellow!
	Let's shake it up a little.
	Barry! Breakfast is ready!
	Coming!
	Hang on a second.
	Hello?
	- Barry?
	- Adam?
	- Oan you believe this is happening?
	- I can't. I'll pick you up.
	Looking sharp.
	Use the stairs. Your father
	paid good money for those.
	Sorry. I'm excited.
	Here's the graduate.
	We're very proud of you, son.
	A perfect report card, all B's.
	Very proud.
	Ma! I got a thing going here.
	- You got lint on your fuzz.
	- Ow! That's me!
	- Wave to us! We'll be in row 118,000.
	- Bye!
	Barry, I told you,
	stop flying in the house!
	- Hey, Adam.
	- Hey, Barry.
	- Is that fuzz gel?
	- A little. Special day, graduation.
	Never thought I'd make it.
	Three days grade school,
	three days high school.
	Those were awkward.
	Three days college. I'm glad I took
	a day and hitchhiked around the hive.
	You did come back different.
	- Hi, Barry.
	- Artie, growing a mustache? Looks good.
	- Hear about Frankie?
	- Yeah.
	- You going to the funeral?
	- No, I'm not going.
	Everybody knows,
	sting someone, you die.
	Don't waste it on a squirrel.
	Such a hothead.
	I guess he could have
	just gotten out of the way.
	I love this incorporating
	an amusement park into our day.
	That's why we don't need vacations.
	Boy, quite a bit of pomp...
	under the circumstances.
	- Well, Adam, today we are men.
	- We are!
	- Bee-men.
	- Amen!
	Hallelujah!
	Students, faculty, distinguished bees,
	please welcome Dean Buzzwell.
	Welcome, New Hive City
	graduating class of...
	...9:15.
	That concludes our ceremonies.
	And begins your career
	at Honex Industries!
	Will we pick ourjob today?
	I heard it's just orientation.
	Heads up! Here we go.
	Keep your hands and antennas
	inside the tram at all times.
	- Wonder what it'll be like?
	- A little scary.
	Welcome to Honex,
	a division of Honesco
	and a part of the Hexagon Group.
	This is it!
	Wow.
	Wow.
	We know that you, as a bee,
	have worked your whole life
	to get to the point where you
	can work for your whole life.
	Honey begins when our valiant Pollen
	Jocks bring the nectar to the hive.
	Our top-secret formula
	is automatically color-corrected,
	scent-adjusted and bubble-contoured
	into this soothing sweet syrup
	with its distinctive
	golden glow you know as...
	Honey!
	- That girl was hot.
	- She's my cousin!
	- She is?
	- Yes, we're all cousins.
	- Right. You're right.
	- At Honex, we constantly strive
	to improve every aspect
	of bee existence.`
