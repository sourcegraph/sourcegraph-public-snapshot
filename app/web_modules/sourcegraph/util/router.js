export function tree(repoPath, rev, path, startLine, endLine) {
	let line = startLine && endLine ? `#L${startLine}-${endLine}` : "";
	return `${repoRev(repoPath, rev)}/-/tree/${path}${line}`;
}

export function repo(repoPath) {
	return `/${repoPath}`;
}

export function repoRev(repoPath, rev) {
	return `${repo(repoPath)}${rev ? `@${rev}` : ""}`;
}

// def constructs the application (not API) URL to a def. The def
// spec may be passed as individual arguments as listed in the
// function prototype, or as one DefSpec object argument.
export function def(repoPath, rev, unitType, unit, path) {
	if (typeof repoPath === "object") {
		let defSpec = repoPath;
		repoPath = defSpec.Repo;
		rev = defSpec.CommitID;
		unitType = defSpec.UnitType;
		unit = defSpec.Unit;
		path = defSpec.Path;
	}
	return `${repo(repoPath, rev)}/-/def/${unitType}/${unit}/-/${path === "." ? "_._" : path}`;
}

export function build(repoPath, id) {
	return `${repo(repoPath)}/-/builds/${id}`;
}

export function repoCommits(repoPath, rev) {
	return `${repoRev(repoPath, rev)}/commits`;
}
