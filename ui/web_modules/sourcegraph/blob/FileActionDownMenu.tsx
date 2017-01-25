import * as autobind from "autobind-decorator";
import * as classNames from "classnames";
import * as React from "react";
import { Router } from "sourcegraph/app/router";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { FlexContainer, Key, Menu, Popover } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { ChevronDown } from "sourcegraph/components/symbols/Primaries";
import { colors, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

const openInGitHubKeyCode = 71;
const openInGitHubKey = "G";

interface Props {
	eventProps: { repo: string, rev: string | null, path: string };
	githubURL: string;
	editorURL: string;
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
export class FileActionDownMenu extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	private githubURL(): string {
		return `${this.props.githubURL}${convertToGitHubLineNumber(this.context.router.location.hash)}`;
	}

	private fileActionEventListener(event: KeyboardEvent): void {
		const eventTarget = event.target as Node;
		if (eventTarget.nodeName === "INPUT" || isNonMonacoTextArea(eventTarget) || event.metaKey || event.ctrlKey) {
			return;
		} else if (event.keyCode === openInGitHubKeyCode || event.key === openInGitHubKey) {
			AnalyticsConstants.Events.OpenInCodeHost_Clicked.logEvent(this.props.eventProps);
			window.open(this.githubURL());
			event.preventDefault();
		}
	}

	private onViewGithubClick(): void {
		AnalyticsConstants.Events.OpenInCodeHost_Clicked.logEvent(this.props.eventProps);
		window.open(this.githubURL());
	}

	render(): JSX.Element {
		// float required to fix Firefox issue.
		return <div style={{ display: "inline-block", padding: whitespace[2], paddingRight: 0, float: "right" }}>
			<Popover left={true}>
				<FlexContainer items="center" style={{ lineHeight: "0", height: 29 }}>
					<div>View</div>
					<ChevronDown color={colors.blueGray()} style={{ marginLeft: 8, top: 0 }} />
				</FlexContainer>
				<Menu className={classNames(base.pa0, base.mr2)} style={{ width: 125 }}>
					<a onClick={this.onViewGithubClick} style={{ textAlign: "left" }} role="menu_item" target="_blank">
						View on GitHub
						<Key shortcut={"G"} style={{ marginLeft: whitespace[2], float: "right" }} />
					</a>
				</Menu>
			</Popover>
			<EventListener
				target={global.document.body}
				event="keydown"
				callback={this.fileActionEventListener} />
		</div>;
	}
};
