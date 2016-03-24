import React from "react";

export function qualifiedNameAndType(def) {
	if (!def) throw new Error("def is null");
	if (!def.FmtStrings) return "(unknown def)";
	const f = def.FmtStrings;
	return [
		f.DefKeyword,
		" ",
		<span className="name" key="name">{f.Name.ScopeQualified}</span>,
		f.NameAndTypeSeparator,
		f.Type.ScopeQualified,
	];
}
