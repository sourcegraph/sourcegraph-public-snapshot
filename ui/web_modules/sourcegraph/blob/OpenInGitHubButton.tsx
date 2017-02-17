import { hover } from "glamor";
import * as React from "react";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Button } from "sourcegraph/components";
import { GitHubLogo } from "sourcegraph/components/symbols";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { convertToGitHubLineNumber } from "sourcegraph/util/convertToGitHubLineNumber";

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

export function OpenInGitHubButton({ repo, rev, path }: Props): JSX.Element {

	function openInGitHub(e: React.MouseEvent<HTMLAnchorElement> | KeyboardEvent): void {
		const lineNumber = convertToGitHubLineNumber(window.location.hash);
		const gitHubURL = `https://${repo}/blob/${rev}/${path}${lineNumber}`;
		Events.OpenInCodeHost_Clicked.logEvent({ repo, rev, path });
		window.open(gitHubURL);
		e.preventDefault();
	}

	function keyHandler(event: KeyboardEvent): void {
		const eventTarget = event.target as Node;
		if (eventTarget.nodeName === "INPUT" || isNonMonacoTextArea(eventTarget) || event.metaKey || event.ctrlKey) {
			return;
		} else if (event.keyCode === openInGitHubKeyCode || (event.key && event.key.toUpperCase() === openInGitHubKey)) {
			openInGitHub(event);
		}
	}

	// float required to fix Firefox issue.
	return <div style={{
		display: "inline-block",
		float: "right",
		padding: whitespace[1],
		paddingLeft: whitespace[2],
		paddingRight: 0,
	}}>
		<a
			href={`https://${repo}/blob/${rev}/${path}`}
			onClick={openInGitHub}
			{ ...layout.hide.sm }>
			<Button
				size="small"
				style={gitHubButtonSx}
				{...hover({ backgroundColor: `${colors.blueGray()} !important` }) }>
				<GitHubLogo width={16} style={{ marginRight: whitespace[2] }} />
				View on GitHub
			</Button>
		</a>
		<EventListener
			target={global.document.body}
			event="keydown"
			callback={keyHandler} />
	</div>;
}
