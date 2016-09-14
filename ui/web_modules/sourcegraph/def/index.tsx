import { Def } from "sourcegraph/api";

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

export function defPath(def: Def): string {
	return `${def.UnitType}/${def.Unit && maybeTransformUnit(def.Unit)}/-/${encodeDefPath(def.Path)}`;
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
