// tslint:disable typedef ordered-imports

import * as isEqual from "lodash/isEqual";
import * as React from "react";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from  "sourcegraph/def/DefActions";
import {Props} from "sourcegraph/blob/Blob";

type State = any;

/*
withJumpDefRedirect wraps a Blob component. It handles both sending jumpToDef requests
to the HTTP API and redirecting when a valid response is received.
*/
export function withJumpToDefRedirect(Blob) {
	class WithJumpToDefRedirect extends Container<Props, State> {
		static contextTypes: React.ValidationMap<any> = {
			router: React.PropTypes.object,
		};

		_dispatcherToken: string;

		componentDidMount(): void {
			this._dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));
		}

		componentWillUnmount(): void {
			Dispatcher.Stores.unregister(this._dispatcherToken);
		}

		reconcileState(state: State, props: Props): void {
			Object.assign(state, props);
			if (props.location && props.location.query) {
					let {line, character, file, commit, repo} = props.location.query;
					if ([line, character, file, repo].every(Boolean)) { // Commit is allowed to be null / undefined.
						if (this.state.path === file && this.state.commitID === commit && this.state.repo === repo) {
							state.soughtJumpDef = {
								line: line,
								character: character,
								file: file,
								commit: commit,
								repo: repo,
							};
						}
					}
			} else {
				state.soughtJumpDef = null;
			}
		}

		onStateTransition(prevState: State, nextState: State): void {
			if (nextState.soughtJumpDef && !isEqual(prevState.soughtJumpDef, nextState.soughtJumpDef)) {
				Dispatcher.Backends.dispatch(new DefActions.WantJumpDef(nextState.soughtJumpDef));
			}
		}

		__onDispatch(action) {
			if (action instanceof DefActions.JumpDefFetched) {
				if (this.state.soughtJumpDef) {
					let position = {
						repo: this.state.repo,
						commit: this.state.commitID,
						file: this.state.path,
						line: this.state.soughtJumpDef.line,
						character: this.state.soughtJumpDef.character,
					};
					if (isEqual(action.pos, position)) {
						if (action.def.Error) {
							// This is only a temporary solution until we have proper error handling.
							console.error(action.def.Error);
						} else {
							(this.context as any).router.replace(action.def.path);
						}
					}
				}
			}
		}

		render(): JSX.Element {
			return <Blob {...this.props} {...this.state} />;
		}
	}
	return WithJumpToDefRedirect;
}
