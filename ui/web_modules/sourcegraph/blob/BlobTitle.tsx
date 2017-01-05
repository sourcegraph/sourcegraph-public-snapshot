import * as autobind from "autobind-decorator";
import * as React from "react";

import { RouteParams } from "sourcegraph/app/routeParams";
import { AuthorsToggleButton } from "sourcegraph/blob/AuthorsToggleButton";
import { UnsupportedLanguageAlert } from "sourcegraph/blob/UnsupportedLanguageAlert";
import { Button, FlexContainer, Heading, PathBreadcrumb } from "sourcegraph/components";
import { GitHubLogo } from "sourcegraph/components/symbols";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { RevSwitcher } from "sourcegraph/repo/RevSwitcher";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { Features } from "sourcegraph/util/features";
import { getPathExtension, isIgnoredExtension, isSupportedExtension } from "sourcegraph/util/supportedExtensions";

interface Props {
	repo: string;
	path: string;
	rev: string | null;
	routes: Object[];
	routeParams: RouteParams;
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
		const {repo, path, rev, routes, routeParams, toggleAuthors, toast } = this.props;

		const extension = getPathExtension(path);
		const isSupported = extension ? isSupportedExtension(extension) : false;
		const isIgnored = extension ? isIgnoredExtension(extension) : false;
		const gitHubURL = `https://${repo}/blob/${rev}/${path}${convertToGitHubLineNumber(window.location.hash)}`;

		return <FlexContainer justify="between" items="center" wrap={true} style={{
			backgroundColor: colors.coolGray1(),
			boxShadow: `0 2px 6px 0px ${colors.black(0.2)}`,
			height: layout.editorToolbarHeight,
			zIndex: 1,
			padding: `${whitespace[2]} ${whitespace[3]}`,
			width: "100%",
		}}>
			<div>
				<Heading level={6} color="white" compact={true}>
					{basename(path)}
					<RevSwitcher
						repo={repo}
						rev={rev}
						routes={routes}
						routeParams={routeParams}
						style={{ marginLeft: whitespace[1] }} />
				</Heading>
				<PathBreadcrumb
					repo={repo}
					path={path}
					rev={rev}
					linkSx={Object.assign({ color: colors.coolGray3() }, typography.size[7])}
					linkHoverSx={{ color: `${colors.coolGray4()} !important` }}
					style={{ marginBottom: 0 }} />
			</div>

			<div style={Object.assign({
				color: "white",
				flex: "1 1",
				paddingRight: whitespace[1],
				textAlign: "right",
			}, typography.size[7])}>

				<a href={gitHubURL} onClick={() => AnalyticsConstants.Events.OpenInCodeHost_Clicked.logEvent({ repo, rev, path })} { ...layout.hide.sm }>
					<Button size="small" style={{
						backgroundColor: "transparent",
						fontSize: "inherit",
						marginRight: whitespace[3],
						paddingLeft: whitespace[2],
						paddingRight: whitespace[2],
					}}>
						<GitHubLogo width={16} style={{
							marginRight: whitespace[2],
							verticalAlign: "text-top",
						}} />
						View on GitHub
					</Button>
				</a>

				{Features.authorsToggle.isEnabled() &&
					<AuthorsToggleButton shortcut="a" keyCode={65} toggleAuthors={toggleAuthors} />
				}

				{!isSupported && !isIgnored &&
					<UnsupportedLanguageAlert ext={extension} style={{ marginLeft: whitespace[3] }} />
				}

				{toast && <div>{toast}</div>}

			</div>
		</FlexContainer>;
	}
};
