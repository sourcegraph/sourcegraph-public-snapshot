package main

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/ini.v1"
)

func iniMustLoad(t *testing.T, source interface{}) *ini.File {
	f, err := getGrafanaConfig(source)
	if err != nil {
		t.Fatal(err.Error())
	}
	return f
}

func Test_grafanaChangeSMTP(t *testing.T) {
	type args struct {
		grafana   GrafanaContext
		newConfig *subscribedSiteConfig
	}
	tests := []struct {
		name       string
		args       args
		wantResult GrafanaChangeResult
		wantConfig string
	}{
		{
			name: "add new SMTP section",
			args: args{
				grafana: GrafanaContext{
					Client: nil, // should not require client
					Config: iniMustLoad(t, []byte(``)),
				},
				newConfig: &subscribedSiteConfig{
					Email: &siteEmailConfig{
						Address: "robert@bobheadxi.dev",
						SMTP: &schema.SMTPServerConfig{
							Host: "localhost",
							Port: 25,
						},
					},
				},
			},
			wantResult: GrafanaChangeResult{
				Problems:     nil,
				ConfigChange: true,
			},
			wantConfig: `[smtp]
enabled      = true
host         = localhost:25
from_address = robert@bobheadxi.dev
from_name    = Sourcegraph Grafana`,
		},
		{
			name: "remove SMTP fields when receiving a nil update",
			args: args{
				grafana: GrafanaContext{
					Client: nil,
					Config: iniMustLoad(t, []byte(`[smtp]
enabled = true
host    = localhost:25`)),
				},
				newConfig: &subscribedSiteConfig{},
			},
			wantResult: GrafanaChangeResult{
				Problems:     nil,
				ConfigChange: true,
			},
			wantConfig: `[smtp]
enabled = false`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := grafanaChangeSMTP(context.Background(), log15.New("test", t.Name()), tt.args.grafana, tt.args.newConfig); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("grafanaChangeSMTP() = %v, want %v", gotResult, tt.wantResult)
			}
			var out bytes.Buffer
			tt.args.grafana.Config.WriteTo(&out)
			gotConfig := strings.Trim(out.String(), "\n")
			if gotConfig != tt.wantConfig {
				t.Errorf(cmp.Diff(gotConfig, tt.wantConfig))
			}
		})
	}
}
