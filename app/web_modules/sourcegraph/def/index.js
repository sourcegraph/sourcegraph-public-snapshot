// @flow weak

import {abs, getRouteParams} from "sourcegraph/app/routePatterns";
import {repoPath, repoRev, repoParam} from "sourcegraph/repo";

export type Def = Object;
export type DefKey = {
	Repo: string;
	CommitID: string;
	UnitType: string;
	Unit: string;
	Path: string; // def path, not file path
};

export type Ref = Object;

export function routeParams(url: string): {repo: string, rev: ?string, def: string} {
	let v = getRouteParams(abs.def, url);
	if (!v) throw new Error(`Invalid def URL: ${url}`);
	return {
		repo: repoPath(repoParam(v.splat[0])),
		rev: repoRev(repoParam(v.splat[0])),
		def: v.splat[1],
	};
}

export function defPath(def: Def): string {
	return `${def.UnitType}/${def.Unit}/-/${def.Path}`;
}
