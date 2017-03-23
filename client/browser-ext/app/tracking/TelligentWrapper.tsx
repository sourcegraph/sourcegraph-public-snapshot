// tslint:disable-next-line
const telligent = require("telligent-tracker");
const telligentFunctionName = "telligent";

export class TelligentWrapper {
	private t: any;
	constructor(appId: string, platform: string, forceSecure: boolean, installedChromeExtension: boolean, url?: string) {
		// Create the initializing function
		window[telligentFunctionName] = function (): void {
			(window[telligentFunctionName].q = window[telligentFunctionName].q || []).push(arguments);
		};

		// Set up the initial queue, if it doesn't already exist
		window[telligentFunctionName].q = new telligent.Telligent((window[telligentFunctionName].q || []), telligentFunctionName);

		this.t = (window as any).telligent;


		let telligentUrl = "sourcegraph-logging.telligentdata.com";
		if (url) {
			telligentUrl = url;
		}
		// Must be called once upon initialization
		this.t("newTracker", "SourcegraphExtensionTracker", telligentUrl, {
			encodeBase64: false,
			appId: appId,
			platform: platform,
			env: process.env.NODE_ENV,
			forceSecureTracker: forceSecure,
		});

		if (installedChromeExtension) {
			this.installedChromeExtension();
		}
	}
	track(eventAction: string, requestPayload: any): void {
		this.t("track", eventAction, requestPayload);
	}

	setUserId(requestPayload: any): void {
		this.t("setUserId", requestPayload);
	}

	addStaticMetadataObject(requestPayload: any): void {
		this.t("addStaticMetadataObject", requestPayload);
	}

	installedChromeExtension(): void {
		this.t("addStaticMetadata", "installed_chrome_extension", "true", "userInfo");
	}
}
