import * as React from "react";
import { createLineFromByteFunc } from "sourcegraph/blob/lineFromByte";
import * as DefActions from "sourcegraph/def/DefActions";
import * as Dispatcher from "sourcegraph/Dispatcher";

import { Def } from "sourcegraph/api";
import { urlToDefInfo } from "sourcegraph/def/routes";

import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

import "sourcegraph/blob/styles/Monaco.raw.css";
import { code_font_face } from "sourcegraph/components/styles/_vars.css";

import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	contents: string;
	repo: string;
	path: string;
	rev: string;

	startByte?: number;
};

export class Blob extends React.Component<Props, null> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	context: {
		siteConfig: { assetsRoot: string };
		router: { push: (url: string) => void };
		eventLogger: { logEventForCategory: (eventCategory: string, eventAction: string, eventLabel: string, eventProperties?: any) => void };
	};

	_hoverProvided: string[];
	_toDispose: monaco.IDisposable[];
	_editor: monaco.editor.IStandaloneCodeEditor;
	_decorationID: string[];
	_mouseDownPosition: monaco.editor.IMouseTarget | null;

	// Finding the line from byte requires UTF-8 encoding the entire buffer,
	// because Sourcegraph uses byte offset and Monaco uses (UTF16) character
	// offset. We cache it here so we don't have to calculate it too many times.
	_lineFromByte: (byteOffset: number) => number;

	constructor(props: Props) {
		super(props);
		this._findInPage = this._findInPage.bind(this);
		this._hoverProvided = [];
		this._toDispose = [];
		this._decorationID = [];
	}

	componentDidMount(): void {
		if ((global as any).require) {
			this._loaderReady();
			return;
		}

		let script = document.createElement("script");
		script.type = "text/javascript";
		script.src = `${this.context.siteConfig.assetsRoot}/vs/loader.js`;
		script.addEventListener("load", this._loaderReady.bind(this));
		document.body.appendChild(script);
	}

	componentWillUnmount(): void {
		this._toDispose.forEach(disposable => {
			disposable.dispose();
		});
		global.document.removeEventListener("keydown", this._findInPage);
	}

	componentDidUpdate(): void {
		if (!this._editor) { return; }
		this._updateEditor();
	}

	_loaderReady(): void {
		if ((global as any).monaco) {
			this._monacoReady();
			return;
		}

		(global as any).require.config({ paths: { "vs": `${this.context.siteConfig.assetsRoot}/vs` } });
		(global as any).require(["vs/editor/editor.main"], this._monacoReady.bind(this));
	}

	_monacoReady(): void {
		this._editor = monaco.editor.create(this.refs["container"] as HTMLDivElement, {
			automaticLayout: true,
			value: this.props.contents,
			readOnly: true,
			scrollBeyondLastLine: false,
			wrappingColumn: 0,
			fontFamily: code_font_face,
			fontSize: 13,
			cursorStyle: "block",
		});
		this._toDispose.push(this._editor);

		global.document.addEventListener("keydown", this._findInPage);
		this._addClickListener();
		this._addReferencesAction();
		this._overrideNavigationKeys();

		this._updateEditor();
	}

	_updateEditor(): void {
		const repoSpec = {
			repo: this.props.repo,
			file: this.props.path,
			rev: this.props.rev,
		};
		const uri = RepoSpecToURI(repoSpec);
		this._updateModel(uri);

		this._scroll();
	}

	_updateModel(uri: monaco.Uri): void {
		let model = monaco.editor.getModel(uri);
		if (model) {
			// If the model doesn't change, we don't need to update the editor.
			return;
		}
		model = monaco.editor.createModel(this.props.contents, "", uri);
		this._toDispose.push(model);
		this._editor.setModel(model);

		const lang = model.getMode().getId();
		if (this._hoverProvided.indexOf(lang) === -1) {
			const token = monaco.languages.registerHoverProvider(lang, HoverProvider);
			this._toDispose.push(token);
			this._hoverProvided.push(lang);
		}
		this._lineFromByte = createLineFromByteFunc(this.props.contents);
	}

	_findInPage(e: KeyboardEvent): void {
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

	_addReferencesAction(): void {
		const palette = this._editor.getAction("editor.action.quickCommand");
		if (palette) {
			(palette as any)._shouldShowInContextMenu = false;
			palette.dispose();
		}
		const action = {
			id: "viewAllReferences",
			label: "View all references",
			contextMenuGroupId: "1_goto",
			run: (e) => this._viewAllReferences(e),
			enablement: {
				tokensAtPosition: ["identifier"],
			},
		};
		this._editor.addAction(action);
	}

	_viewAllReferences(editor: monaco.editor.ICommonCodeEditor): monaco.Promise<void> {
		const pos = editor.getPosition();
		const model = editor.getModel();

		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REFERENCES, AnalyticsConstants.ACTION_CLICK, "ClickedViewReferences", {repo: this.props.repo, path: this.props.path, rev: this.props.rev});

		return new monaco.Promise<void>(() => {
			defAtPosition(model, pos).then((def) => {
				const url = urlToDefInfo(def);
				this.context.router.push(url);
			});
		});
	}

	_addClickListener(): void {
		this._editor.onMouseDown(({target, event}) => {
			const mac = navigator.userAgent.indexOf("Macintosh") >= 0;
			if (event.rightButton || (event.ctrlKey && mac)) {
				this._mouseDownPosition = null;
				return;
			}
			this._mouseDownPosition = target;
		});

		this._editor.onMouseUp(({target}) => {
			if (!this._mouseDownPosition || !target.position.equals(this._mouseDownPosition.position)) {
				return;
			}
			const saved = this._mouseDownPosition;
			const ident = saved.element.className.indexOf("identifier") > 0;
			if (saved.position && ident) {
				const pos = {
					repo: this.props.repo,
					commit: this.props.rev,
					file: this.props.path,
					line: target.position.lineNumber - 1,
					character: target.position.column - 1,
				};
				Dispatcher.Backends.dispatch(new DefActions.WantJumpDef(pos));
			}
		});
	}

	_scroll(): void {
		let startLine;
		const matches = /#(\d+)-(\d+)/.exec(global.window.location.hash);
		if (matches) {
			const start = parseInt(matches[1], 10);
			const end = parseInt(matches[2], 10);
			this._highlightLines(start, end);
			startLine = Math.min(start, end);
		} else if (this.props.startByte) {
			startLine = this._lineFromByte(this.props.startByte) - 1;
		} else {
			return;
		}
		const linesInViewPort = this._editor.getDomNode().offsetHeight / 20;
		const middleLine = Math.floor(startLine + (linesInViewPort / 4));
		this._editor.revealLineInCenter(middleLine);
	}

	_overrideNavigationKeys(): void {
		// Monaco overrides the back and forward history commands, so we
		// implement our own here. AFAICT, there isn't a good way
		// to unbind a default keybinding.
		 /* tslint:disable */
		 // Disable tslint because it doesn't like bitwise operators.
		this._editor.addCommand(monaco.KeyCode.LeftArrow | monaco.KeyMod.CtrlCmd, () => {
			global.window.history.back();
		}, "");
		this._editor.addCommand(monaco.KeyCode.RightArrow | monaco.KeyMod.CtrlCmd, () => {
			global.window.history.forward();
		}, "");
		 /* tslint:enable */
		this._editor.addCommand(monaco.KeyCode.Home, () => {
			this._editor.revealLine(1);
		}, "");
		this._editor.addCommand(monaco.KeyCode.End, () => {
			this._editor.revealLine(
				this._editor.getModel().getLineCount()
			);
		}, "");
	}

	_highlightLines(startLine: number, endLine: number): void {
		const range = new monaco.Range(
			startLine,
			this._editor.getModel().getLineMinColumn(startLine),
			endLine,
			this._editor.getModel().getLineMaxColumn(endLine),
		);
		this._editor.setSelection(range);
	}

	render(): JSX.Element {
		return <div ref="container" style={{ display: "flex", flex: "auto", width: "100%" }} />;
	}
}

