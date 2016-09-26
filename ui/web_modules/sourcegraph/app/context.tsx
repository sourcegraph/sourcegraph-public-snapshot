import {EmailAddrList, User} from "sourcegraph/api";
import {ExternalToken} from "sourcegraph/user";
import {EventLogger} from "sourcegraph/util/EventLogger";
import {testOnly} from "sourcegraph/util/testOnly";

class Context {
	xhrHeaders: {[key: string]: string};
	userAgentIsBot: boolean;
	user: User | null;
	emails: EmailAddrList | null;
	gitHubToken: ExternalToken | null;
	intercomHash: string;

	appURL: string; // base URL for app (e.g., https://sourcegraph.com or http://localhost:3080)
	assetsRoot: string; // URL path to image/font/etc. assets on server
	buildVars: { // from the build process (sgtool)
		Version: string;
	};

	features: Features;

	hasPrivateGitHubToken(): boolean {
		return Boolean(this.gitHubToken && this.gitHubToken.scope.includes("repo") && this.gitHubToken.scope.includes("read:org"));
	}

	hasHookGitHubToken(): boolean {
		return Boolean(this.gitHubToken && this.gitHubToken.scope.includes("admin:repo_hook"));
	}
}

export const context = new Context();

export interface Features {
	Authors: any;
	GodocRefs: any;
};

// Sets the values of the context given a JSContext object from the server.
export function reset(args: {appURL: string, assetsRoot: string, buildVars: {Version: string}, features: Features}): void {
	if (typeof args.features !== "undefined") {
		context.features = args.features;
	}
	delete args.features;
	if (typeof args.appURL === "undefined" || typeof args.assetsRoot === "undefined" || typeof args.buildVars === "undefined") {
		throw new Error("appURL, assetsRoot, and buildVars must all be set");
	}
	Object.assign(context, args);

	EventLogger.init();
}

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
