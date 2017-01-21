import * as autobind from "autobind-decorator";
import * as classNames from "classnames";
import * as React from "react";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { FlexContainer, Key, Menu, Popover } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { ChevronDown } from "sourcegraph/components/symbols/Primaries";
import { colors, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

const openInGitHubKeyCode: number = 71;
const openInGitHubKey: string = "G";

interface Props {
	eventProps: { repo: string, rev: string | null, path: string };
	githubURL: string;
	editorURL: string;
}

@autobind
export class FileActionDownMenu extends React.Component<Props, {}> {
	private fileActionEventListener(event: KeyboardEvent): void {
		const eventTarget = event.target as Node;
		if (eventTarget.nodeName === "INPUT" || isNonMonacoTextArea(eventTarget) || event.metaKey || event.ctrlKey) {
			return;
		} else if (event.keyCode === openInGitHubKeyCode || event.key === openInGitHubKey) {
			AnalyticsConstants.Events.OpenInCodeHost_Clicked.logEvent(this.props.eventProps);
			window.open(this.props.githubURL);
			event.preventDefault();
		}
	}

	render(): JSX.Element {
		return <div style={{ display: "inline-block", padding: whitespace[2], paddingRight: 0 }}>
			<Popover left={true}>
				<FlexContainer items="center" style={{ lineHeight: "0", height: 29 }}>
					<div>View</div>
					<ChevronDown color={colors.blueGray()} style={{ marginLeft: 8, top: 0 }} />
				</FlexContainer>
				<Menu className={classNames(base.pa0, base.mr2)} style={{ width: 125 }}>
					<a href={this.props.githubURL} onClick={() => AnalyticsConstants.Events.OpenInCodeHost_Clicked.logEvent(this.props.eventProps)} style={{ textAlign: "left" }} role="menu_item">
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
