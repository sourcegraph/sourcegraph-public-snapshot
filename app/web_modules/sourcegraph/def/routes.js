// @flow

import urlTo from "sourcegraph/util/urlTo";
import {rel} from "sourcegraph/app/routePatterns";
import {defPath} from "sourcegraph/def";
import type {Route} from "react-router";
import type {Def} from "sourcegraph/def";

let _components;

export const routes: Array<Route> =[
	{
		path: `${rel.def}/-/refs`,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				if (!_components) {
					const withResolvedRepoRev = require("sourcegraph/repo/withResolvedRepoRev").default;
					const withDef = require("sourcegraph/def/withDef").default;
					_components = {
						main: withResolvedRepoRev(withDef(require("sourcegraph/def/RefsMain").default)),
					};
				}
				callback(null, _components);
			});
		},
	},
	{
		path: rel.def,
		getComponents: (location, callback) => {
			require.ensure([], (require) => {
				callback(null, {
					main: require("sourcegraph/blob/BlobLoader").default,
				}, [
					require("sourcegraph/def/withDefAndRefs").default,
					require("sourcegraph/def/blobWithDefBox").default,
				]);
			});
		},
	},
];

function defParams(def: Def): Object {
	return {splat: [`${def.Repo}@${def.CommitID}`, defPath(def)]};
}

export function urlToDef(def: Def): string {
	return urlTo("def", defParams(def));
}

export function urlToDefRefs(def: Def, file?: string): string {
	let u = urlTo("defRefs", defParams(def));
	if (file) return `${u}?file=${encodeURIComponent(file)}`;
	return u;
}

export function urlToDef2(repo: string, rev: string, def: string): string {
	return urlTo("def", {splat: [`${repo}@${rev}`, def]});
}
