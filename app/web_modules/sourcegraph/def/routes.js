// @flow

import urlTo from "sourcegraph/util/urlTo";
import {urlToTree} from "sourcegraph/tree/routes";
import {rel} from "sourcegraph/app/routePatterns";
import {defPath} from "sourcegraph/def";
import type {Route} from "react-router";
import type {Def} from "sourcegraph/def";

import withResolvedRepoRev from "sourcegraph/repo/withResolvedRepoRev";
import withDef from "sourcegraph/def/withDef";
import DefInfo from "sourcegraph/def/DefInfo";
import DefNavContext from "sourcegraph/def/DefNavContext";

// TODO these routes didn't work with async loading. Fix them.
const infoRoute = {
	path: "info",
	components: {
		main: withResolvedRepoRev(withDef(DefInfo)),
		repoNavContext: DefNavContext,
	},
};

export const routes: Array<Route> =[
	{
		path: `${rel.def}*/-/`,
		getChildRoutes: (location, callback) => {
			callback(null, [infoRoute]);
		},
	},
	{
		path: rel.def,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/blob/BlobLoader").default,
					repoNavContext: require("sourcegraph/def/DefNavContext").default,
				}, [
					require("sourcegraph/def/withDefAndRefLocations").default,
					require("sourcegraph/def/blobWithDefBox").default,
				]);
			});
		},
	},
];

function defParams(def: Def, rev: ?string): Object {
	rev = rev === null ? def.CommitID : rev;
	const revPart = rev ? `@${rev || def.CommitID}` : "";
	return {splat: [`${def.Repo}${revPart}`, defPath(def)]};
}

export function urlToDef(def: Def, rev: ?string): string {
	rev = rev === null ? def.CommitID : rev;
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

// TODO: add revision
export function urlToDefInfo(def: Def): string {
	return urlTo("defInfo", defParams(def));
}

export function urlToDef2(repo: string, rev: string, def: string): string {
	return urlTo("def", {splat: [`${repo}@${rev}`, def]});
}
