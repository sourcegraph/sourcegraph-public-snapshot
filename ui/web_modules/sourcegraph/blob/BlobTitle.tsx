import * as React from "react";
import * as Relay from "react-relay";

import { AuthorsToggleButton } from "sourcegraph/blob/AuthorsToggleButton";
import { CommitInfoBar } from "sourcegraph/blob/CommitInfoBar/CommitInfoBar";
import { FileActionDownMenu } from "sourcegraph/blob/FileActionDownMenu";
import { UnsupportedLanguageAlert } from "sourcegraph/blob/UnsupportedLanguageAlert";
import { FlexContainer, Heading, PathBreadcrumb } from "sourcegraph/components";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { Features } from "sourcegraph/util/features";
import { getPathExtension, isIgnoredExtension, isSupportedExtension } from "sourcegraph/util/supportedExtensions";

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

function convertToGitHubLineNumber(hash: string): string {
	if (!hash || !hash.startsWith("#L")) {
		return "";
	}
	let lines: string[] = hash.split("#L");
	if (lines.length !== 2) {
		return "";
	}
	lines = lines[1].split("-");
	if (lines.length === 1) {
		// single line
		return `#L${lines[0]}`;
	} else if (lines.length === 2) {
		// line range
		return `#L${lines[0]}-L${lines[1]}`;
	}
	return "";
}

class BlobTitleComponent extends React.Component<Props & { root: GQL.IRoot }, {}> {
	render(): JSX.Element {
		const {repo, path, toggleAuthors, toast } = this.props;
		let rev = this.props.rev;
		if (rev === null) {
			if (this.props.root.repository) {
				rev = this.props.root.repository.defaultBranch;
			} else {
				rev = "master";
			}
		}

		const extension = getPathExtension(path);
		const isSupported = extension ? isSupportedExtension(extension) : false;
		const isIgnored = extension ? isIgnoredExtension(extension) : false;
		const gitHubURL = `https://${repo}/blob/${rev}/${path}${convertToGitHubLineNumber(window.location.hash)}`;

		return <div style={{ width: "100%" }}>
			<FlexContainer justify="between" items="center" wrap={true} style={{
				backgroundColor: colors.blueGrayD2(),
				boxShadow: `0 2px 6px 0px ${colors.black(0.2)}`,
				height: layout.editorToolbarHeight,
				padding: `${whitespace[2]} ${whitespace[3]}`,
				width: "100%",
			}}>
				<div>
					<Heading style={{ display: "inline-block" }} level={6} color="white" compact={true}>
						{basename(path)}
					</Heading>
					<PathBreadcrumb
						repo={repo}
						path={path}
						rev={rev}
						linkSx={Object.assign({ color: colors.blueGrayL1() }, typography.size[7])}
						linkHoverSx={{ color: `${colors.blueGrayL3()} !important` }}
						style={{ display: "inline-block", marginBottom: 0, paddingLeft: whitespace[2] }} />
				</div>
				<div>
					<div style={Object.assign({
						color: "white",
						flex: "1 1",
						paddingRight: whitespace[1],
						textAlign: "right",
					}, typography.size[7])}>

						{Features.authorsToggle.isEnabled() &&
							<AuthorsToggleButton shortcut="a" keyCode={65} toggleAuthors={toggleAuthors} />
						}

						<FileActionDownMenu eventProps={{ repo, rev, path }} githubURL={gitHubURL} editorURL={path} />

						{!isSupported && !isIgnored &&
							<UnsupportedLanguageAlert ext={extension} style={{ marginLeft: whitespace[3] }} />
						}

						{toast && <div>{toast}</div>}
					</div>
				</div>
			</FlexContainer>
			{Features.projectWow.isEnabled() && <CommitInfoBar repo={repo} rev={rev} path={path} />}
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
