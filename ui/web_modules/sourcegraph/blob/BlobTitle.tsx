import * as React from "react";
import * as Relay from "react-relay";

import { AuthorsToggleButton } from "sourcegraph/blob/AuthorsToggleButton";
import { CommitInfoBar } from "sourcegraph/blob/CommitInfoBar/CommitInfoBar";
import { OpenInGitHubButton } from "sourcegraph/blob/OpenInGitHubButton";
import { UnsupportedLanguageAlert } from "sourcegraph/blob/UnsupportedLanguageAlert";
import { FlexContainer, Heading, PathBreadcrumb } from "sourcegraph/components";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { Features } from "sourcegraph/util/features";
import { getPathExtension, isBetaExtension, isIgnoredExtension, isSupportedExtension } from "sourcegraph/util/supportedExtensions";
import { prettifyRev } from "sourcegraph/workbench/utils";

interface Props {
	repo: string;
	path: string;
	rev: string | null;
	toast: string | null;
	toggleAuthors: () => void;
}

function basename(path: string): string {
	const base = path.split("/").pop();
	return base || path;
};

class BlobTitleComponent extends React.Component<Props & { root: GQL.IRoot }, {}> {
	render(): JSX.Element {
		const { repo, path, toggleAuthors, toast } = this.props;
		let rev = prettifyRev(this.props.rev);
		if (rev === null) {
			if (this.props.root.repository) {
				rev = this.props.root.repository.defaultBranch;
			} else {
				rev = "master";
			}
		}

		const extension = getPathExtension(path);
		const isSupported = extension ? isSupportedExtension(extension) : false;
		const isBeta = extension ? isBetaExtension(extension) : false;
		const isIgnored = extension ? isIgnoredExtension(extension) : false;
		const isRoot = path === "";

		return <div style={{ width: "100%" }}>
			<FlexContainer justify="between" items="center" style={{
				backgroundColor: colors.blueGrayD2(),
				boxShadow: `0 2px 6px 0px ${colors.black(0.2)}`,
				minHeight: layout.editorToolbarHeight,
				padding: `0 ${whitespace[3]}`,
				position: "relative",
				width: "100%",
				zIndex: 3,
			}}>
				{!isRoot &&
					[
						<FlexContainer key="left" style={{ overflow: "hidden" }}>
							<Heading style={{ display: "inline-block" }} level={6} color="white" compact={true}>
								{basename(path)}
							</Heading>
							<PathBreadcrumb
								repo={repo}
								path={path}
								rev={rev}
								linkSx={Object.assign({ color: colors.blueGrayL1() }, typography.size[7])}
								linkHoverSx={{ color: `${colors.blueGrayL3()} !important` }}
								style={{
									color: colors.blueGrayL1(),
									overflow: "hidden",
									paddingLeft: whitespace[2],
									paddingRight: whitespace[2],
									paddingTop: 2,
								}} />
						</FlexContainer>,
						<div key="right" style={{ flex: "0 0 auto" }}>
							<div style={Object.assign({
								color: "white",
								textAlign: "right",
							}, typography.size[7])}>

								{!isSupported && !isIgnored &&
									<UnsupportedLanguageAlert ext={extension} inBeta={isBeta} />
								}

								<AuthorsToggleButton shortcut="a" keyCode={65} toggleAuthors={toggleAuthors} />
								{repo.startsWith("github.com") &&
									<OpenInGitHubButton repo={repo} path={path} rev={rev} />
								}

								{toast && <div>{toast}</div>}
							</div>
						</div>
					]
				}
			</FlexContainer>
			{Features.commitInfoBar.isEnabled() && <CommitInfoBar repo={repo} rev={rev} path={path} />}
		</div>;
	}
};

const BlobTitleContainer = Relay.createContainer(BlobTitleComponent, {
	initialVariables: {
		repo: "",
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				repository(uri: $repo) {
					defaultBranch
				}
			}
		`,
	},
});

export const BlobTitle = function (props: Props): JSX.Element {
	return <Relay.RootContainer
		Component={BlobTitleContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: props,
		}}
	/>;
};
