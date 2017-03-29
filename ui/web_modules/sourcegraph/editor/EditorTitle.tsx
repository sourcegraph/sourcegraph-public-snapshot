import * as autobind from "autobind-decorator";
import * as React from "react";

import { BlobTitle } from "sourcegraph/blob/BlobTitle";
import { toggleCodeLens } from "sourcegraph/workbench/ConfigurationService";
import { PathSpec, RouterContext } from "sourcegraph/workbench/utils";

interface Props {
	pathspec: PathSpec;
}

@autobind
export class EditorTitle extends React.Component<Props, {}> {
	render(): JSX.Element {
		let { repo, rev, path } = this.props.pathspec;
		if (rev === "HEAD") {
			rev = null;
		}
		return <RouterContext>
			<BlobTitle
				repo={repo}
				rev={rev}
				path={path}
				toggleAuthors={toggleCodeLens}
			/>
		</RouterContext>;
	}
}
