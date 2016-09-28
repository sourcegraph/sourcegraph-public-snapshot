// tslint:disable: typedef ordered-imports

import { Location } from "history";
import * as React from "react";
import { InjectedRouter } from "react-router";
import Helmet from "react-helmet";
import * as debounce from "lodash/debounce";
import { Editor } from "sourcegraph/editor/Editor";
import { EditorComponent } from "sourcegraph/editor/EditorComponent";
import "sourcegraph/blob/BlobBackend";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import * as Style from "sourcegraph/blob/styles/Blob.css";
import { trimRepo } from "sourcegraph/repo";
import { urlToBlob } from "sourcegraph/blob/routes";
import {IEditorOpenedEvent} from "sourcegraph/editor/EditorService";
import {ChromeExtensionToast} from "sourcegraph/components/ChromeExtensionToast";
import {OnboardingModals} from "sourcegraph/components/OnboardingModals";
import {URI} from "sourcegraph/core/uri";

type Props = {
	repo: string;
	rev: string | null;
	commitID?: string;
	path: string;
	startLine?: number;
	startCol?: number;
	endLine?: number;
	endCol?: number;
	location: Location;

	// TODO(sqs): now that BlobMain no longer uses the blob directly
	// (the EditorService fetches it), we can save on a network RTT by
	// eliminating the WantFile dispatch.
}

// BlobMain wraps the Editor component for the primary code view.
export class BlobMain extends React.Component<Props, any> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: {
		router: InjectedRouter,
	};

	private _editor?: Editor;
	private _shortCircuitURLNavigationOnEditorOpened: boolean = false;

	constructor(props: Props) {
		super(props);
		this._setEditor = this._setEditor.bind(this);
		this._onKeyDownForFindInPage = this._onKeyDownForFindInPage.bind(this);
		this._onResize = debounce(this._onResize.bind(this), 300, { leading: true, trailing: true });
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

	componentWillReceiveProps(nextProps: Props): void {
		if (this._editor) {
			this._editorPropsChanged(this.props, nextProps);
		}
	}

	_setEditor(editor: Editor | null): void {
		this._editor = editor || undefined;
		if (this._editor) {
			this._editorPropsChanged(null, this.props);
			this._editor.onDidOpenEditor(e => this._onEditorOpened(e));
		}
	}

	_editorPropsChanged(prevProps: Props | null, nextProps: Props): void {
		if (!this._editor) {
			throw new Error("editor is not ready");
		}
		if (!prevProps || (prevProps.repo !== nextProps.repo || prevProps.rev !== nextProps.rev || prevProps.commitID !== nextProps.commitID || prevProps.path !== nextProps.path || prevProps.startLine !== nextProps.startLine || prevProps.startCol !== nextProps.startCol || prevProps.endLine !== nextProps.endLine || prevProps.endCol !== nextProps.endCol)) {
			if (nextProps.commitID) {
				// Use absolute commit IDs for the editor model URI.
				const uri = URI.pathInRepo(nextProps.repo, nextProps.commitID, nextProps.path);

				let range: monaco.IRange;
				if (typeof nextProps.startLine === "number") {
					const rop = RangeOrPosition.fromOneIndexed(nextProps.startLine, nextProps.startCol, nextProps.endLine, nextProps.endCol);
					range = rop.toMonacoRangeAllowEmpty();
				} else {
					// Without a start line, set the cursor to start at the beginning of the document.
					range = RangeOrPosition.fromOneIndexed(1).toMonacoRangeAllowEmpty();
				}

				// If you are new to this code, you may be confused how this method interacts with _onEditorOpened.
				// You may also wonder how _shortCircuitURLNavigationOnEditorOpened works. Here's an explanation:
				//
				// Calling this._editor.setInput() below will indirectly invoke _onEditorOpened. The other way
				// _onEditorOpened is invoked is through a jump-to-def. When a user does a jump-to-def, we need to
				// update the URL so that React/flux will fetch blob contents, etc. Doing this updates props.location,
				// and therefore this method to be invoked.
				//
				// We have the following cases to handle:
				// - initial page load: starts with _editorPropsChanged, we tell monaco where to move the cursor,
				///  (the second argument to this._editor.setInput below) and do not need to update the URL
				// - jump-to-def: starts with _onEditorOpened, which calls router.push(url) and eventually invokes this
				//   method (at which point, we've already updated the URL and don't need to do so again)
				// - browser "back": starts with _editorPropsChanged (since props.location is updated), we simply
				//   tell monaco which uri to open and at what range. This is determined entirely by nextProps.location
				//   and we don't need _onEditorOpened to update the URL (the browser already did so)
				//
				// Therefore, whenever we invoke _onEditorOpened from _editorPropsChanged, we NEVER update URL.
				// Jump-to-def starts with _onEditorOpened, and starting from that code path we ALWAYS update URL.

				this._shortCircuitURLNavigationOnEditorOpened = true;
				this._editor.setInput(uri, range).then(() => {
					// Always reset this bit after opening the editor.
					this._shortCircuitURLNavigationOnEditorOpened = false;
				});
			}
		}
	}

	_onKeyDownForFindInPage(e: KeyboardEvent): void {
		// TODO(sqs): can make this unnecessary?
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

	_onEditorOpened(e: IEditorOpenedEvent): void {
		if (this._shortCircuitURLNavigationOnEditorOpened) {
			return;
		}

		let {repo, rev, path} = URI.repoParams(e.model.uri);

		// If same repo, use the rev from the URL, so that we don't
		// change the address bar around a lot (bad UX).
		//
		// TODO(sqs): this will break true cross-rev same-repo jumps.
		if (repo === this.props.repo) {
			rev = this.props.rev;
		}

		let url = urlToBlob(repo, rev, path);

		const sel = e.editor.getSelection();
		if (!sel.isEmpty() || sel.startLineNumber !== 1) {
			let startCol: number | undefined = sel.startColumn;
			let endCol: number | undefined = sel.endColumn;
			if (e.model.getLineMinColumn(sel.startLineNumber) === startCol) {
				startCol = undefined;
			}
			if (e.model.getLineMaxColumn(sel.endLineNumber) === endCol) {
				endCol = undefined;
			}

			// HACK
			if (endCol <= 1 && !startCol) {
				endCol = undefined;
			}

			const r = RangeOrPosition.fromOneIndexed(sel.startLineNumber, startCol, sel.endLineNumber, endCol);
			url = `${url}#L${r.toString()}`;
		}

		this.context.router.push(url);
	}

	render(): JSX.Element | null {
		let title = trimRepo(this.props.repo);
		const pathParts = this.props.path ? this.props.path.split("/") : null;
		if (pathParts) {
			title = `${pathParts[pathParts.length - 1]} · ${title}`;
		}

		return (
			<div className={Style.container}>
				<Helmet title={title} />
				<ChromeExtensionToast location={this.props.location}/>
				<OnboardingModals location={this.props.location}/>
				<EditorComponent editorRef={this._setEditor} style={{ display: "flex", flex: "auto", width: "100%" }} />
			</div>
		);
	}
}
