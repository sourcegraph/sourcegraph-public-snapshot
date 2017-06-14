import { sourcegraphUrl } from "../utils/context";
import { EventLogger } from "./EventLogger";
import { TelligentWrapper } from "./TelligentWrapper";

export class InPageEventLogger extends EventLogger {
	private userId: string | null;
	private telligentWrapper: TelligentWrapper;

	constructor() {
		super();
		// remove http or https from address since telligent adds it back in
		const telligentUrl = sourcegraphUrl.replace("http://", "").replace("https://", "");
		this.telligentWrapper = new TelligentWrapper("SourcegraphExtension", "PhabricatorExtension", false, false, `${telligentUrl}/.bi-logger`);
	}

	setUserId(userId: string | null): void {
		this.userId = userId;
		this.telligentWrapper.setUserId(userId);
	}

	protected sendEvent(eventAction: string, eventProps: any): void {
		eventProps.userId = this.userId;
		this.telligentWrapper.track(eventAction, eventProps);
	}

}
