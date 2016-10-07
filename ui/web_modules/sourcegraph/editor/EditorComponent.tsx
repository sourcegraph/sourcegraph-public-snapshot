// tslint:disable typedef ordered-imports
import * as React from "react";
import * as invariant from "invariant";
import {Editor} from "sourcegraph/editor/Editor";
import "sourcegraph/blob/styles/Monaco.css";

type Props = {
	editorRef?: (editor: Editor | null) => void;
	style?: Object;
};

interface State {
	monacoLoaded?: boolean;
};

// EditorComponent wraps the Monaco loader and code editor. Embedders
// must use the editorRef property to get the Monaco editor object and
// manipulate it using the Monaco API.
export class EditorComponent extends React.Component<Props, State> {

	private _container: HTMLElement;
	private _editor: Editor | null = null;

	componentWillUnmount(): void {
		if (this.props.editorRef) {
			this.props.editorRef(null);
		}
		if (this._editor) {
			this._editor.dispose();
		}
	}

	render(): JSX.Element {
		const otherProps = Object.assign({}, this.props);
		delete otherProps.editorRef;
		return <div ref={(e) => this.setContainer(e)} {...otherProps} />;
	}

	private setContainer(e: HTMLElement): void {
		this._container = e;

		if (e) {
			invariant(!this._editor, "editor is already initialized");
			this._editor = new Editor(this._container);
		} else if (this._editor) {
			this._editor.dispose();
			this._editor = null;
		}

		if (this.props.editorRef) {
			this.props.editorRef(this._editor || null);
		}
	}
}
