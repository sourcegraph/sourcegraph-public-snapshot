import * as autobind from "autobind-decorator";
import * as React from "react";
import { ApolloProvider } from "react-apollo";

import { BlobTitle } from "sourcegraph/blob/BlobTitle";
import { gqlClient } from "sourcegraph/util/gqlClient";
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
		return <ApolloProvider client={gqlClient} >
			<RouterContext>
				<BlobTitle
					repo={repo}
					rev={rev}
					path={path}
					toggleAuthors={toggleCodeLens}
				/>
			</RouterContext>
		</ApolloProvider>;
	}
}
