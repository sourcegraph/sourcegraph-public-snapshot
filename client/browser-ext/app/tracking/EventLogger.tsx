import { isE2ETest } from "../utils";
import { getPlatformName } from "../utils";
import { Features } from "../utils/Features";

export abstract class EventLogger {


	logHover(eventProperties: Object = {}): void {
		this.logEventForCategory("BrowserExtension", "Hover", "SymbolHovered", eventProperties);
	}

	logClick(eventProperties: Object = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "TooltipDocked", eventProperties);
	}

	logSelectText(eventProperties: Object = {}): void {
		this.logEventForCategory("BrowserExtension", "Select", "TextSelected", eventProperties);
	}

	logJumpToDef(eventProperties: Object = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "GoToDefClicked", eventProperties);
	}

	logFindRefs(eventProperties: Object = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "FindRefsClicked", eventProperties);
	}

	logSearch(eventProperties: Object = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "SearchClicked", eventProperties);
	}

	logOpenFile(eventProperties: Object = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "FileOpened", eventProperties);
	}

	logAuthClicked(eventProperties: Object = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "AuthRedirected", eventProperties);
	}

	protected abstract sendEvent(eventAction: string, eventProps: any): void;

	private logToConsole(eventAction: string, object: any): void {
		return;

		// TODO(john): unbreak below...
		// if (!Features.eventLogDebug.isEnabled()) {
		// 	return;
		// }
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
