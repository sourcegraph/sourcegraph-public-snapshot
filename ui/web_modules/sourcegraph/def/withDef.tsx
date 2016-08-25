// tslint:disable: typedef ordered-imports

import * as React from "react";
import {DefStore} from "sourcegraph/def/DefStore";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import "sourcegraph/def/DefBackend";
import * as DefActions from "sourcegraph/def/DefActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Container} from "sourcegraph/Container";
import {routeParams as defRouteParams} from "sourcegraph/def/index";
import {Header} from "sourcegraph/components/Header";
import {httpStatusCode} from "sourcegraph/util/httpStatusCode";
import Helmet from "react-helmet";
import {Store} from "sourcegraph/Store";

// withDef fetches the def specified in the params. It also fetches
// the def stored in DefStore.highlightedDef.
export function withDef(Component) {
	interface Props {
		repo: string;
		rev?: string;
		params: any;
		isCloning?: boolean;
		def?: string;
	}

	type State = any;

	class WithDef extends Container<Props, State> {
		stores(): Store<any>[] {
			return [DefStore, RepoStore];
		}

		reconcileState(state: State, props: Props): void {
			Object.assign(state, props);

			if (!props.def) {
				state.def = props.params ? props.params.splat[1] : null;
			}
			state.defObj = state.def ? DefStore.defs.get(state.repo, state.rev, state.def) : null;
			state.commitID = state.defObj && !state.defObj.Error ? state.defObj.CommitID : null;
			state.highlightedDefObj = null;
			state.isCloning = RepoStore.repos.isCloning(state.repo);
		}

		onStateTransition(prevState: State, nextState: State): void {
			if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
				Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.repo, nextState.rev, nextState.def));
			}

			if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
				let {repo, rev, def} = defRouteParams(nextState.highlightedDef);
				Dispatcher.Backends.dispatch(new DefActions.WantDef(repo, rev, def));
			}
		}

		render(): JSX.Element | null {
			if (this.state.isCloning) {
				return (
					<Header title="Cloning this repository" loading={true} />
				);
			}

			if (this.state.defObj && this.state.defObj.Error) {
				return (
					<div>
						<Helmet title={"Not Found"} />
						<Header
							title={`${httpStatusCode(this.state.defObj.Error)}`}
							subtitle={`Definition is not available.`} />
					</div>
				);
			}

			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithDef;
}
