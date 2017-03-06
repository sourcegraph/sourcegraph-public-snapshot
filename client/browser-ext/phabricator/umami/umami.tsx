/**
 * This is done before all other imports to ensure that the event logger is set ahead of time.
 */
import { InPageEventLogger } from "../../app/tracking/InPageEventLogger";
import { setEventLogger, setPhabricatorInstance, setSourcegraphUrl } from "../../app/utils/context";
import { UMAMI_SOURCEGRAPH_URL, umamiPhabricatorInstance } from "./constants";
setEventLogger(new InPageEventLogger("SourcegraphExtension", "PhabricatorExtension"));
setSourcegraphUrl(UMAMI_SOURCEGRAPH_URL);
setPhabricatorInstance(umamiPhabricatorInstance);

import { init } from "../init";
init();
