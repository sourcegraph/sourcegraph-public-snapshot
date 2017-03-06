import { PhabricatorInstance } from "../../app/utils/classes";

const umamiPhabricatorRepoMap = {
	"moiosa": "mobile/ios",
	"inbox": "infra/boxer",
	"wewea": "weaver/weaver",
	"wesuperfinem": "web/superfine-monorepo",
	"weent": "web/entities-monorepo",
	"odche": "odp/cherami",
	"decer": "devexp/cerberus",
	"deuberfxj": "devexp/uberfx-java",
	"gocom": "go-common",
	"rtfilt": "rt/filter-go",
	"dagor": "data/gorilla-websocket",
	"instap": "infra/stapi-go",
	"enwonkago": "engsec/wonka-go",
	"engal": "engsec/galileo-go",
	"enpoli": "engsec/polizei",
	"frfraudg": "fraud/fraud-go-common",
	"couberloc": "communications/uberlocales",
	"rprpc": "rpc/rpcinit",
	"inunsa": "infra/uns",
	"odcheram": "odp/cherami-client-go",
	"dacur": "data/curator",
	"rtfliprj": "rt/flipr-java-client",
	"deubercom": "devexp/uber-common-configuration",
	"inuconfi": "infra/uconfig-client-java",
};

export const UMAMI_SOURCEGRAPH_URL = "https://sourcegraph.sgpxy.dev.uberinternal.com";
const PHABRICATOR_STAGING_URI = "phabricator-staging";

export const umamiPhabricatorInstance = new PhabricatorInstance(umamiPhabricatorRepoMap, PHABRICATOR_STAGING_URI);
