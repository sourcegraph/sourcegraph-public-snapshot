import { isE2ETest } from "../utils";
import { getPlatformName } from "../utils";

export abstract class EventLogger {

	logHover(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Hover", "SymbolHovered", eventProperties);
	}

	logClick(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "TooltipDocked", eventProperties);
	}

	logSelectText(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Select", "TextSelected", eventProperties);
	}

	logJumpToDef(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "GoToDefClicked", eventProperties);
	}

	logFindRefs(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "FindRefsClicked", eventProperties);
	}

	logSearch(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "SearchClicked", eventProperties);
	}

	logOpenFile(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "FileOpened", eventProperties);
	}

	logAuthClicked(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "AuthRedirected", eventProperties);
	}

	logSourcegraphSearch(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "SourcegraphSearchClicked", eventProperties);
	}

	logSourcegraphSearchTabClicked(eventProperties: any = {}): void {
		this.logEventForCategory("BrowserExtension", "Click", "SourcegraphSearchTabClicked", eventProperties);
	}

	protected abstract sendEvent(eventAction: string, eventProps: any): void;

	private defaultProperties(): any {
		return {
			path_name: window.location.pathname,
			Platform: getPlatformName(),
		};
	}

	private logEventForCategory(eventCategory: string, eventAction: string, eventLabel: string, eventProperties: any = {}): void {
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

		this.sendEvent(eventAction, decoratedEventProps);
	}
}
