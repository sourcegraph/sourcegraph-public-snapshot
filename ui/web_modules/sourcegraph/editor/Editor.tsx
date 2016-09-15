// tslint:disable typedef ordered-imports
import * as React from "react";
import { InjectedRouter } from "react-router";
import * as invariant from "invariant";

import { Def } from "sourcegraph/api";
import { urlToDefInfo } from "sourcegraph/def/routes";
import { makeRepoRev } from "sourcegraph/repo";
import { urlToBlob, parseBlobURL } from "sourcegraph/blob/routes";

import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

import "sourcegraph/blob/styles/Monaco.css";
import { code_font_face } from "sourcegraph/components/styles/_vars.css";

import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

import {EventLogger} from "sourcegraph/util/EventLogger";

const fetch = singleflightFetch(defaultFetch);

interface Props {
	repo: string;
	path: string;
	rev: string | null;
	contents: string | null;

	startLine?: number;
	endLine?: number;

	editorRef?: (editor: monaco.editor.IStandaloneCodeEditor | null) => void;
};

interface State {
	userManuallyScrolledToLineViaSelection?: number | null;
}

// Editor wraps the Monaco code editor.
export class Editor extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	state: State = {};

	context: {
		siteConfig: { assetsRoot: string };
		router: InjectedRouter
	};

	_hoverProvided: string[];
	_toDispose: monaco.IDisposable[];
	_editor: monaco.editor.IStandaloneCodeEditor;
	_decorationID: string[];

	public _mouseDownOnIdent: boolean = false;
	private _mouseDownPosition: monaco.editor.IMouseTarget | null = null;
	private _mouseDownIsRightButton: boolean = false;

	constructor(props: Props) {
		super(props);
		this._hoverProvided = [];
		this._toDispose = [];
		this._decorationID = [];
	}

	componentDidMount(): void {
		if ((global as any).require) {
			this._loaderReady();
		} else {
			let script = document.createElement("script");
			script.type = "text/javascript";
			script.src = `${this.context.siteConfig.assetsRoot}/vs/loader.js`;
			script.addEventListener("load", () => this._loaderReady());
			document.body.appendChild(script);
		}
	}

	componentWillUnmount(): void {
		this._toDispose.forEach(disposable => {
			disposable.dispose();
		});
		if (this.props.editorRef) {
			this.props.editorRef(null);
		}
	}

	componentWillReceiveProps(nextProps: Props): void {
		if (this._editor) {
			this._editorPropsChanged(this.props, nextProps);
		}
	}

	_loaderReady(): void {
		if ((global as any).monaco) {
			this._monacoReady();
		} else {
			(global as any).require.config({ paths: { "vs": `${this.context.siteConfig.assetsRoot}/vs` } });
			(global as any).require(["vs/editor/editor.main"], () => this._monacoReady());
		}
	}

	_monacoReady(): void {
		invariant(!this._editor, "editor is already initialized");

		this._editor = monaco.editor.create(this.refs["container"] as HTMLDivElement, {
			readOnly: true,
			automaticLayout: true,
			scrollBeyondLastLine: false,
			wrappingColumn: 0,
			fontFamily: code_font_face,
			fontSize: 13,
		});
		this._toDispose.push(this._editor);

		this._editorPropsChanged(null, this.props);

		if (this.props.editorRef) {
			this.props.editorRef(this._editor);
		}

		this._addClickHandler();

		// Remove the "Command Palette" item from the context menu.
		const palette = this._editor.getAction("editor.action.quickCommand");
		if (palette) {
			(palette as any)._shouldShowInContextMenu = false;
			palette.dispose();
		}

		// Add the "View All References" item to the context menu.
		this._editor.addAction({
			id: "viewAllReferences",
			label: "View All References",
			contextMenuGroupId: "1_goto",
			run: (e) => this._viewAllReferences(e),
		});

		// Monaco overrides the back and forward history commands, so
		// we implement our own here. There currently isn't a way to
		// unbind a default keybinding.
		/* tslint:disable no-bitwise */
		this._editor.addCommand(monaco.KeyCode.LeftArrow | monaco.KeyMod.CtrlCmd, () => {
			global.window.history.back();
		}, "");
		this._editor.addCommand(monaco.KeyCode.RightArrow | monaco.KeyMod.CtrlCmd, () => {
			global.window.history.forward();
		}, "");
		/* tslint:enable no-bitwise */
		this._editor.addCommand(monaco.KeyCode.Home, () => {
			this._editor.revealLine(1);
		}, "");
		this._editor.addCommand(monaco.KeyCode.End, () => {
			this._editor.revealLine(
				this._editor.getModel().getLineCount()
			);
		}, "");
	}

	_editorPropsChanged(prev: Props | null, next: Props): void {
		if (!prev || prev.repo !== next.repo || prev.rev !== next.rev || prev.path !== next.path || prev.contents !== next.contents) {
			this._updateModel(next.repo, next.rev, next.path, next.contents);
		}
		if (!prev || next.startLine !== this.props.startLine || next.endLine !== this.props.endLine) {
			if (typeof next.startLine === "number") {
				if (next.startLine !== this.state.userManuallyScrolledToLineViaSelection) {
					this._editor.revealLineInCenterIfOutsideViewport(next.startLine);
					this._highlightLines(next.startLine, next.endLine);
					this._editor.focus();
				}
			}
		}
	}

	_updateModel(repo: string, rev: string | null, path: string, contents: string | null): void {
		const uri = uriForTreeEntry(repo, rev, path);
		let model = monaco.editor.getModel(uri);
		if (!model) {
			model = monaco.editor.createModel(contents || "Loading...", "", uri);
			this._toDispose.push(model);
		}
		if (contents && model.getValue() !== contents) {
			model.setValue(contents);
		}
		if (this._editor.getModel().id !== model.id) {
			this._editor.setModel(model);
		}

		const lang = model.getMode().getId();
		if (this._hoverProvided.indexOf(lang) === -1) {
			const token = monaco.languages.registerHoverProvider(lang, this);
			this._toDispose.push(token);
			this._hoverProvided.push(lang);
		}
	}

	_highlightLines(startLine: number, endLine?: number): void {
		endLine = typeof endLine === "number" ? endLine : startLine;
		const range = new monaco.Range(
			startLine,
			this._editor.getModel().getLineMinColumn(startLine),
			endLine,
			this._editor.getModel().getLineMaxColumn(endLine),
		);
		this._editor.setSelection(range);
	}

	_viewAllReferences(editor: monaco.editor.ICommonCodeEditor): monaco.Promise<void> {
		// HACK: Handle the case where we've selected the line
		// (because we just jumped to a def on the line) and we
		// right-click and choose "View All References". The cursor
		// will be at the end of the line, but we want to act on the
		// token we right-clicked on.
		let pos: monaco.Position;
		if (this._mouseDownPosition && this._mouseDownIsRightButton && !editor.getSelection().isEmpty()) {
			pos = this._mouseDownPosition.position;
		} else {
			pos = editor.getPosition();
		}
		const model = editor.getModel();

		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REFERENCES, AnalyticsConstants.ACTION_CLICK, "ClickedViewReferences", { repo: this.props.repo, path: this.props.path, rev: this.props.rev });

		return new monaco.Promise<void>(() => {
			defAtPosition(model, pos).then((resp) => {
				if (resp) {
					window.location.href = urlToDefInfo(resp.def);
				}
			});
		});
	}

	_addClickHandler(): void {
		this._editor.onMouseDown(({target, event}) => {
			if (target.type === monaco.editor.MouseTargetType.UNKNOWN) {
				return;
			}

			this._mouseDownPosition = target;
			this._mouseDownIsRightButton = event.rightButton;

			// Record if this is a click starting on a clickable thing,
			// so we know in the onSelectionChange handler to ignore it.
			this._mouseDownOnIdent = target.element.className.indexOf("identifier") !== -1;
		});

		this._editor.onMouseUp(({event, target}) => {
			if (!this._mouseDownPosition || !target.position || !target.position.equals(this._mouseDownPosition.position) || this._mouseDownIsRightButton) {
				return;
			}
			const saved = this._mouseDownPosition;
			const ident = saved.element.className.indexOf("identifier") > 0;
			if (saved.position && ident) {
				EventLogger.logEventForCategory(
					AnalyticsConstants.CATEGORY_DEF,
					AnalyticsConstants.ACTION_CLICK,
					"BlobTokenClicked",
					{
						srcRepo: this.props.repo,
						srcRev: this.props.rev || "",
						srcPath: this.props.path,
						language: this._editor.getModel().getModeId(),
					}
				);

				fetchJumpToDef(this._editor.getModel(), target.position).then((resp: JumpToDefResponse) => {
					if (!resp) {
						return;
					}

					if (event.ctrlKey || event.altKey || event.metaKey || event.shiftKey || !event.leftButton) {
						global.window.open(resp.Path, "_blank");
					} else {
						// TODO(monaco): If you have selected a line and then click on something that causes
						// you to jump to that line, it deselects the line because the Blob props do not change
						// (because the URL #L123 is unchanged). It should reselect and rescroll to the line.
						this.context.router.push(resp.Path);
					}
				});
			}
		});
	}

	provideHover(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<monaco.languages.Hover> {
		return defAtPosition(model, position).then((resp: HoverInfoResponse) => {
			let contents: monaco.MarkedString[] = [];
			if (resp) {
				const def = resp.def;
				if (resp.Title) {
					contents.push(resp.Title);
				} else {
					contents.push(`**${def.Name}** ${def.FmtStrings ? def.FmtStrings.Type.Unqualified.trim() : ""}`);
					const serverDoc = def.Docs ? def.Docs[0].Data : "";
					let docs = serverDoc.replace(/\s+/g, " ");
					if (docs.length > 400) {
						docs = docs.substring(0, 380);
						docs = docs + "...";
					}
					if (docs) {
						contents.push(docs);
					}
				}
				contents.push("*Right-click to view all references.*");

				EventLogger.logEventForCategory(
					AnalyticsConstants.CATEGORY_DEF,
					AnalyticsConstants.ACTION_HOVER,
					"Hovering",
					{
						repo: this.props.repo,
						rev: this.props.rev || "",
						path: this.props.path,
						language: model.getModeId(),
					}
				);
			}

			const word = model.getWordAtPosition(position) || new monaco.Range(position.lineNumber, position.column, position.lineNumber, position.column);
			return {
				range: new monaco.Range(position.lineNumber, word.startColumn, position.lineNumber, word.endColumn),
				contents: contents,
			};
		});
	}

	render(): JSX.Element {
		return <div ref="container" style={{ display: "flex", flex: "auto", width: "100%" }} />;
	}
}

