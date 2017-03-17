import { hover } from "glamor";
import * as React from "react";

import { IDisposable } from "vs/base/common/lifecycle";

import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Button } from "sourcegraph/components";
import { GitHubLogo } from "sourcegraph/components/symbols";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { getCurrentWorkspace, onWorkspaceUpdated } from "sourcegraph/workbench/services";

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

interface State {
	revState?: { zapRev?: string, commitID?: string, branch?: string };
}

export class OpenInGitHubButton extends React.Component<Props, State> {
	disposables: IDisposable[];

	constructor(props: Props) {
		super(props);
		this.state = {
			revState: getCurrentWorkspace().revState,
		};
		this.disposables = [];
	}

	componentDidMount(): void {
		this.disposables.push(onWorkspaceUpdated(workspace => this.setState({
			revState: workspace.revState,
		})));
	}

	componentWillUnmount(): void {
		this.disposables.forEach(disposable => disposable.dispose());
	}

	openInGitHub(e: React.MouseEvent<HTMLAnchorElement> | KeyboardEvent): void {
		const commitID = this.state.revState && this.state.revState.zapRev ? this.state.revState.commitID : this.props.rev;
		const gitHubURL = `https://${this.props.repo}/blob/${commitID}/${this.props.path}`;
		Events.OpenInCodeHost_Clicked.logEvent({ repo: this.props.repo, rev: this.props.rev, path: this.props.path });
		window.open(gitHubURL);
		e.preventDefault();
	}

	keyHandler(event: KeyboardEvent): void {
		const eventTarget = event.target as Node;
		if (eventTarget.nodeName === "INPUT" || isNonMonacoTextArea(eventTarget) || event.metaKey || event.ctrlKey) {
			return;
		} else if (event.keyCode === openInGitHubKeyCode || (event.key && event.key.toUpperCase() === openInGitHubKey)) {
			this.openInGitHub(event);
		}
	}

	// float required to fix Firefox issue.
	render(): JSX.Element {
		const commitID = this.state.revState && this.state.revState.zapRev ? this.state.revState.commitID : this.props.rev;

		return <div style={{
			display: "inline-block",
			float: "right",
			marginLeft: whitespace[2],
		}}>
			<a
				href={`https://${this.props.repo}/blob/${commitID}/${this.props.path}`}
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
				callback={e => this.keyHandler(e as KeyboardEvent)} />
		</div>;
	}
}
