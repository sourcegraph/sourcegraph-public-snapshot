package spec

import (
	"fmt"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func errorMessages(errs []error) []string {
	var messages []string
	for _, e := range errs {
		messages = append(messages, e.Error())
	}
	return messages
}

func TestEnvironmentResourcePostgreSQLSpecValidate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		spec       *EnvironmentResourcePostgreSQLSpec
		wantErrors autogold.Value
	}{
		{
			name:       "nil",
			spec:       nil,
			wantErrors: nil,
		},
		{
			name:       "defaults",
			spec:       &EnvironmentResourcePostgreSQLSpec{},
			wantErrors: nil,
		},
		{
			name: "odd CPU",
			spec: &EnvironmentResourcePostgreSQLSpec{
				CPU: pointers.Ptr(3),
			},
			wantErrors: autogold.Expect([]string{"postgreSQL.cpu must be 1 or a multiple of 2"}),
		},
		{
			name: "too little memory for CPU",
			spec: &EnvironmentResourcePostgreSQLSpec{
				CPU:      pointers.Ptr(12),
				MemoryGB: pointers.Ptr(4),
			},
			wantErrors: autogold.Expect([]string{"postgreSQL.memoryGB must be >= postgreSQL.cpu"}),
		},
		{
			name: "too much memory for CPU",
			spec: &EnvironmentResourcePostgreSQLSpec{
				MemoryGB: pointers.Ptr(12),
			},
			wantErrors: autogold.Expect([]string{"postgreSQL.memoryGB must be <= 6*postgreSQL.cpu"}),
		},
		{
			name: "odd CPU, too much memory for CPU",
			spec: &EnvironmentResourcePostgreSQLSpec{
				CPU:      pointers.Ptr(5),
				MemoryGB: pointers.Ptr(50),
			},
			wantErrors: autogold.Expect([]string{
				"postgreSQL.cpu must be 1 or a multiple of 2",
				"postgreSQL.memoryGB must be <= 6*postgreSQL.cpu",
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.spec.Validate()
			if tc.wantErrors == nil {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
				tc.wantErrors.Equal(t, errorMessages(errs))
			}
		})
	}
}

func TestEnvironmentInstancesResourcesSpecValdiate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		spec       *EnvironmentInstancesResourcesSpec
		wantErrors autogold.Value
	}{
		{
			name:       "nil",
			spec:       nil,
			wantErrors: nil,
		},
		{
			name: "ok 1Gi",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    1,
				Memory: "1Gi",
			},
			wantErrors: nil,
		},
		{
			name: "ok 512Mi",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    1,
				Memory: "512Mi",
			},
			wantErrors: nil,
		},
		{
			name: "cpu, memory too low",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    0,
				Memory: "256Mi",
			},
			wantErrors: autogold.Expect([]string{"resources.cpu must be >= 1", "resources.memory must be >= 512Mi"}),
		},
		{
			name: "cpu, memory too high",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    10,
				Memory: "60Gi",
			},
			wantErrors: autogold.Expect([]string{
				"resources.cpu > 8 not supported - consider decreasing scaling.maxRequestConcurrency and increasing scaling.maxCount instead",
				"resources.memory > 32Gi not supported - consider decreasing scaling.maxRequestConcurrency and increasing scaling.maxCount instead",
			}),
		},
		{
			name: "cpu too high for memory",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    8,
				Memory: "1Gi",
			},
			wantErrors: autogold.Expect([]string{"resources.cpu > 6 requires resources.memory >= 4Gi"}),
		},
		{
			name: "memory too high for cpu",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    1,
				Memory: "32Gi",
			},
			wantErrors: autogold.Expect([]string{"resources.memory > 24Gi requires resources.cpu >= 8"}),
		},
		{
			name: "invalid memory unit",
			spec: &EnvironmentInstancesResourcesSpec{
				CPU:    1,
				Memory: "8GiB",
			},
			wantErrors: autogold.Expect([]string{"resources.memory is invalid: units: unknown unit GiB in 8GiB"}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.spec.Validate()
			if tc.wantErrors == nil {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
				tc.wantErrors.Equal(t, errorMessages(errs))
			}
		})
	}
}

