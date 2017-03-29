import { isE2ETest } from "../utils";
import { getPlatformName } from "../utils";
import { Features } from "../utils/Features";

export abstract class EventLogger {

	logJumpToDef(eventProperties: Object = {}): void {
		this.logEventForCategory("Def", "Click", "JumpDef", eventProperties);
	}

	logHover(eventProperties: Object = {}): void {
		this.logEventForCategory("Def", "Hover", "HighlightDef", eventProperties);
	}

	logOpenFile(eventProperties: Object = {}): void {
		this.logEventForCategory("File", "Click", "ChromeExtensionSgButtonClicked", eventProperties);
	}

	logAuthClicked(eventProperties: Object = {}): void {
		this.logEventForCategory("Auth", "Redirect", "ChromeExtensionAuthButtonClicked", eventProperties);
	}

	protected abstract sendEvent(eventAction: string, eventProps: any): void;

	private logToConsole(eventAction: string, object: any): void {
		if (!Features.eventLogDebug.isEnabled()) {
			return;
		}
		// tslint:disable-next-line
		console.debug("%cEVENT %s", "color: #aaa", eventAction, object);
	}

	private defaultProperties(): Object {
		return {
			path_name: window.location.pathname,
			Platform: getPlatformName(),
		};
	}

	private logEventForCategory(eventCategory: string, eventAction: string, eventLabel: string, eventProperties: Object = {}): void {
		if (isE2ETest()) {
			return;
		}

		const decoratedEventProps = Object.assign({}, eventProperties, this.defaultProperties(),
			{
				eventLabel,
				eventCategory,
				eventAction,
			},
		);

		this.logToConsole(eventAction, decoratedEventProps);
		this.sendEvent(eventAction, decoratedEventProps);
	}

}
