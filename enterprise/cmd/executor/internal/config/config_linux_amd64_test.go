package config_test

func TestConfig_Validate_Firecracker(t *testing.T) {
	tests := []struct {
		name        string
		getterFunc  env.GetterFunc
		expectedErr error
	}{
		{
			name: "Valid config",
			getterFunc: func(name string, defaultValue, description string) string {
				switch name {
				case "EXECUTOR_QUEUE_NAME":
					return "batches"
				case "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				case "EXECUTOR_FRONTEND_PASSWORD":
					return "some-password"
				case "EXECUTOR_USE_FIRECRACKER":
					return "true"
				default:
					return defaultValue
				}
			},
		},
		{
			name: "Single cpu per job",
			getterFunc: func(name string, defaultValue, description string) string {
				switch name {
				case "EXECUTOR_QUEUE_NAME":
					return "batches"
				case "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				case "EXECUTOR_FRONTEND_PASSWORD":
					return "some-password"
				case "EXECUTOR_USE_FIRECRACKER":
					return "true"
				case "EXECUTOR_JOB_NUM_CPUS":
					return "1"
				default:
					return defaultValue
				}
			},
		},
		{
			name: "Odd number of cpus per job",
			getterFunc: func(name string, defaultValue, description string) string {
				switch name {
				case "EXECUTOR_QUEUE_NAME":
					return "batches"
				case "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				case "EXECUTOR_FRONTEND_PASSWORD":
					return "some-password"
				case "EXECUTOR_USE_FIRECRACKER":
					return "true"
				case "EXECUTOR_JOB_NUM_CPUS":
					return "5"
				default:
					return defaultValue
				}
			},
			expectedErr: errors.New("EXECUTOR_JOB_NUM_CPUS must be 1 or an even number"),
		},
		{
			name: "Invalid disk size",
			getterFunc: func(name string, defaultValue, description string) string {
				switch name {
				case "EXECUTOR_QUEUE_NAME":
					return "batches"
				case "EXECUTOR_FRONTEND_URL":
					return "http://some-url.com"
				case "EXECUTOR_FRONTEND_PASSWORD":
					return "some-password"
				case "EXECUTOR_USE_FIRECRACKER":
					return "true"
				case "EXECUTOR_FIRECRACKER_DISK_SPACE":
					return "12q"
				default:
					return defaultValue
				}
			},
			expectedErr: errors.New("invalid disk size provided for EXECUTOR_FIRECRACKER_DISK_SPACE: 12q"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.SetMockGetter(test.getterFunc)
			cfg.Load()

			err := cfg.Validate()
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
