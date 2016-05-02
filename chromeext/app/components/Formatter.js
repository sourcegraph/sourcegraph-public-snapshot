import React from "react";

export function qualifiedNameAndType(def, opts) {
	if (!def) throw new Error("def is null");
	if (!def.FmtStrings) return "(unknown def)";
	const f = def.FmtStrings;

	let name = f.Name.ScopeQualified;
	if (f.Name.Unqualified) {
		let parts = name.split(f.Name.Unqualified);
		name = [
			parts[0],
			<span key="unqualified" className={opts && opts.unqualifiedNameClass}>{f.Name.Unqualified}</span>,
		];
	}

	return [
		f.DefKeyword,
		" ",
		<span key="name" className={opts && opts.nameClass} style={opts && opts.nameClass ? {} : {fontWeight: "bold"}}>{name}</span>, // give default bold styling if not provided
		f.NameAndTypeSeparator,
		f.Type.ScopeQualified,
	];
}
