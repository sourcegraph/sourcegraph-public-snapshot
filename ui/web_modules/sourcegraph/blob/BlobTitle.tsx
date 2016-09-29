import * as React from "react";
import {Link} from "react-router";
import {Base, Heading} from "sourcegraph/components";
import {colors, typography} from "sourcegraph/components/utils";
import {RevSwitcherContainer} from "sourcegraph/repo/RevSwitcherContainer";

interface Props {
	repo: string;
	path: string;
	repoObj: any;
	rev: string;
	commitID: string;
	routes: Array<Object>;
	routeParams: any;
	isCloning: boolean;
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

export const BlobTitle = ({
	repo,
	path,
	repoObj,
	rev,
	commitID,
	routes,
	routeParams,
	isCloning,
}: Props) => <Base style={sx} px={3} py={2}>
	<Heading level={5} color="white" mb={0}>
		{path}
		{commitID && <RevSwitcherContainer
			repo={repo}
			repoObj={repoObj}
			rev={rev}
			commitID={commitID}
			routes={routes}
			routeParams={routeParams}
			isCloning={isCloning} />}
	</Heading>
	<Link style={subSx} to={`/${repo}`} >{repo}</Link>
</Base>;
