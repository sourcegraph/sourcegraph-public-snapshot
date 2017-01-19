import * as autobind from "autobind-decorator";
import * as React from "react";

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

@autobind
export class BlobTitle extends React.Component<Props, {}> {
	constructor(props: Props) {
		super(props);
		this.onEditorLineSelected = this.onEditorLineSelected.bind(this);
	}

	componentDidMount(): void {
		window.document.addEventListener("editorLineSelected", this.onEditorLineSelected);
	}

	componentWillUnmount(): void {
		window.document.removeEventListener("editorLineSelected", this.onEditorLineSelected);
	}

	onEditorLineSelected(): void {
		// This component depends on knowing the URL, which may change via history.replaceState when editor cursor
		// position/selection is updated; when it does, redraw the component to update the GitHub URL line selection.
		this.forceUpdate();
	}

	render(): JSX.Element {
		const {repo, path, toggleAuthors, toast } = this.props;
		const rev = this.props.rev || "master";

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
