import * as React from "react";
import { PathBreadcrumb } from "sourcegraph/components/PathBreadcrumb";

interface Props {
	params: any;
	style?: React.CSSProperties;
};

export function RepoNavContext({params, style}: Props): JSX.Element {
	const path = Array.isArray(params.splat) ? params.splat[1] : null;
	const repo = Array.isArray(params.splat) ? params.splat[0] : params.splat;
	// on the root of the tree, splat is a string

	return <PathBreadcrumb toFile={false} repo={repo} path={path} style={style} rev={null} />;
}
