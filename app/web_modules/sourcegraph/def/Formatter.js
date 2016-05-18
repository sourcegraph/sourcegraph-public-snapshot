// @flow weak

import React from "react";
import type {Def} from "sourcegraph/def/index";

export type DefFormatOptions = {
	nameClass?: string;
	unqualifiedNameClass?: string;
}

export function qualifiedNameAndType(def, opts: ?DefFormatOptions) {
	if (!def) throw new Error("def is null");
	if (!def.FmtStrings) return "(unknown def)";
	const f = def.FmtStrings;

	let name = f.Name.ScopeQualified;
	if (f.Name.Unqualified && name) {
		let parts = name.split(f.Name.Unqualified);
		name = [
			parts.slice(0, parts.length - 1).join(f.Name.Unqualified),
			<span key="unqualified" className={opts && opts.unqualifiedNameClass}>{f.Name.Unqualified}</span>,
		];
	}

	return [
		f.DefKeyword,
		f.DefKeyword ? " " : "",
		<span key="name" className={opts && opts.nameClass} style={opts && opts.nameClass ? {} : {fontWeight: "bold"}}>{name}</span>, // give default bold styling if not provided
		f.NameAndTypeSeparator,
		f.Type.ScopeQualified,
	];
}

// defTitleOK reports if it's safe to call defTitle with def.
export function defTitleOK(def: Def): bool {
	return def && def.FmtStrings && def.Unit;
}

// defTitle returns a title for def.
// It uses logic similar to qualifiedNameAndType, but it prepends package name,
// omits type information, and produces a string rather than HTML.
// defTitle is safe to call if and only if defTitleOK returns true for def.
export function defTitle(def: Def): string {
	const f = def.FmtStrings;

	let name = f.Name.ScopeQualified;
	if (f.Name.Unqualified && name) {
		let parts = name.split(f.Name.Unqualified);
		name = [
			parts.slice(0, parts.length - 1).join(f.Name.Unqualified),
			f.Name.Unqualified,
		].join("");
	}

	// Last '/'-separated element of unit.
	let unit = def.Unit;
	let i = unit.lastIndexOf("/");
	if (i !== -1) {
		unit = unit.substring(i + 1);
	}

	return `${unit}.${name}`;
}