interface HoverInfoResponse {
	def: Def;
	Title?: string;
}

function defAtPosition(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<HoverInfoResponse> {
	const line = position.lineNumber - 1;
	const col = position.column - 1;
	const {repo, rev, path} = treeEntryFromUri(model.uri);
	return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/hover-info?file=${path}&line=${line}&character=${col}`)
		.then(checkStatus)
		.then(resp => resp.json())
		.catch(err => null);
}

interface JumpToDefResponse {
	Path: string;
}

function fetchJumpToDef(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<JumpToDefResponse> {
	const line = position.lineNumber - 1;
	const col = position.column - 1;
	const {repo, rev, path} = treeEntryFromUri(model.uri);
	return fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/jump-def?file=${path}&line=${line}&character=${col}`)
		.then(checkStatus)
		.then(resp => resp.json())
		.catch(err => null);
}

function uriForTreeEntry(repo: string, rev: string | null, path: string): monaco.Uri {
	return monaco.Uri.from({
		scheme: "sourcegraph",
		path: urlToBlob(repo, rev, path),
	});
}

function treeEntryFromUri(uri: monaco.Uri): {repo: string, rev: string | null, path: string} {
	invariant(uri.scheme === "sourcegraph", `unexpected uri scheme: ${uri}`);
	return parseBlobURL(uri.path);
}
