import {Location} from "history";
import * as React from "react";
import {withJumpToDefRedirect} from "sourcegraph/blob/withJumpToDefRedirect";
import {Component} from "sourcegraph/Component";
import {Def} from "sourcegraph/def/index";

export interface Props {
	startlineCallback?: any;
	location: Location;
	contents: string;
	textSize?: string;
	annotations?: {
		Annotations: any[];
		LineStartBytes: any[];
	};
	lineNumbers?: boolean;
	skipAnns?: boolean;
	startLine?: number;
	startCol?: number;
	startByte?: number;
	endLine?: number;
	endCol?: number;
	endByte?: number;
	scrollToStartLine?: boolean;
	highlightedDef: string | null;
	highlightedDefObj: Def | null;

	// activeDef is the def ID ("UnitType/Unit/-/Path") only of the currently
	// active def (i.e., the subject of the current DefInfo page). It should
	// not be the whole def URL path (it should not include the repo and rev).
	activeDef?: string;
	activeDefRepo?: string;

	// For linking line numbers to the file they came from (e.g., in
	// ref snippets).
	repo: string;
	rev?: string;
	commitID?: string;
	path: string;

	// contentsOffsetLine indicates that the contents string does not
	// start at line 1 within the file, but rather some other line number.
	// It must be specified when startLine > 1 but the contents don't begin at
	// the first line of the file.
	contentsOffsetLine?: number;

	highlightSelectedLines?: boolean;

	// dispatchSelections is whether this Blob should emit BlobActions.SelectCharRange
	// actions when the text selection changes. It should be true for the main file view but
	// not for secondary file views (e.g., usage examples).
	dispatchSelections?: boolean;

	// display line expanders is whether or not to show only the top line expander,
	// the bottom line expander, or both
	displayLineExpanders?: string;

	displayRanges?: any;
}

interface State {
	contents: string;
};

// BlobTestOnly should only be used on its own for testing purposes. Normally, 
// you should be using Blob that's at the bottom of this file. 
export class BlobTestOnly extends Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	context: {
		siteConfig: {assetsRoot: string};
	};

	state: State = {
		contents: "",
	};

	_editor?: monaco.editor.IStandaloneCodeEditor;

	componentDidMount(): void {
		if ((global as any).require) {
			this.loaderReady();
			return;
		}

		let script = document.createElement("script");
		script.type = "text/javascript";
		script.src = `${this.context.siteConfig.assetsRoot}/vs/loader.js`;
		script.addEventListener("load", this.loaderReady.bind(this));
		document.body.appendChild(script);
	}

	componentWillUnmount(): void {
		if (this._editor) {
			this._editor.dispose();
		}
	}

	loaderReady(): void {
		if ((global as any).monaco) {
			this.monacoReady();
			return;
		}

		(global as any).require.config({paths: {"vs": `${this.context.siteConfig.assetsRoot}/vs`}});
		(global as any).require(["vs/editor/editor.main"], this.monacoReady.bind(this));
	}

	monacoReady(): void {
		this._editor = monaco.editor.create(this.refs["container"] as HTMLDivElement, {
			value: this.state.contents,
			language: "go",
			readOnly: true,
			scrollBeyondLastLine: false,
			contextmenu: false,
			wrappingColumn: 0,
		});
	}

	reconcileState(state: State, props: Props): void {
		state.contents = props.contents;
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (!this._editor) {
			return;
		}
		if (nextState.contents !== prevState.contents) {
			this._editor.setValue(nextState.contents);
		}
	}

	render(): JSX.Element | null {
		return <div ref="container" style={{width: "100%", height: "500px"}} />;
	}
}

let blob = withJumpToDefRedirect(BlobTestOnly);
export {blob as Blob};
