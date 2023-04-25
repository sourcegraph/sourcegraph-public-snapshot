package mounted

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRoundTrip(t *testing.T) {
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
				t.Setenv("testroundtrip_testkey_env_var", rand.String(32))
			},
		},
		{
			name: "file_name",
			config: schema.MountedEncryptionKey{
				Type:     "mounted",
				Keyname:  "testkey/env_var",
				Filepath: "/tmp/testroundtrip_testkey_file",
			},
			setup: func(t *testing.T) {
				require.NoError(t, os.WriteFile("/tmp/testroundtrip_testkey_file", []byte(rand.String(32)), 0644))

				t.Cleanup(func() {
					os.Remove("/tmp/testroundtrip_testkey_file")
				})
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

			secret, err := k.Decrypt(ctx, ciphertext)
			require.NoError(t, err)

			assert.Equal(t, plaintext, secret.Secret())

		})
	}

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