func TestEnvironmentJobScheduleSpecFindMaxCronInterval(t *testing.T) {
	for _, tc := range []struct {
		name string
		spec EnvironmentJobScheduleSpec

		wantInterval time.Duration
		wantError    autogold.Value
	}{
		{
			name: "typical interval",
			spec: EnvironmentJobScheduleSpec{
				Cron: "0 * * * *",
			},
			wantInterval: time.Hour,
		},
		{
			name: "interval too small",
			spec: EnvironmentJobScheduleSpec{
				Cron: "* * * * *",
			},
			wantError: autogold.Expect("the longest interval must be >15m, got 1m0s"),
		},
		{
			name: "hourly and weekends off",
			spec: EnvironmentJobScheduleSpec{
				Cron: "0 * * * 1-5",
			},
			wantInterval: 49 * time.Hour, // 2 weekend days
		},
		{
			name: "monthly interval forbidden",
			spec: EnvironmentJobScheduleSpec{
				Cron: "0 0 1 * *", // once per month
			},
			wantError: autogold.Expect("the longest interval must be <8 days, got 744h0m0s"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			now := time.Now()
			interval, err := tc.spec.FindMaxCronInterval(now)
			if tc.wantError != nil {
				assert.Error(t, err)
				tc.wantError.Equal(t, err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantInterval, *interval)

			if tc.wantError != nil {
				return
			}

			// Make sure at various times from now, the interval doesn't change.
			for _, fromNow := range []time.Duration{
				12 * time.Hour,
				48 * time.Hour,
				64 * time.Hour,
				128 * time.Hour,
				256 * time.Hour,
			} {
				t.Run(fmt.Sprintf("%s from now", fromNow.String()), func(t *testing.T) {
					spec := tc.spec
					newInterval, err := spec.FindMaxCronInterval(now.Add(fromNow))
					require.NoError(t, err)
					assert.Equal(t, interval.String(), newInterval.String())
				})
			}
		})
	}
}

func TestEnvironmentDeploySpec_Validate(t *testing.T) {
	tests := []struct {
		name     string
		spec     EnvironmentDeploySpec
		wantErrs autogold.Value
	}{
		{
			name: "manual type with subscription",
			spec: EnvironmentDeploySpec{
				Type:         EnvironmentDeployTypeManual,
				Subscription: &EnvironmentDeployTypeSubscriptionSpec{},
			},
			wantErrs: autogold.Expect([]string{"subscription deploy spec provided when type is manual"}),
		},
		{
			name: "subscription type with manual",
			spec: EnvironmentDeploySpec{
				Type:   EnvironmentDeployTypeSubscription,
				Manual: &EnvironmentDeployManualSpec{},
			},
			wantErrs: autogold.Expect([]string{"manual deploy spec provided when type is subscription"}),
		},
		{
			name: "subscription type without subscription",
			spec: EnvironmentDeploySpec{
				Type: EnvironmentDeployTypeSubscription,
			},
			wantErrs: autogold.Expect([]string{"no subscription specified when deploy type is subscription"}),
		},
		{
			name: "subscription type with empty tag",
			spec: EnvironmentDeploySpec{
				Type:         EnvironmentDeployTypeSubscription,
				Subscription: &EnvironmentDeployTypeSubscriptionSpec{},
			},
			wantErrs: autogold.Expect([]string{"no tag in image subscription specified"}),
		},
		{
			name: "subscription type with tag",
			spec: EnvironmentDeploySpec{
				Type: EnvironmentDeployTypeSubscription,
				Subscription: &EnvironmentDeployTypeSubscriptionSpec{
					Tag: "insiders",
				},
			},
		},
		{
			name: "rollout type",
			spec: EnvironmentDeploySpec{
				Type: EnvironmentDeployTypeRollout,
			},
		},
		{
			name: "invalid type",
			spec: EnvironmentDeploySpec{
				Type: "invalid",
			},
			wantErrs: autogold.Expect([]string{`invalid deploy type "invalid"`}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := stringifyErrors(tc.spec.Validate())
			if tc.wantErrs == nil {
				assert.Empty(t, errs)
			} else {
				tc.wantErrs.Equal(t, errs)
			}
		})
	}
}

func stringifyErrors(errs []error) (values []string) {
	for _, errs := range errs {
		values = append(values, errs.Error())
	}
	return values
}
