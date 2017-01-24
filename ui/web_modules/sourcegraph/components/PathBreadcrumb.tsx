import { hover } from "glamor";
import * as React from "react";
import { Link } from "react-router";
import { urlToTree } from "sourcegraph/tree/routes";

interface Props {
	repo: string;
	path: string | null;
	rev: string | null;
	style?: React.CSSProperties;
	linkSx?: React.CSSProperties;
	linkHoverSx?: React.CSSProperties;
	toFile?: boolean;
}

export function PathBreadcrumb({ repo, path, rev, linkSx, linkHoverSx, style, toFile = true }: Props): JSX.Element {

	const links: JSX.Element[] = [];
	const linkHover = linkHoverSx ? hover(linkHoverSx) : null;

	const repoParts = repo.split("/");
	const repoLink: (string | JSX.Element)[] = [];
	repoParts.forEach((dir, i) => {
		repoLink.push(dir);
		if (i === repoParts.length - 1) {
			return;
		}
		repoLink.push(<span style={{
			paddingLeft: 4,
			paddingRight: 4,
		}} key={i}>/</span>);
	});

	links[0] = <Link key={0}
		to={urlToTree(repo, rev, [])}
		style={linkSx}
		{...linkHover}
	>
		{repoLink}
	</Link>;

	if (path !== null) {
		const pathToFile = path.split("/").slice(0, -1);
		const pathToDir = path.split("/");
		const pathCrumb = toFile ? pathToFile : pathToDir;
		const crumbs = pathCrumb.map((item, index) => <span key={index + 1} style={linkSx}>
			<span style={{ display: "inline-block", paddingLeft: 4, paddingRight: 4 }}>/</span>
			<Link
				to={urlToTree(repo, rev, pathToFile.slice(0, index + 1))}
				style={linkSx}
				{...linkHover}>
				{item}
			</Link>
		</span>);

		links.push(...crumbs);
	}

	return <div style={style}>{links}</div>;
};
