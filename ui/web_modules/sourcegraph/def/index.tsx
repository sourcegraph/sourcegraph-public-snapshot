import {abs, getRouteParams} from "sourcegraph/app/routePatterns";
import {repoParam, repoPath, repoRev} from "sourcegraph/repo/index";

export interface Options {};

export interface Repo {
	URI: string;
	Description: string;
};

export interface Def {
	Repo: string;
	UnitType: string;
	Unit: string;
	Path: string;
	FmtStrings: any;
	Docs: any;
	Kind: string;
	File: string;
	Error: any;
};

export interface DefKey {
	Repo: string;
	CommitID: string;
	UnitType: string;
	Unit: string;
	Path: string; // def path, not file path
};

export interface AuthorshipInfo {
	LastCommitDate: string;
	LastCommitID: string;
};
export interface DefAuthorship extends AuthorshipInfo {
	Bytes: number;
	BytesProportion: number;
};
export interface DefAuthor extends DefAuthorship {
	Email: string;
	AvatarURL: string;
};

export interface Ref {};

export function routeParams(url: string): {repo: string, rev: string | null, def: string, err: string | null} {
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
export function fastParseDefPath(url: string): string | null {
	const i = url.indexOf(_defPathIndicator);
	if (i === -1) { return null; }
	return url.slice(i + _defPathIndicator.length);
}

export function defPath(def: Def): string {
	return `${def.UnitType}/${maybeTransformUnit(def.Unit)}/-/${encodeDefPath(def.Path)}`;
}

// maybeTransformUnit handles if def.Unit is ".". URLs with a
// "/./" will be automatically modified by the browser, so we
// transform it to "/_._/".
function maybeTransformUnit(unit: string): string {
	if (unit === ".") {
		return "_._";
	}
	return unit;
}

export function encodeDefPath(path: string): string {
	if (path && path !== "") {
		return path.replace("#", encodeURIComponent("#"));
	}
	return path;
}

export interface RefLocationsKey {
	repo: string;
	commitID: string | null;
	rev?: any;
	def: string;
	repos: string[];
}

export interface LocalRefLocationKey {
	Path: string;
	Count: number;
}

export interface LocalRefLocationsKey {
	TotalFiles: number;
	Files: LocalRefLocationKey[];
}

export interface ExamplesKey {
	repo: string;
	commitID: string | null;
	def: string;
}
