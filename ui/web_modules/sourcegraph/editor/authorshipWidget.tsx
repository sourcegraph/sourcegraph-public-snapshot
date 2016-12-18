import { merge } from "glamor";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { Heading, Panel, User } from "sourcegraph/components";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { IContentWidget, IContentWidgetPosition } from "vs/editor/browser/editorBrowser.d";

export const AuthorshipWidgetID = "contentwidget.authorship.widget";

interface Props {
	blame: GQL.IHunk;
	repo: string;
	rev: string;
	onClose: () => void;
}

function openCommit(props: Props): void {
	let url = `${props.repo}/commit/${props.blame.rev}#diff-${props.rev}`;
	AnalyticsConstants.Events.CodeLensCommitRedirect_Clicked.logEvent({ url });
	window.open(`https://${url}`, "_newtab");
}

export function CodeLensAuthorWidget(props: Props): JSX.Element {
	const { gravatarHash, name, rev, message } = props.blame;
	return <Panel style={{ minWidth: 320, color: colors.text() }} hoverLevel="low">
		<div style={{ margin: whitespace[3] }}>
			<User
				avatar={`https://secure.gravatar.com/avatar/${gravatarHash}?s=128&d=retro`}
				nickname={name}
				simple={true} />
			<div onClick={props.onClose.bind(this)} style={{
				padding: whitespace[1],
				position: "absolute",
				right: whitespace[3],
				top: whitespace[3],
			}}>
				<Close color={colors.coolGray3()} width={14} />
			</div>
			<Heading level={6} style={{ marginTop: whitespace[3] }}>{message}</Heading>
			<a onClick={() => openCommit(props)} {...merge(
				{ color: colors.coolGray3(), fontFamily: typography.fontStack.code },
				typography.small,
			) }>Commit {rev.substr(0, 6)}</a>
		</div>
	</Panel>;
};

export class AuthorshipWidget implements IContentWidget {
	private data: GQL.IHunk;
	private domNode: HTMLElement;
	private element: JSX.Element;

	constructor(data: GQL.IHunk, element: JSX.Element) {
		this.data = data;
		this.element = element;
	};

	getId(): string {
		return AuthorshipWidgetID;
	}

	getDomNode(): HTMLElement {
		if (!this.domNode) {
			let node = document.createElement("div");
			ReactDOM.render(this.element, node);
			this.domNode = node;
		}
		return this.domNode;
	}

	getPosition(): IContentWidgetPosition {
		const {data} = this;
		return {
			position: {
				lineNumber: data.startLine,
				column: data.startByte,
			},
			preference: [0],
		};
	};
}
