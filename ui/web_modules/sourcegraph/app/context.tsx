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

	hasChromeExtensionInstalled(): boolean {
		return document.getElementById("sourcegraph-app-bootstrap") !== null;
	}
}

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
