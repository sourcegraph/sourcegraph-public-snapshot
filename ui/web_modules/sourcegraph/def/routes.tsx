// tslint:disable: typedef ordered-imports curly

import {urlTo} from "sourcegraph/util/urlTo";
import {urlToTree} from "sourcegraph/tree/routes";
import {rel} from "sourcegraph/app/routePatterns";
import {defPath} from "sourcegraph/def/index";
import {Def} from "sourcegraph/def/index";
import {DefInfo} from "sourcegraph/def/DefInfo";
import {DefNavContext} from "sourcegraph/def/DefNavContext";
import {BlobLoader} from "sourcegraph/blob/BlobLoader";
import {withLastSrclibDataVersion} from "sourcegraph/blob/withLastSrclibDataVersion";
import {withDefAndRefLocations} from "sourcegraph/def/withDefAndRefLocations";
import {blobWithDefBox} from "sourcegraph/def/blobWithDefBox";
import {withResolvedRepoRev} from "sourcegraph/repo/withResolvedRepoRev";
import {withDef} from "sourcegraph/def/withDef";

let _defInfoComponents;
let _defComponents;

export const routes: any[] = [
	{
		path: rel.defInfo,
		repoNavContext: false,
		getComponents: (location, callback) => {
			if (!_defInfoComponents) {
				_defInfoComponents = {
					main: withResolvedRepoRev(withDef(DefInfo)),
				};
			}
			callback(null, _defInfoComponents);
		},
	},
	{
		path: rel.def,
		keepScrollPositionOnRouteChangeKey: "file",
		getComponents: (location, callback) => {
			if (!_defComponents) {
				_defComponents = {
					main: BlobLoader,
					repoNavContext: withResolvedRepoRev(DefNavContext),
				};
			}
			callback(null, _defComponents);
		},
		blobLoaderHelpers: [withLastSrclibDataVersion, withDefAndRefLocations, blobWithDefBox],
	},
];

function defParams(def: Def, rev?: string | null): any {
	const revPart = rev ? `@${rev}` : "";
	return {splat: [`${def.Repo}${revPart}`, defPath(def)]};
}

// urlToDef returns a URL to the given def at the given revision.
export function urlToDef(def: Def, rev?: string | null): string {
	if ((def.File === null || def.Kind === "package")) {
		// The def's File field refers to a directory (e.g., in the
		// case of a Go package). We can't show a dir in this view,
		// so just redirect to the dir listing.
		//
		// TODO(sqs): Improve handling of this case.
		let file = def.File === "." ? "" : def.File;
		return urlToTree(def.Repo, rev || null, file);
	}
	return urlTo("def", defParams(def, rev));
}

// urlToDefInfo returns a URL to the given def's info at the given revision.
export function urlToDefInfo(def: Def, rev?: string | null): string {
	if ((def.File === null || def.Kind === "package")) {
		// The def's File field refers to a directory (e.g., in the
		// case of a Go package). We can't show a dir in this view,
		// so just redirect to the dir listing.
		//
		// TODO(sqs): Improve handling of this case.
		let file = def.File === "." ? "" : def.File;
		return urlToTree(def.Repo, rev || null, file);
	}
	return urlTo("defInfo", defParams(def, rev));
}

// urlToRepoDef returns a URL to the given repositories def at the given revision.
export function urlToRepoDef(repo: string, rev: string | null, def: string): string {
	const revPart = rev ? `@${rev}` : "";
	return urlTo("def", {splat: [`${repo}${revPart}`, def]} as any);
}

// fastURLToRepoDef is a faster version of urlToRepoDef that hardcodes the route
// construction. It is brittle to route structure changes, but it is acceptable to
// use (to improve perf) it if you need to call it a lot.
export function fastURLToRepoDef(repo: string, rev: string | null, def: string): string {
	return `/${repo}${rev ? `@${rev}` : ""}/-/def/${def}`;
}
