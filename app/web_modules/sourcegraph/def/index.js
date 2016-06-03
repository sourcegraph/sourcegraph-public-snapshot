// @flow weak

import {abs, getRouteParams} from "sourcegraph/app/routePatterns";
import {repoPath, repoRev, repoParam} from "sourcegraph/repo";

export type Repo = Object;
export type Def = Object;
export type DefKey = {
	Repo: string;
	CommitID: string;
	UnitType: string;
	Unit: string;
	Path: string; // def path, not file path
};

export type Ref = Object;

// Refs streaming pagnination assumes that the per page amount will be
// consistent for each fetch.
export const RefLocsPerPage = 30;

export function routeParams(url: string): {repo: string, rev: ?string, def: string, err: ?string} {
	let v = getRouteParams(abs.def, url);
	if (!v) {
		return {
			repo: "",
			rev: null,
			def: "",
			err: `Invalid def URL: ${url}`,
		};
	}
	return {
		repo: repoPath(repoParam(v.splat[0])),
		rev: repoRev(repoParam(v.splat[0])),
		def: v.splat[1],
		err: null,
	};
}

// fastParseDefPath quickly parses "TYPE/-/UNIT/-/PATH" from a URL
// or pathname like "/REPO/-/def/TYPE/-/UNIT/-/PATH". It is much
// faster than routeParams but should only be called on URLs that are
// known to be def URLs.
const _defPathIndicator = "/-/def/";
export function fastParseDefPath(url: string): ?string {
	const i = url.indexOf(_defPathIndicator);
	if (i === -1) return null;
	return url.slice(i + _defPathIndicator.length);
}

export function defPath(def: Def): string {
	return `${def.UnitType}/${def.Unit}/-/${encodeDefPath(def.Path)}`;
}

export function encodeDefPath(path: string): string {
	if (path && path !== "") {
		return path.replace("#", encodeURIComponent("#"));
	}
	return path;
}

export type RefLocationsKey = {
	repo: string;
	commitID: string;
	def: string;
	page?: number;
	perPage?: number;
	repos: Array<string>;
}
