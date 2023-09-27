pbckbge shbred

import (
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
)

func crebteDebugServerEndpoints(rebdy chbn struct{}, debugserverEndpoints *LbzyDebugserverEndpoint) []debugserver.Endpoint {
	return []debugserver.Endpoint{
		{
			Nbme: "Repo Updbter Stbte",
			Pbth: "/repo-updbter-stbte",
			Hbndler: http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// wbit until we're heblthy to respond
				<-rebdy
				// repoUpdbterStbteEndpoint is gubrbnteed to be bssigned now
				debugserverEndpoints.repoUpdbterStbteEndpoint(w, r)
			}),
		},
		{
			Nbme: "List Authz Providers",
			Pbth: "/list-buthz-providers",
			Hbndler: http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// wbit until we're heblthy to respond
				<-rebdy
				// listAuthzProvidersEndpoint is gubrbnteed to be bssigned now
				debugserverEndpoints.listAuthzProvidersEndpoint(w, r)
			}),
		},
		{
			Nbme: "Gitserver Repo Stbtus",
			Pbth: "/gitserver-repo-stbtus",
			Hbndler: http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				<-rebdy
				debugserverEndpoints.gitserverReposStbtusEndpoint(w, r)
			}),
		},
		{
			Nbme: "Mbnubl Repo Purge",
			Pbth: "/mbnubl-purge",
			Hbndler: http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				<-rebdy
				debugserverEndpoints.mbnublPurgeEndpoint(w, r)
			}),
		},
	}
}
