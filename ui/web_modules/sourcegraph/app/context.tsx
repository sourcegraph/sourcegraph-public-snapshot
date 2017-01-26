import { EmailAddrList, User } from "sourcegraph/api";
import { ExternalToken } from "sourcegraph/user";
import { testOnly } from "sourcegraph/util/testOnly";

class Context {
	xhrHeaders: { [key: string]: string };
	csrfToken: string;
	userAgentIsBot: boolean;
	user: User | null;
	emails: EmailAddrList | null;
	gitHubToken: ExternalToken | null;
	googleToken: ExternalToken | null;
	sentryDSN: string;
	intercomHash: string;

	appURL: string; // base URL for app (e.g., https://sourcegraph.com or http://localhost:3080)
	assetsRoot: string; // URL path to image/font/etc. assets on server
	buildVars: { // from the build process (sgtool)
		Version: string;
		Date: string;
	};
	authEnabled: boolean;
	trackingAppID: string | null;

	constructor(ctx: any) {
		Object.assign(this, ctx);
	}

	hasPrivateGitHubToken(): boolean {
		return Boolean(this.gitHubToken && this.gitHubToken.scope.includes("repo") && this.gitHubToken.scope.includes("read:org"));
	}

	hasHookGitHubToken(): boolean {
		return Boolean(this.gitHubToken && this.gitHubToken.scope.includes("admin:repo_hook"));
	}

	hasOrganizationGitHubToken(): boolean {
		return Boolean(this.gitHubToken && this.gitHubToken.scope.includes("read:org"));
	}

	hasPrivateGoogleToken(): boolean {
		return Boolean(this.googleToken && this.googleToken.scope.includes("https://www.googleapis.com/auth/cloud-platform") && this.googleToken.scope.includes("https://www.googleapis.com/auth/userinfo.email") && this.googleToken.scope.includes("https://www.googleapis.com/auth/userinfo.profile"));
	}

	/**
	 * the chrome extension is detected when it creates a div with id `sourcegraph-app-background` on page.
	 * for on-premise or testing instances of Sourcegraph, the chrome extension never runs, so this will return false.
	 * proceed with caution, and think about using /util/shouldPromptToInstallBrowserExtension instead.
	 */
	hasChromeExtensionInstalled(): boolean {
		return document.getElementById("sourcegraph-app-background") !== null;
	}

	isSourcegraphCloud(): boolean {
		return SOURCEGRAPH_CLOUD_URL_PATTERN.test(this.appURL);
	}
}

const SOURCEGRAPH_CLOUD_URL_PATTERN = /^https?:\/\/sourcegraph.com/i;

export const context = new Context(global.__sourcegraphJSContext);

export function mockUser(user: User | null, f: () => void): void {
	testOnly();

	let prevUser = context.user;
	context.user = user;
	try {
		f();
	} finally {
		context.user = prevUser;
	}
};
