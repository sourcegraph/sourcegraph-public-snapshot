/**
 * This is done before all other imports to ensure that the event logger is set ahead of time.
 */
import { InPageEventLogger } from "../../app/tracking/InPageEventLogger";
import { setEventLogger, setPhabricatorInstance, setSourcegraphUrl } from "../../app/utils/context";
import { SGDEV_SOURCEGRAPH_URL, sgDevPhabricatorInstance } from "./constants";
setEventLogger(new InPageEventLogger("SourcegraphExtension", "PhabricatorExtension"));
setSourcegraphUrl(SGDEV_SOURCEGRAPH_URL);
setPhabricatorInstance(sgDevPhabricatorInstance);

import { init } from "../init";
init();
