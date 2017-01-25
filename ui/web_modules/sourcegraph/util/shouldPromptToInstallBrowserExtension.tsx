import { context } from "sourcegraph/app/context";

const MOBILE_USERAGENT_PATTERN = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i;
const SUPPORTED_USERAGENT_PATTERN = /Chrome/i;
export function shouldPromptToInstallBrowserExtension(): boolean {
	let isMobile = MOBILE_USERAGENT_PATTERN.test(navigator.userAgent);
	let isSupportedBrowser = SUPPORTED_USERAGENT_PATTERN.test(navigator.userAgent);
	// is a supported browser, is not mobile, extension isn't installed, and this is Sourcegraph.com
	return isSupportedBrowser && !isMobile && !context.hasChromeExtensionInstalled() && context.isSourcegraphCloud();
}
