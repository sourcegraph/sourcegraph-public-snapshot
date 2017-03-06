import { PhabricatorInstance } from "../../app/utils/classes";

const sgdevPhabricatorRepoMap = {
	"thyme": "sourcegraph-thyme",
	"uzap": "uber-zap",
	"uberzap": "uber-zap",
	"kubernetes": "kubernetes",
	"kube": "kubernetes",
	"jaeger": "uber-jaeger",
	"angular": "angular",
	"joda": "joda-time",
	"jodatime": "joda-time",
	"jaegerjava": "uber-jaeger-java",
	"spray": "uber-kafka-spraynozzle",
	"kafkaspraynozzle": "uber-kafka-spraynozzle",
	"jrides": "uber-rides-java-sdk",
	"uberridesjava": "uber-rides-java-sdk",
	"krest": "uber-kafka-rest",
	"kafkarest": "uber-kafka-rest",
	"broke": "broken-test",
};

export const SGDEV_SOURCEGRAPH_URL = "http://node.aws.sgdev.org:30000";
const PHABRICATOR_STAGING_URI = "phabricator-staging";

export const sgDevPhabricatorInstance = new PhabricatorInstance(sgdevPhabricatorRepoMap, PHABRICATOR_STAGING_URI);
