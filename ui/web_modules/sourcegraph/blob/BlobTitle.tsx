import * as React from "react";
import { RouteParams } from "sourcegraph/app/routeParams";
import { UnsupportedLanguageAlert } from "sourcegraph/blob/UnsupportedLanguageAlert";
import { Button, FlexContainer, Heading, PathBreadcrumb, ToggleSwitch } from "sourcegraph/components";
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
	toggleAuthors: (visible: boolean) => void;
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

export function BlobTitle({
	toggleAuthors,
	repo,
	path,
	rev,
	routes,
	routeParams,
	toast,
}: Props): JSX.Element {
	const extension = getPathExtension(path);
	const isSupported = extension ? isSupportedExtension(extension) : false;
	const isIgnored = extension ? isIgnoredExtension(extension) : false;
	// Tech debt: BlobMain won't pass new location on line clicks, so use window.location.
	// We must register an explicit onClick handler on the GitHub anchor link to detect line hash changes.
	const gitHubURL = () => `https://${repo}/blob/${rev}/${path}${convertToGitHubLineNumber(window.location.hash)}`;

	function goToGitHub(e: React.MouseEvent<any>): void {
		e.preventDefault();
		AnalyticsConstants.Events.OpenInCodeHost_Clicked.logEvent({ repo, rev, path });
		window.location.href = gitHubURL();
	}

	return <FlexContainer justify="between" items="center" wrap={true} style={{
		backgroundColor: colors.coolGray1(),
		boxShadow: `0 2px 6px 0px ${colors.black(0.2)}`,
		minHeight: layout.editorToolbarHeight,
		zIndex: 1,
		padding: `${whitespace[2]} ${whitespace[3]}`,
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

			<a href={gitHubURL()} onClick={(e) => goToGitHub(e)} { ...layout.hide.sm }>
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

			{Features.authorsToggle.isEnabled() && <div
				style={{ display: "inline-block" }}
				{ ...layout.hide.sm}>
				<ToggleSwitch
					labels={true}
					size="small"
					defaultChecked={Features.codeLens.isEnabled()}
					onChange={(visible) => toggleAuthors(visible)}
					style={{ marginRight: whitespace[1], position: "relative", top: -2 }}
					/> <strong>Show authors</strong>
			</div>}

			{!isSupported && !isIgnored && <UnsupportedLanguageAlert ext={extension} style={{ marginLeft: whitespace[3] }} />}
			{toast && <div>{toast}</div>}

		</div>
	</FlexContainer>;
};
