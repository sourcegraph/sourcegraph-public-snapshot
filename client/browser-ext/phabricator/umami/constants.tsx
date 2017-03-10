import { PhabricatorInstance } from "../../app/utils/classes";

const umamiPhabricatorRepoMap = {
	"moiosa": "code.uber.internal/mobile/ios",
	"inbox": "code.uber.internal/infra/boxer",
	"wewea": "code.uber.internal/weaver/weaver",
	"wesuperfinem": "code.uber.internal/web/superfine-monorepo",
	"weent": "code.uber.internal/web/entities-monorepo",
	"odche": "code.uber.internal/odp/cherami",
	"decer": "code.uber.internal/devexp/cerberus",
	"deuberfxj": "code.uber.internal/devexp/uberfx-java",
	"gocom": "code.uber.internal/go-common",
	"rtfilt": "code.uber.internal/rt/filter-go",
	"dagor": "code.uber.internal/data/gorilla-websocket",
	"instap": "code.uber.internal/infra/stapi-go",
	"enwonkago": "code.uber.internal/engsec/wonka-go",
	"engal": "code.uber.internal/engsec/galileo-go",
	"enpoli": "code.uber.internal/engsec/polizei",
	"frfraudg": "code.uber.internal/fraud/fraud-go-common",
	"couberloc": "code.uber.internal/communications/uberlocales",
	"rprpc": "code.uber.internal/rpc/rpcinit",
	"inunsa": "code.uber.internal/infra/uns",
	"odcheram": "code.uber.internal/odp/cherami-client-go",
	"dacur": "code.uber.internal/data/curator",
	"rtfliprj": "code.uber.internal/rt/flipr-java-client",
	"deubercom": "code.uber.internal/devexp/uber-common-configuration",
	"inuconfi": "code.uber.internal/infra/uconfig-client-java",
};

export const UMAMI_SOURCEGRAPH_URL = "https://sourcegraph.sgpxy.dev.uberinternal.com";
const PHABRICATOR_STAGING_URI = "phabricator-staging";

export const umamiPhabricatorInstance = new PhabricatorInstance(umamiPhabricatorRepoMap, PHABRICATOR_STAGING_URI);
