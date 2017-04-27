import { PhabricatorInstance } from "../../app/utils/classes";

const sgdevPhabricatorRepoMap = {
	"nzap": "gitolite.aws.sgdev.org/uber-go/zap.git",
	"zap": "gitolite.aws.sgdev.org/uber-go/zap.git",
	"nmux": "gitolite.aws.sgdev.org/mux.git",
	"nbroke": "gitolite.aws.sgdev.org/sourcegraph/broken-test",
	"njoda": "gitolite.aws.sgdev.org/JodaOrg/joda-time",
	"nangular": "gitolite.aws.sgdev.org/angular/angular",
	"nrides": "gitolite.aws.sgdev.org/uber/rides-java-sdk",
};

export const SGDEV_SOURCEGRAPH_URL = "http://node.aws.sgdev.org:30000";

export const sgDevPhabricatorInstance = new PhabricatorInstance(sgdevPhabricatorRepoMap, "sgdev_phabricator");
