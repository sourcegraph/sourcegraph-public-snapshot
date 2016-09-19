// tslint:disable: typedef ordered-imports

import { Location } from "history";
import * as React from "react";
import { InjectedRouter } from "react-router";
import Helmet from "react-helmet";
import * as debounce from "lodash/debounce";
import { Editor } from "sourcegraph/editor/Editor";
import "sourcegraph/blob/BlobBackend";
import { lineRange } from "sourcegraph/blob/lineCol";
import * as Style from "sourcegraph/blob/styles/Blob.css";
import { trimRepo } from "sourcegraph/repo";
import { httpStatusCode } from "sourcegraph/util/httpStatusCode";
import { Header } from "sourcegraph/components/Header";
import { urlToBlob } from "sourcegraph/blob/routes";
import { SearchModal } from "sourcegraph/search/modal/SearchModal";

interface Props {
	repo: string;
	rev: string | null;
	commitID?: string;
	path: string;
	blob?: any;
	startLine?: number;
	startCol?: number;
	endLine?: number;
	endCol?: number;
	location: Location;
}

type State = any;

// BlobMain wraps the Editor component for the primary code view.
export class BlobMain extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: {
		router: InjectedRouter,
	};

	private _editor: monaco.editor.IStandaloneCodeEditor | null = null;
	private _editorComponent: Editor | null;

	constructor(props: Props) {
		super(props);
		this._setEditor = this._setEditor.bind(this);
		this._onKeyDownForFindInPage = this._onKeyDownForFindInPage.bind(this);
		this._onResize = debounce(this._onResize.bind(this), 300, { leading: true, trailing: true });
		this._onSelectionChange = debounce(this._onSelectionChange.bind(this), 100);
	}

	componentDidMount(): void {
		global.document.addEventListener("keydown", this._onKeyDownForFindInPage);
		global.document.addEventListener("resize", this._onResize);

		global.document.body.style.overflowY = "hidden";
	}

	componentWillUnmount(): void {
		global.document.removeEventListener("keydown", this._onKeyDownForFindInPage);
		global.document.removeEventListener("resize", this._onResize);

		global.document.body.style.overflowY = "auto";
	}

	_setEditor(editor: monaco.editor.IStandaloneCodeEditor | null): void {
		this._editor = editor;
		if (this._editor) {
			this._editor.onDidChangeCursorSelection(this._onSelectionChange);
		}
	}

	_onKeyDownForFindInPage(e: KeyboardEvent): void {
		const mac = navigator.userAgent.indexOf("Macintosh") >= 0;
		const ctrl = mac ? e.metaKey : e.ctrlKey;
		const FKey = 70;
		if (e.keyCode === FKey && ctrl) {
			if (this._editor) {
				e.preventDefault();
				(document.getElementsByClassName("inputarea")[0] as any).focus();
				this._editor.trigger("keyboard", "actions.find", {});
			}
		}
	}

	_onResize(e: Event): void {
		if (this._editor) {
			this._editor.layout();
		}
	}

	_onSelectionChange(e: monaco.editor.ICursorSelectionChangedEvent): void {
		// this is here because the api calls are coming from the find command or something else we don't want to capture
		if (e.source === "api") {
			return;
		}

		const start = e.selection.startLineNumber;
		let end = e.selection.endLineNumber;
		if (e.selection.endColumn === 1 && end === start + 1) {
			end -= 1; // if the cursor on the last line doesn't highlight anything, ignore line
		}
		const path = urlToBlob(this.props.repo, this.props.rev, this.props.path);

		if (e.selection.isEmpty()) {
			if (this._editorComponent && this._editorComponent._mouseDownOnIdent) {
				// Click handler will trigger jump-to-def.
				return;
			}
			this.setState({ userManuallyScrolledToLineViaSelection: null}, () => {
				this.context.router.replace(path);
			});
			return;
		}

		// Record that the user manually scrolled to this line so that the props change
		// (due to changing the URL hash) doesn't trigger a jerky duplicate scroll to
		// the same line.
		this.setState({ userManuallyScrolledToLineViaSelection: start }, () => {
			this.context.router.replace(`${path}#L${lineRange(start, end)}`);
		});
	}

	render(): JSX.Element | null {
		if (this.props.blob && this.props.blob.Error) {
			let msg;
			switch (this.props.blob.Error.response.status) {
				case 413:
					msg = "Sorry, this file is too large to display.";
					break;
				default:
					msg = "File is not available.";
			}
			return (
				<Header
					title={`${httpStatusCode(this.props.blob.Error)}`}
					subtitle={msg} />
			);
		}

		let title = trimRepo(this.props.repo);
		const pathParts = this.props.path ? this.props.path.split("/") : null;
		if (pathParts) {
			title = `${pathParts[pathParts.length - 1]} · ${title}`;
		}

		return (
			<div className={Style.container}>
				<SearchModal repo={this.props.repo} commitID={this.props.commitID || this.props.rev || ""} />
				<Helmet title={title} />
				{this.props.blob && typeof this.props.blob.ContentsString === "string" && <Editor
					repo={this.props.repo}
					rev={this.props.rev}
					path={this.props.path}
					contents={this.props.blob.ContentsString}
					editorRef={this._setEditor}
					ref={(c) => this._editorComponent = c}
					startLine={this.props.startLine}
					endLine={this.props.endLine}
					startCol={this.props.startCol}
					endCol={this.props.endCol} />}
			</div>
		);
	}
}
