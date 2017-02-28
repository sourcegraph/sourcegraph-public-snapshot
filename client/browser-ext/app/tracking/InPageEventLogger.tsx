import { EventLogger } from "./EventLogger";
import { TelligentWrapper } from "./TelligentWrapper";

export class InPageEventLogger extends EventLogger {
	private telligentWrapper: TelligentWrapper;

	constructor(appId: string, platformId: string) {
		super();
		this.telligentWrapper = new TelligentWrapper(appId, platformId, false, false);
	}

	setUserName(username: string | null): void {
		if (username !== null) {
			this.telligentWrapper.setUserId(username);
		}
	}

	protected logEventToTelligent(eventAction: string, eventProps: any): void {
		this.telligentWrapper.track(eventAction, eventProps);
	}

}
