import {hover} from "glamor";
import * as React from "react";
import {Link} from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import {UnsupportedLanguageAlert} from "sourcegraph/blob/UnsupportedLanguageAlert";
import {FlexContainer, Heading} from "sourcegraph/components";
import {colors, typography} from "sourcegraph/components/utils";
import {whitespace} from "sourcegraph/components/utils/index";
import {RevSwitcher} from "sourcegraph/repo/RevSwitcher";
import {urlToRepo} from "sourcegraph/repo/routes";
import {urlToTree} from "sourcegraph/tree/routes";
import {getPathExtension, supportedExtensions} from "sourcegraph/util/supportedExtensions";

interface Props {
	repo: string;
	path: string;
	repoObj: any;
	rev: string;
	commitID: string;
	routes: Object[];
	routeParams: RouteParams;
	isCloning: boolean;
	toast: string | null;
}

const sx = {
	backgroundColor: colors.coolGray1(),
	boxShadow: `0 2px 6px 0px ${colors.black(0.2)}`,
	zIndex: 1,
	padding: `${whitespace[2]} ${whitespace[3]}`,
};

const subSx = Object.assign({},
	{ color: colors.coolGray3() },
	typography.size[7],
);

const subHover = {
	color: `${colors.coolGray4()} !important`,
};

const toastSx = Object.assign({},
	{
		color: colors.orange(),
		marginTop: "auto",
		marginBottom: "auto",
	},
	typography.size[8],
);

function BreadCrumb({repo, path, rev}: {repo: string, path: string, rev: string}): JSX.Element {
	const pathToFile = path.split("/").slice(0, -1);
	const links: JSX.Element[] = [];
	links[0] = 	<Link
		key={0}
		{...hover(subHover)}
		style={subSx}
		to={urlToRepo(repo)}>{repo.split("/").join(" / ")}
	</Link>;

	const crumbs = pathToFile.map((item, index) => <span key={index + 1}
	style={subSx}
	>&nbsp;/&nbsp;
		<Link
			style={subSx}
			{...hover(subHover)}
			to={urlToTree(repo, rev, pathToFile.slice(0, index + 1))}>
			{item}
		</Link>
	</span>);

	links.push(...crumbs);

	return <span>
		{links}
	</span>;
};

function basename(path: string): string {
	const base = path.split("/").pop();
	return base || path;
};

export function BlobTitle({
	repo,
	path,
	repoObj,
	rev,
	commitID,
	routes,
	routeParams,
	isCloning,
	toast,
}: Props): JSX.Element {
	const extension = getPathExtension(path);
	const isSupported = extension ? supportedExtensions.indexOf(extension) !== -1 : false;

	return <div style={sx}>
		<FlexContainer justify="between">
			<div>
				<Heading level={5} color="white" style={{marginBottom: 0}}>
					{basename(path)}
					{commitID && <RevSwitcher
						repo={repo}
						repoObj={repoObj}
						rev={rev}
						commitID={commitID}
						routes={routes}
						routeParams={routeParams}
						isCloning={isCloning} />}
				</Heading>
				<BreadCrumb repo={repo} path={path} rev={rev} />
			</div>
			{!isSupported && <UnsupportedLanguageAlert ext={extension}/>}
			{toast && <div style={toastSx}>{toast}</div>}
		</FlexContainer>
	</div>;
};
