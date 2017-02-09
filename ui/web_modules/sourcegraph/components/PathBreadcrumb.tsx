import { hover } from "glamor";
import * as React from "react";
import { Link } from "react-router";
import { RepoLink } from "sourcegraph/components";
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

interface CrumbProps {
	url: string;
	style?: React.CSSProperties;
	linkClassName?: string;
	linkSx?: React.CSSProperties;
	children?: React.ReactNode[];
	key?: string | number;
}

interface PathCrumb {
	dirName: string;
	url: string;
}

const crumbSpacing = 4;

function getPathCrumbs(
	repo: string,
	rev: string | null,
	path: string,
	toFile: boolean,
): PathCrumb[] {
	const pathParts = path.split("/");
	const pathToFile = pathParts.slice(0, -1);
	const pathArray = toFile ? pathToFile : pathParts;
	return pathArray.map(
		(dir, i) => {
			return {
				dirName: dir,
				url: urlToTree(repo, rev, pathToFile.slice(0, i + 1)),
			};
		}
	);
}

function Crumb({ style, url, linkSx, linkClassName, children }: CrumbProps): JSX.Element {
	return <span style={style}>
		<span style={{ display: "inline-block", paddingLeft: crumbSpacing, paddingRight: crumbSpacing }}>/</span>
		<Link to={url} style={linkSx} className={linkClassName}>{children}</Link>
	</span>;
}

export function PathBreadcrumb({ repo, path, rev, linkSx, linkHoverSx, style, toFile = true }: Props): JSX.Element {

	const links: JSX.Element[] = [];
	const linkHoverClass = linkHoverSx ? hover(linkHoverSx).toString() : "";

	links.push(
		<RepoLink
			repo={repo}
			rev={rev}
			style={linkSx}
			key="RepoLink"
			className={linkHoverClass}
			spacing={crumbSpacing} />
	);

	if (path !== null) {
		const crumbs = getPathCrumbs(repo, rev, path, toFile);
		const crumbEls = crumbs.map((crumb, i) => <Crumb
			key={i}
			url={crumb.url}
			linkSx={linkSx}
			linkClassName={linkHoverClass}>{crumb.dirName}</Crumb>);
		links.push(...crumbEls);
	}

	return <div style={style}>{links}</div>;
}
