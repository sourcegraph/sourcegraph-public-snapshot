import * as autobind from "autobind-decorator";
import { hover } from "glamor";
import * as React from "react";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Button, Key } from "sourcegraph/components";
import { GitHubLogo } from "sourcegraph/components/symbols";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

const openInGitHubKeyCode = 71;
const openInGitHubKey = "G";
const gitHubButtonSx = Object.assign({
	backgroundColor: colors.blueGrayD1(),
	fontSize: "inherit",
	paddingLeft: whitespace[2],
	paddingRight: whitespace[2],
}, typography.size[7]);

interface Props {
	repo: string;
	rev: string | null;
	path: string;
}

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
export class OpenInGitHubButton extends React.Component<Props, {}> {
	private githubURL(): string {
		return `https://${this.props.repo}/blob/${this.props.rev}/${this.props.path}${convertToGitHubLineNumber(window.location.hash)}`; // We do not sync the line-number URL hash with react router due to UI issues. Use window.location as a source of truth instead.
	}

	private openInGitHubKeyHandler(event: KeyboardEvent): void {
		const { repo, rev, path } = this.props;
		const eventTarget = event.target as Node;
		if (eventTarget.nodeName === "INPUT" || isNonMonacoTextArea(eventTarget) || event.metaKey || event.ctrlKey) {
			return;
		} else if (event.keyCode === openInGitHubKeyCode || event.key === openInGitHubKey) {
			AnalyticsConstants.Events.OpenInCodeHost_Clicked.logEvent({ repo, rev, path });
			window.open(this.githubURL());
			event.preventDefault();
		}
	}

	private onViewGithubClick(): void {
		const { repo, rev, path } = this.props;
		AnalyticsConstants.Events.OpenInCodeHost_Clicked.logEvent({ repo, rev, path });
		window.open(this.githubURL());
	}

	render(): JSX.Element {
		// float required to fix Firefox issue.
		return <div style={{
			display: "inline-block",
			float: "right",
			padding: whitespace[1],
			paddingLeft: whitespace[2],
			paddingRight: 0,
		}}>
			<a
				target="new"
				onClick={this.onViewGithubClick}
				{ ...layout.hide.sm }>
				<Button
					size="small"
					style={gitHubButtonSx}
					{...hover({ backgroundColor: `${colors.blueGray()} !important` }) }>
					<GitHubLogo width={16} style={{ marginRight: whitespace[2] }} />
					View on GitHub <Key shortcut={openInGitHubKey} style={{ marginLeft: whitespace[2] }} />
				</Button>
			</a>
			<EventListener
				target={global.document.body}
				event="keydown"
				callback={this.openInGitHubKeyHandler} />
		</div>;
	}
};
