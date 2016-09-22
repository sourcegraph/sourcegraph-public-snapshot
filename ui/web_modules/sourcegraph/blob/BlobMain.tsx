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
import { SearchModal } from "sourcegraph/search/modal/SearchModal";
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
	private _suppressNavigationOnEditorOpened: boolean = false;

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

				let range: monaco.IRange | undefined;
				if (typeof nextProps.startLine === "number") {
					const rop = RangeOrPosition.fromOneIndexed(nextProps.startLine, nextProps.startCol, nextProps.endLine, nextProps.endCol);
					if (rop) {
						range = rop.toMonacoRangeAllowEmpty();
					}
				}

				this._suppressNavigationOnEditorOpened = Boolean(prevProps && prevProps.location !== nextProps.location && nextProps.location.action === "POP" && !(prevProps.repo === nextProps.repo && prevProps.rev === nextProps.rev && prevProps.path === nextProps.path));
				this._editor.setInput(uri, this._suppressNavigationOnEditorOpened ? undefined : range).then(() => {
					this._suppressNavigationOnEditorOpened = false;
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
		if (this._suppressNavigationOnEditorOpened) {
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

		// TODO(sqs): There is still some glitchiness with
		// back/forward. For example, if you are in a file and jump to
		// another def in the file and then go back, you don't go back
		// to the "mark" (the point from which you jumped to the def);
		// you go back to the previous def you had jumped to.

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
				<SearchModal repo={this.props.repo} commitID={this.props.commitID} rev={this.props.rev}/>
				<Helmet title={title} />
				<ChromeExtensionToast location={this.props.location}/>
				<OnboardingModals location={this.props.location}/>
				<EditorComponent editorRef={this._setEditor} style={{ display: "flex", flex: "auto", width: "100%" }} />
			</div>
		);
	}
}
