import { hover } from "glamor";
import * as React from "react";
import { Link } from "react-router";
import { urlToRepo } from "sourcegraph/repo/routes";
import { urlToTree } from "sourcegraph/tree/routes";

interface Props {
	repo: string;
	path: string | null;
	rev: string | null;
	style?: React.CSSProperties;
	linkSx?: React.CSSProperties;
	linkHoverSx?: React.CSSProperties;
	sep?: string;
	toFile?: boolean;
}

export function PathBreadcrumb({ repo, path, rev, sep = "/", linkSx, linkHoverSx, style, toFile = true }: Props): JSX.Element {

	const links: JSX.Element[] = [];
	const linkHover = linkHoverSx ? hover(linkHoverSx) : null;

	if (repo) {
		links[0] = <Link key={0}
			to={urlToRepo(repo)}
			style={linkSx}
			{...linkHover}
			>
			{repo.split("/").join(" / ")}
		</Link>;
	}

	if (path !== null) {
		const pathToFile = path.split("/").slice(0, -1);
		const pathToDir = path.split("/");
		const pathCrumb = toFile ? pathToFile : pathToDir;
		const crumbs = pathCrumb.map((item, index) => <span key={index + 1} style={linkSx}>
			<span style={{ display: "inline-block", paddingLeft: 4, paddingRight: 4 }}>{sep}</span>
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
