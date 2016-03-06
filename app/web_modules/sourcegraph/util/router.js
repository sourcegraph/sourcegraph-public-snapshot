export function tree(repo, rev, path, startLine, endLine) {
	let line = startLine && endLine ? `#L${startLine}-${endLine}` : "";
	return `/${repo}${rev ? `@${rev}` : ""}/.tree/${path}${line}`;
}
