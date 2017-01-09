import * as autobind from "autobind-decorator";
import * as React from "react";

import { BlobStore } from "sourcegraph/blob/BlobStore";
import { BlobTitle } from "sourcegraph/blob/BlobTitle";
import { toggleCodeLens } from "sourcegraph/workbench/ConfigurationService";
import { PathSpec, RouterContext } from "sourcegraph/workbench/utils";

interface Props {
	pathspec: PathSpec;
}

@autobind
export class EditorTitle extends React.Component<Props, {}> {
	toDispose: { remove(): void };

	componentWillUpdate(): void {
		this.toDispose = BlobStore.addListener(this.forceUpdate);
	}

	componentWillUnmount(): void {
		this.toDispose.remove();
	}

	render(): JSX.Element {
		let {repo, rev, path} = this.props.pathspec;
		if (rev === "HEAD") {
			rev = null;
		}
		return <RouterContext>
			<BlobTitle
				repo={repo}
				rev={rev}
				path={path}

				toggleAuthors={toggleCodeLens}
				toast={BlobStore.toast}
				/>
		</RouterContext>;
	}
}
