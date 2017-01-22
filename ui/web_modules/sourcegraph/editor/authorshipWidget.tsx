import { merge } from "glamor";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { ICodeEditor, IEditorMouseEvent } from "vs/editor/browser/editorBrowser";
import { IContentWidget, IContentWidgetPosition } from "vs/editor/browser/editorBrowser.d";
import { CommandsRegistry } from "vs/platform/commands/common/commands";
import { ServicesAccessor } from "vs/platform/instantiation/common/instantiation";

import { Heading, Panel, User } from "sourcegraph/components";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance } from "sourcegraph/editor/Editor";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

export const AuthorshipWidgetID = "contentwidget.authorship.widget";

interface Props {
	blame: GQL.IHunk;
	repo: string;
	rev: string;
	onClose: () => void;
}

export function CodeLensAuthorWidget(props: Props): JSX.Element {
	if (!props.blame.author || !props.blame.author.person) {
		return <div />;
	}

	const { rev, message } = props.blame;
	const { gravatarHash, name } = props.blame.author.person;
	const commitURL = `http://${props.repo}/commit/${props.blame.rev}#diff-${props.rev}`;

	return <Panel hoverLevel="low" style={{
		minWidth: 320,
		color: colors.text(),
		marginLeft: -2,
		marginTop: 4,
	}}>
		<div style={{ margin: whitespace[3] }}>
			<User
				avatar={`https://secure.gravatar.com/avatar/${gravatarHash}?s=128&d=retro`}
				nickname={name}
				simple={true} />
			<div onClick={props.onClose} style={{
				cursor: "pointer",
				padding: 1,
				position: "absolute",
				right: whitespace[3],
				top: whitespace[3],
			}}>
				<Close color={colors.blueGray()} width={14} />
			</div>
			<Heading level={6} style={{ marginTop: whitespace[3] }}>{message}</Heading>
			<a
				href={commitURL}
				onClick={() => AnalyticsConstants.Events.CodeLensCommitRedirect_Clicked.logEvent({ commitURL })}
				target="_blank"
				{...merge({
					color: colors.blueGray(),
					fontFamily: typography.fontStack.code
				}, typography.small,
				) }>
				Commit {rev.substr(0, 6)}
			</a>
		</div>
	</Panel>;
};

export class AuthorshipWidget implements IContentWidget {
	private domNode: HTMLElement;

	constructor(
		public blame: GQL.IHunk,
		private element: JSX.Element,
	) {
		//
	};

	getId(): string {
		return AuthorshipWidgetID;
	}

	getDomNode(): HTMLElement {
		if (!this.domNode) {
			let node = document.createElement("div");
			node.style.marginTop = "-20px";
			ReactDOM.render(this.element, node);
			this.domNode = node;
		}
		return this.domNode;
	}

	getPosition(): IContentWidgetPosition {
		return {
			position: {
				lineNumber: this.blame.startLine,
				column: this.blame.startByte,
			},
			preference: [1, 0],
		};
	};
}

let authorWidget: AuthorshipWidget | null = null;

function showAuthorshipPopup(accessor: ServicesAccessor, blame: GQL.IHunk): void {
	if (authorWidget && blame === authorWidget.blame) {
		removeWidget();
		return;
	}
	removeWidget();

	const editor = getEditorInstance();

	const model = editor.getModel();
	blame.startByte = model.getLineFirstNonWhitespaceColumn(blame.startLine);
	const {repo, rev} = URIUtils.repoParams(editor.getModel().uri);

	const authorshipCodeLensElement = <CodeLensAuthorWidget blame={blame} repo={repo} rev={rev || ""} onClose={removeWidget} />;
	authorWidget = new AuthorshipWidget(blame, authorshipCodeLensElement);

	addListeners(editor);
	editor.addContentWidget(authorWidget);
	AnalyticsConstants.Events.CodeLensCommit_Clicked.logEvent(blame);
}

CommandsRegistry.registerCommand("codelens.authorship.commit", showAuthorshipPopup);

export function removeWidget(): void {
	if (!authorWidget) {
		return;
	}
	const editor = getEditorInstance();
	editor.removeContentWidget(authorWidget);
	authorWidget = null;
}

function addListeners(editor: ICodeEditor): void {
	editor.onMouseUp((e: IEditorMouseEvent) => {
		if (!e.target.detail
			|| e.target.detail !== AuthorshipWidgetID
			&& !e.target.detail.startsWith("codeLens")) {
			removeWidget();
		}
	});
}
