import * as React from "react";
import { Link } from "react-router";
import { urlToTree } from "sourcegraph/tree/routes";

export function RepoLink({ repo, rev, style, className, spacing = 4 }: {
	repo: string;
	rev: string | null;
	style?: React.CSSProperties;
	className?: string;
	spacing?: number;
}): JSX.Element {
	const repoParts = repo.split("/");
	const repoLink: JSX.Element[] = [];

	repoParts.forEach((dir, i) => {
		repoLink.push(<span key={dir + i}>{dir}</span>);
		if (i === repoParts.length - 1) {
			return;
		}
		repoLink.push(<span style={{
			paddingLeft: spacing,
			paddingRight: spacing,
		}} key={i}>/</span>);
	});

	return <Link key={0} to={urlToTree(repo, rev, [])} style={style} className={className}>{repoLink}</Link>;
}