// We have to make a request to the server to find the def at a position because
// the client does not have srclib annotation data. This involves a ton of
// string munging because we can't save the data types in a good way. A monaco
// position is slightly different than a Sourcegraph one.
const fetch = singleflightFetch(defaultFetch);
class HoverProvider {
	static provideHover(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<monaco.languages.Hover> {
		return defAtPosition(model, position).then((def: Def) => {
			if (!def) {
				throw new Error("def not found");
			}
			const word = model.getWordAtPosition(position);
			const title = `**${def.Name}** ${def.FmtStrings.Type.Unqualified}`;
			const serverDoc = def.Docs ? def.Docs[0].Data : "";
			const viewRefsSuggestion = "*Right click to view all references.*";
			let docs = serverDoc.replace(/\s+/g, " ");
			if (docs.length > 400) {
				docs = docs.substring(0, 380);
				docs = docs + "...";
			}
			return {
				range: new monaco.Range(position.lineNumber, word.startColumn, position.lineNumber, word.endColumn),
				contents: [title, docs, viewRefsSuggestion],
			};
		});
	}
}

function defAtPosition(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<Def> {
	const url = hoverURL(model.uri, position);
	return fetch(url)
		.then(checkStatus)
		.then(response => response.json())
		.then(data => data.def)
		.catch(error => { console.error(error); });
}

// The hover-info end point returns a full def. We need to ask the server for
// information about the symbol at the point, because we don't have enough info
// on the client.
function hoverURL(uri: monaco.Uri, position: monaco.Position): string {
	const line = position.lineNumber - 1;
	const col = position.column - 1;
	const {repo: repo, file: file, rev: rev} = pathToURI(uri);
	return `/.api/repos/${repo}@${rev}/-/hover-info?file=${file}&line=${line}&character=${col}`;
}

interface RepoSpec {
	repo: string;
	file: string;
	rev: string;
}

function pathToURI(uri: monaco.Uri): RepoSpec {
	const matches = /(.*)\/-\/(.*)\/-\/(.*)/.exec(uri.fsPath);
	if (!matches || matches.length < 4) { throw new Error("invalid argument, model URI probably set incorrectly"); }
	const repo = matches[1];
	const rev = matches[2];
	const file = matches[3];
	return { repo: repo, rev: rev, file: file };
}

function RepoSpecToURI({repo, file, rev}: RepoSpec): monaco.Uri {
	return monaco.Uri.file(`${repo}/-/${rev}/-/${file}`);
}
