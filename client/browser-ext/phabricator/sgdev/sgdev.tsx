/**
 * This is done before all other imports to ensure that the event logger is set ahead of time.
 */
import { setEventLogger, setPhabricatorInstance, setSourcegraphUrl } from "../../app/utils/context";
import { SGDEV_SOURCEGRAPH_URL, sgDevPhabricatorInstance } from "./constants";
setSourcegraphUrl(SGDEV_SOURCEGRAPH_URL);
setPhabricatorInstance(sgDevPhabricatorInstance);

import { InPageEventLogger } from "../../app/tracking/InPageEventLogger";
setEventLogger(new InPageEventLogger());

import { init } from "../init";
init();
