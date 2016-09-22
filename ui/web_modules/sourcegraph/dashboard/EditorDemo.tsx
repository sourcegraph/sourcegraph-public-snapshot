// tslint:disable: typedef ordered-imports
import * as React from "react";
import { Editor } from "sourcegraph/editor/Editor";
import {BlobStore, keyForFile} from "sourcegraph/blob/BlobStore";
import {Container} from "sourcegraph/Container";
import {Store} from "sourcegraph/Store";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import "sourcegraph/blob/BlobBackend";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { EditorComponent } from "sourcegraph/editor/EditorComponent";
import {URI} from "sourcegraph/core/uri";

type Props = {
	repo: string;
	rev: string;
	path: string;
	startLine: number;
};
type State = Props & {
	contents: string | null;
};

// The purpose of this file is to easily render one example during the
// onboarding flow that will allow the user to hover over examples in
// a sandboxed mode. This is similar to the regular RefsContainer
// except that css styles and eventlisteners / additional fetches for
// information are removed.
export class EditorDemo extends Container<Props, State> {
	constructor(props: Props) {
		super(props);
		this._setEditor = this._setEditor.bind(this);
	}

	stores(): Store<any>[] {
		return [BlobStore];
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		const treeEntry = BlobStore.files[keyForFile(state.repo, state.rev, state.path)] || null;
		state.contents = treeEntry && treeEntry.ContentsString ? treeEntry.ContentsString : null;
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.path !== nextState.path) {
			Dispatcher.Backends.dispatch(new BlobActions.WantFile(nextState.repo, nextState.rev, nextState.path));
		}
	}

	private _setEditor(editor: Editor) {
		if (editor) {
			const range = new monaco.Range(this.props.startLine, 1, this.props.startLine, 1);
			const uri = URI.pathInRepo(this.props.repo, this.props.rev, this.props.path);
			editor.setInput(uri, range);
		}
	}

	render(): JSX.Element | null {
		return (
			<EditorComponent editorRef={this._setEditor} style={{display: "flex", flexDirection: "column", height:"225px", border: "solid 1px #efefef"}} />
		);
	}
}
