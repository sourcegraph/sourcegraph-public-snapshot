import * as React from "react";
import { gql, graphql } from "react-apollo";

import { AuthorsToggleButton } from "sourcegraph/blob/AuthorsToggleButton";
import { OpenInGitHubButton } from "sourcegraph/blob/OpenInGitHubButton";
import { UnsupportedLanguageAlert } from "sourcegraph/blob/UnsupportedLanguageAlert";
import { FlexContainer, Heading, PathBreadcrumb } from "sourcegraph/components";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { getPathExtension, isBetaExtension, isIgnoredExtension, isSupportedExtension } from "sourcegraph/util/supportedExtensions";
import { prettifyRev } from "sourcegraph/workbench/utils";

interface Props {
	repo: string;
	path: string;
	rev: string | null;
	toggleAuthors: () => void;
	loading?: boolean;
	root?: GQL.IRoot;
}

function basename(path: string): string {
	const base = path.split("/").pop();
	return base || path;
};

function BlobTitleComponent(props: Props): JSX.Element {
	if (props.loading || !props.root) {
		return <div />;
	}

	const { repo, path, toggleAuthors } = props;
	let rev = prettifyRev(props.rev);
	if (rev === null) {
		if (props.root.repository) {
			rev = props.root.repository.defaultBranch;
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
			minHeight: layout.EDITOR_TITLE_HEIGHT,
			padding: `0 ${whitespace[3]}`,
			position: "relative",
			width: "100%",
			zIndex: 3,
		}}>
			{!isRoot &&
				[
					<FlexContainer key="left" style={{ overflow: "hidden" }}>
						<Heading style={{ display: "inline-block", whiteSpace: "nowrap" }} level={6} color="white" compact={true}>
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
						</div>
					</div>
				]
			}
		</FlexContainer>
	</div>;
};

export const BlobTitle = graphql(gql`
	query BlobTitle($repo: String!) {
		root {
			repository(uri: $repo) {
				defaultBranch
			}
		}
	}`, {
		props: props => ({ ...props.ownProps, ...props.data }),
		options: props => ({ variables: { repo: props.repo } }),
	})(BlobTitleComponent);
