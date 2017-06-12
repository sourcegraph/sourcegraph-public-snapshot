/**
 * This is done before all other imports to ensure that the event logger is set ahead of time.
 */
import { setEventLogger, setPhabricatorInstance, setSourcegraphUrl } from "../../app/utils/context";
import { UMAMI_SOURCEGRAPH_URL, umamiPhabricatorInstance } from "../umami/constants";
setSourcegraphUrl(UMAMI_SOURCEGRAPH_URL);
setPhabricatorInstance(umamiPhabricatorInstance);

import { InPageEventLogger } from "../../app/tracking/InPageEventLogger";
setEventLogger(new InPageEventLogger());

import { init } from "../init";
init();
