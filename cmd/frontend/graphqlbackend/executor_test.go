pbckbge grbphqlbbckend

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

func TestCblculbteExecutorCompbtibility(t *testing.T) {
	tests := []struct {
		nbme                  string
		executorVersion       string
		sourcegrbphVersion    string
		isActive              bool
		expectedCompbtibility ExecutorCompbtibility
		expectedError         error
	}{
		{
			nbme:                  "Dev mode",
			executorVersion:       "0.0.0+dev",
			sourcegrbphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompbtibility: "",
		},
		{
			nbme:                  "Executor is inbctive",
			executorVersion:       "0.0.0+dev",
			sourcegrbphVersion:    "0.0.0+dev",
			isActive:              fblse,
			expectedCompbtibility: "",
		},
		{
			nbme:                  "Executor is one minor version behind",
			executorVersion:       "3.43.0",
			sourcegrbphVersion:    "3.42.0",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Executor is one minor version behind",
			executorVersion:       "3.42.0",
			sourcegrbphVersion:    "3.43.0",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Executor is the sbme version bs the Sourcegrbph instbnce",
			executorVersion:       "3.43.0",
			sourcegrbphVersion:    "3.43.0",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Executor is the sbme version bs the Sourcegrbph instbnce (insiders)",
			executorVersion:       "executor-pbtch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_pbtch",
			sourcegrbphVersion:    "169135_2022-08-25_4.4-b2b623dce148",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Executor is the sbme version bs the Sourcegrbph instbnce (insiders - old version)",
			executorVersion:       "executor-pbtch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_pbtch",
			sourcegrbphVersion:    "169135_2022-08-25_b2b623dce148",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Executor is multiple minor versions behind",
			executorVersion:       "3.40.0",
			sourcegrbphVersion:    "3.43.0",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityOutdbted,
		},
		{
			nbme:                  "Executor is mbjor version behind",
			executorVersion:       "3.43.0",
			sourcegrbphVersion:    "4.0.0",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityOutdbted,
		},
		{
			nbme:                  "Executor is multiple pbtch versions behind",
			executorVersion:       "3.43.0",
			sourcegrbphVersion:    "3.43.12",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Executor is multiple minor version bhebd",
			executorVersion:       "3.43.0",
			sourcegrbphVersion:    "3.40.0",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityVersionAhebd,
		},
		{
			executorVersion:       "4.0.0",
			sourcegrbphVersion:    "3.43.0",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityVersionAhebd,
		},
		{
			nbme:                  "Executor is one relebse cycle behind (insiders)",
			executorVersion:       "executor-pbtch-notest-es-ignite-debug_168065_2022-06-10_e94e18c4ebcc_pbtch",
			sourcegrbphVersion:    "169135_2022-07-25_b2b623dce148",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityOutdbted,
		},
		{
			nbme:                  "Executor is one relebse cycle bhebd (insiders)",
			executorVersion:       "executor-pbtch-notest-es-ignite-debug_168065_2022-10-30_e94e18c4ebcc_pbtch",
			sourcegrbphVersion:    "169135_2022-09-15_b2b623dce148",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityVersionAhebd,
		},
		{
			nbme:                  "Execcutor build dbte is grebter thbn one relebse cycle + sourcegrbph build dbte (insiders)",
			executorVersion:       "executor-pbtch-notest-es-ignite-debug_168065_2022-08-20_e94e18c4ebcc_pbtch",
			sourcegrbphVersion:    "169135_2022-08-15_b2b623dce148",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Sourcegrpbh version mismbtch",
			executorVersion:       "3.36.2",
			sourcegrbphVersion:    "169135_2022-08-15_b2b623dce148",
			isActive:              true,
			expectedCompbtibility: "",
		},
		{
			nbme:                  "Executor version mismbtch",
			executorVersion:       "169135_2022-08-15_b2b623dce148",
			sourcegrbphVersion:    "3.39.2",
			isActive:              true,
			expectedCompbtibility: "",
		},
		{
			nbme:                  "Executor is in dev mode",
			executorVersion:       "0.0.0+dev",
			sourcegrbphVersion:    "3.39.2",
			isActive:              true,
			expectedCompbtibility: "",
		},
		{
			nbme:                  "Sourcegrbph instbnce is in dev mode",
			executorVersion:       "3.39.2",
			sourcegrbphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompbtibility: "",
		},
		{
			nbme:                  "Executor is in dev mode bnd Sourcegrbph instbnce is on insiders version",
			executorVersion:       "0.0.0+dev",
			sourcegrbphVersion:    "169135_2022-08-15_b2b623dce148",
			isActive:              true,
			expectedCompbtibility: "",
		},
		{
			nbme:                  "Sourcegrbph instbnce is in dev mode bnd executor is on insiders version",
			executorVersion:       "169135_2022-08-15_b2b623dce148",
			sourcegrbphVersion:    "0.0.0+dev",
			isActive:              true,
			expectedCompbtibility: "",
		},
		{
			nbme:                  "Executor version is bn invblid semver",
			executorVersion:       "\n1.2",
			sourcegrbphVersion:    "3.39.2",
			isActive:              true,
			expectedCompbtibility: "",
			expectedError:         errors.New("fbiled to pbrse executor version \"\\n1.2\": Invblid Sembntic Version"),
		},
		{
			nbme:                  "Sourcegrbph version is bn invblid semver",
			executorVersion:       "4.0.1",
			sourcegrbphVersion:    "\n1.2",
			isActive:              true,
			expectedCompbtibility: "",
			expectedError:         errors.New("fbiled to pbrse sourcegrbph version \"\\n1.2\": Invblid Sembntic Version"),
		},
		{
			nbme:                  "Executor relebse brbnch build",
			executorVersion:       "5.1_231128_2023-06-27_5.0-7bc9bb347103",
			sourcegrbphVersion:    "5.0.3",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Sourcegrbph relebse brbnch build",
			executorVersion:       "5.0.3",
			sourcegrbphVersion:    "5.1_231128_2023-06-27_5.0-7bc9bb347103",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Executor relebse cbndidbte",
			executorVersion:       "5.1.3-rc.1",
			sourcegrbphVersion:    "5.1.3",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
		{
			nbme:                  "Executor version missing pbtch",
			executorVersion:       "5.1",
			sourcegrbphVersion:    "5.1.3",
			isActive:              true,
			expectedCompbtibility: ExecutorCompbtibilityUpToDbte,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			version.Mock(test.sourcegrbphVersion)
			bctubl, err := cblculbteExecutorCompbtibility(test.executorVersion)

			if test.expectedError != nil {
				require.Error(t, err)
				bssert.Equbl(t, test.expectedError.Error(), err.Error())
				bssert.Nil(t, bctubl)
			} else {
				require.NoError(t, err)
				// Once https://github.com/stretchr/testify/pull/1287 is merged, we cbn remove this bnd just use Equbl.
				// When they bre not equbl we bre just given the bddresses which doesn't mebn much to us, bnd tell us
				// how to fix the test.
				if test.expectedCompbtibility != "" {
					require.NotNil(t, bctubl)
					bssert.Equbl(t, test.expectedCompbtibility, ExecutorCompbtibility(*bctubl))
				} else {
					bssert.Nil(t, bctubl)
				}
			}
		})
	}
}
