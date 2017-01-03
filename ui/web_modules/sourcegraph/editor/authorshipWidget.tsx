import { merge } from "glamor";
import * as React from "react";
import * as ReactDOM from "react-dom";
import { Heading, Panel, User } from "sourcegraph/components";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance } from "sourcegraph/editor/Editor";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { ICodeEditor, IEditorMouseEvent } from "vs/editor/browser/editorBrowser";
import { IContentWidget, IContentWidgetPosition } from "vs/editor/browser/editorBrowser.d";
import { CommandsRegistry } from "vs/platform/commands/common/commands";
import { ServicesAccessor } from "vs/platform/instantiation/common/instantiation";

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
				<a>
					<Close color={colors.coolGray3()} width={14} />
				</a>
			</div>
			<Heading level={6} style={{ marginTop: whitespace[3] }}>{message}</Heading>
			<a onClick={() => openCommit(props)} {...merge(
				{ color: colors.coolGray3(), fontFamily: typography.fontStack.code },
				typography.small,
			) }>{rev.substr(0, 6)}</a>
		</div>
	</Panel>;
};

export class AuthorshipWidget implements IContentWidget {
	private domNode: HTMLElement;

	constructor(private blame: GQL.IHunk, private element: JSX.Element) {
		//
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
		return {
			position: {
				lineNumber: this.blame.startLine,
				column: this.blame.startByte,
			},
			preference: [0],
		};
	};
}

let authorWidget: AuthorshipWidget;

function showAuthorshipPopup(accessor: ServicesAccessor, args: GQL.IHunk): void {
	removeWidget();

	const editor = getEditorInstance();

	const model = editor.getModel();
	args.startByte = model.getLineFirstNonWhitespaceColumn(args.startLine);
	const {repo, rev} = URIUtils.repoParams(editor.getModel().uri);

	const authorshipCodeLensElement = <CodeLensAuthorWidget blame={args} repo={repo} rev={rev || ""} onClose={removeWidget} />;
	authorWidget = new AuthorshipWidget(args, authorshipCodeLensElement);

	addListeners(editor);
	editor.addContentWidget(authorWidget);
	AnalyticsConstants.Events.CodeLensCommit_Clicked.logEvent(args);
}

CommandsRegistry.registerCommand("codelens.authorship.commit", showAuthorshipPopup);

export function removeWidget(): void {
	if (!authorWidget) {
		return;
	}
	const editor = getEditorInstance();
	editor.removeContentWidget(authorWidget);
}

function addListeners(editor: ICodeEditor): void {
	editor.onMouseUp((e: IEditorMouseEvent) => {
		if (e.target.detail !== AuthorshipWidgetID) {
			removeWidget();
		}
	});
}
