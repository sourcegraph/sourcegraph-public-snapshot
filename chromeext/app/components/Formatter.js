import React from "react";

// TECH DEBT: this file was copied from "app/web_modules/sourcegraph/def/Formatter.js"

export function qualifiedNameAndType(def, opts) {
	if (!def) throw new Error("def is null");
	if (!def.FmtStrings) return "(unknown def)";
	const f = def.FmtStrings;

	let name = f.Name.ScopeQualified;
	if (f.Name.Unqualified) {
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
