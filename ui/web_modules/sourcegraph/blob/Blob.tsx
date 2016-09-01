import * as React from "react";
import { createLineFromByteFunc } from "sourcegraph/blob/lineFromByte";
import { EventListener } from "sourcegraph/Component";
import * as DefActions from "sourcegraph/def/DefActions";
import * as Dispatcher from "sourcegraph/Dispatcher";

import { Def } from "sourcegraph/def";
import { urlToDefInfo } from "sourcegraph/def/routes";

import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

import "sourcegraph/blob/styles/Monaco.raw.css";

interface Props {
	contents: string;
	repo: string;
	path: string;
	rev: string;

	startByte?: number;
	endByte?: number;
};

export class Blob extends React.Component<Props, null> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	context: {
		siteConfig: { assetsRoot: string };
		router: { push: (url: string) => void };
	};

	_hoverProvided: string[];
	_toDispose: monaco.IDisposable[];
	_editor: monaco.editor.IStandaloneCodeEditor;
	_lineFromByte: (byteOffset: number) => number;

	constructor(props: Props) {
		super(props);
		this._findInPage = this._findInPage.bind(this);
		this._hoverProvided = [];
		this._toDispose = [];
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
	}

	componentDidUpdate(): void {
		if (!this._editor) { return; }
		this._updateEditor(this._editor);
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
		const editor = monaco.editor.create(this.refs["container"] as HTMLDivElement, {
			automaticLayout: true,
			value: this.props.contents,
			readOnly: true,
			scrollBeyondLastLine: false,
			wrappingColumn: 0,
			fontFamily: "Menlo, Consolas", // TODO figure out how to import font family from typography
			fontSize: 13,
		});
		this._editor = editor;
		this._updateEditor(editor);
		this._addReferencesAction(editor);
		this._toDispose.push(editor);
	}

	_updateEditor(editor: monaco.editor.IStandaloneCodeEditor): void {
		const repoSpec = {
			repo: this.props.repo,
			file: this.props.path,
			rev: this.props.rev,
		};
		const uri = RepoSpecToURI(repoSpec);

		let model = monaco.editor.getModel(uri);
		if (!model) {
			model = monaco.editor.createModel(this.props.contents, "", uri);
			this._toDispose.push(model);
		}
		editor.setModel(model);

		const lang = model.getMode().getId();
		if (this._hoverProvided.indexOf(lang) === -1) {
			const token = monaco.languages.registerHoverProvider(lang, HoverProvider);
			this._toDispose.push(token);
		}

		this._lineFromByte = createLineFromByteFunc(this.props.contents);
		this._scroll(editor);
		this._highlight(editor);
		this._addMouseDownListener(editor);
	}

	_findInPage(e: Event): void {
		const mac = navigator.userAgent.indexOf("Macintosh") >= 0;
		const ctrl = mac ? (e as KeyboardEvent).metaKey : (e as KeyboardEvent).ctrlKey;
		const FKey = 70;
		if ((e as KeyboardEvent).keyCode === FKey && ctrl) {
			if (this._editor) {
				e.preventDefault();
				(document.getElementsByClassName("inputarea")[0] as any).focus(); // HACK
				this._editor.trigger("keyboard", "actions.find", {});
			}
		}
	}

	_addReferencesAction(editor: monaco.editor.IStandaloneCodeEditor): void {
		const palette = editor.getAction("editor.action.quickCommand");
		if (palette) {
			(palette as any)._shouldShowInContextMenu = false;
			palette.dispose();
		}
		const action = {
			id: "viewAllReferences",
			label: "View all references",
			contextMenuGroupId: "1_goto",
			run: (e) => this._viewAllReferences(e),
		};
		editor.addAction(action);
	}

	_viewAllReferences(editor: monaco.editor.ICommonCodeEditor): monaco.Promise<void> {
		const pos = (editor as any).getPosition();
		const model = editor.getModel();

		return new monaco.Promise<void>(() => {
			defAtPosition(model, pos).then((def) => {
				const url = urlToDefInfo(def);
				this.context.router.push(url);
			});
		});
	}

	_addMouseDownListener(editor: monaco.editor.IStandaloneCodeEditor): void {
		editor.onMouseDown(({target, event}) => {
			if (event.rightButton) {
				return;
			}
			const ident = target.element.className.indexOf("identifier") > 0;
			if (target.position && ident) {
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

	_scroll(editor: monaco.editor.IStandaloneCodeEditor): void {
		if (this.props.startByte) {
			const startLine = this._lineFromByte(this.props.startByte);
			editor.revealLineInCenter(startLine);
		}
	}

	_highlight(editor: monaco.editor.IStandaloneCodeEditor): void {
		if (this.props.startByte && this.props.endByte) {
			const model = editor.getModel();

			const startLine = this._lineFromByte(this.props.startByte);
			const startCol = model.getLineMinColumn(startLine);

			const endLine = this._lineFromByte(this.props.endByte);
			const endCol = model.getLineMaxColumn(endLine);

			const range = new monaco.Range(startLine, startCol, endLine, endCol);
			editor.setSelection(range);
		}
	}

	render(): JSX.Element {
		return <div style={{ display: "flex", flex: "auto", width: "100%" }}>
			<div ref="container" style={{ width: "100%" }} />
			<EventListener target={global.document} event="keydown" callback={this._findInPage} />
		</div>;
	}
}

// We have to make a request to the server to find the def at a position because
// the client does not have srclib annotation data.
// This involves a ton of string munging because we can't save the data types in a good way.
// A monaco position is slightly different than a Sourcegraph one.
const fetch = singleflightFetch(defaultFetch);
class HoverProvider {
	static provideHover(model: monaco.editor.IReadOnlyModel, position: monaco.Position): monaco.Thenable<monaco.languages.Hover> {
		return defAtPosition(model, position).then((def: Def) => {
			const word = model.getWordAtPosition(position);
			const title = `${def.Name} ${def.FmtStrings.Type.Unqualified}`;
			const docs = def.Docs ? def.Docs[0].Data : "";
			return {
				range: new monaco.Range(position.lineNumber, word.startColumn, position.lineNumber, word.endColumn),
				contents: [title, docs],
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

// The hover-info end point returns a full def.
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
