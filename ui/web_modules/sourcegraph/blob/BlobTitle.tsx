import {hover} from "glamor";
import * as React from "react";
import {Link} from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import {UnsupportedLanguageAlert} from "sourcegraph/blob/UnsupportedLanguageAlert";
import {Base, FlexContainer, Heading} from "sourcegraph/components";
import {colors, typography} from "sourcegraph/components/utils";
import {RevSwitcherContainer} from "sourcegraph/repo/RevSwitcherContainer";
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

function getFilePath(repo: string, path: string): JSX.Element[] {
	const filePathArray = repo.split("/").concat(path.split("/"));
	const isGitHubRepo = filePathArray[0] === "github.com";
	filePathArray.pop();

	return filePathArray.map((item, i, array) => {
		const relPath = isGitHubRepo
			? array.slice(3, i + 1).join("/")
			: array.slice(0, i + 1).join("/");

		if (isGitHubRepo && i >= 1 && i <= 2) { return <span key={i} />; };

		return item === "github.com"
			? <span key={i}>
				<Link
					{...hover(subHover)}
					style={subSx}
					to={`/${repo}`}>
					{repo.split("/").join(" / ")}</Link>
			</span>
			: <span key={i}>
				&nbsp;/&nbsp;
				<Link
					{...hover(subHover)}
					style={subSx}
					to={`/${repo}/-/tree/${relPath}`} >{item}</Link>
			</span>;
	});
};

function getFilename(repo: string, path: string): string | undefined {
	return repo.split("/").concat(path.split("/")).pop();
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

	return <Base style={sx} px={3} py={2}>
		<FlexContainer justify="between">
			<div>
				<Heading level={5} color="white" mb={0}>
					{getFilename(repo, path)}
					{commitID && <RevSwitcherContainer
						repo={repo}
						repoObj={repoObj}
						rev={rev}
						commitID={commitID}
						routes={routes}
						routeParams={routeParams}
						isCloning={isCloning} />}
				</Heading>
				<span style={subSx}>{getFilePath(repo, path)}</span>
			</div>
			{!isSupported && <UnsupportedLanguageAlert ext={extension}/>}
			{toast && <div style={toastSx}>{toast}</div>}
		</FlexContainer>
	</Base>;
};
