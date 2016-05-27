// @flow

import urlTo from "sourcegraph/util/urlTo";
import {urlToTree} from "sourcegraph/tree/routes";
import {rel} from "sourcegraph/app/routePatterns";
import {defPath} from "sourcegraph/def";
import type {Route} from "react-router";
import type {Def} from "sourcegraph/def";

export const routes: Array<Route> = [
	{
		path: rel.defInfo,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				const withResolvedRepoRev = require("sourcegraph/repo/withResolvedRepoRev").default;
				const withDef = require("sourcegraph/def/withDef").default;
				callback(null, {
					main: withResolvedRepoRev(withDef(require("sourcegraph/def/DefInfo").default)),
					repoNavContext: withResolvedRepoRev(require("sourcegraph/def/DefNavContext").default),
				});
			});
		},
	},
	{
		path: rel.def,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				const withResolvedRepoRev = require("sourcegraph/repo/withResolvedRepoRev").default;
				callback(null, {
					main: require("sourcegraph/blob/BlobLoader").default,
					repoNavContext: withResolvedRepoRev(require("sourcegraph/def/DefNavContext").default),
				}, [
					require("sourcegraph/blob/withLastSrclibDataVersion").default,
					require("sourcegraph/def/withDefAndRefLocations").default,
					require("sourcegraph/def/blobWithDefBox").default,
				]);
			});
		},
	},
];

function defParams(def: Def, rev?: ?string): Object {
	const revPart = rev ? `@${rev}` : "";
	return {splat: [`${def.Repo}${revPart}`, defPath(def)]};
}

// urlToDef returns a URL to the given def at the given revision.
export function urlToDef(def: Def, rev?: ?string): string {
	if ((def.File === null || def.Kind === "package")) {
		// The def's File field refers to a directory (e.g., in the
		// case of a Go package). We can't show a dir in this view,
		// so just redirect to the dir listing.
		//
		// TODO(sqs): Improve handling of this case.
		let file = def.File === "." ? "" : def.File;
		return urlToTree(def.Repo, rev, file);
	}
	return urlTo("def", defParams(def, rev));
}

// urlToDefInfo returns a URL to the given def's info at the given revision.
export function urlToDefInfo(def: Def, rev?: ?string): string {
	if ((def.File === null || def.Kind === "package")) {
		// The def's File field refers to a directory (e.g., in the
		// case of a Go package). We can't show a dir in this view,
		// so just redirect to the dir listing.
		//
		// TODO(sqs): Improve handling of this case.
		let file = def.File === "." ? "" : def.File;
		return urlToTree(def.Repo, rev, file);
	}
	return urlTo("defInfo", defParams(def, rev));
}

// urlToRepoDef returns a URL to the given repositories def at the given revision.
export function urlToRepoDef(repo: string, rev: ?string, def: string): string {
	const revPart = rev ? `@${rev}` : "";
	return urlTo("def", {splat: [`${repo}${revPart}`, def]});
}

// fastURLToRepoDef is a faster version of urlToRepoDef that hardcodes the route
// construction. It is brittle to route structure changes, but it is acceptable to
// use (to improve perf) it if you need to call it a lot.
export function fastURLToRepoDef(repo: string, rev: ?string, def: string): string {
	return `/${repo}${rev ? `@${rev}` : ""}/-/def/${def}`;
}
