import * as React from "react";
import {Container} from "sourcegraph/Container";
import {Location} from "sourcegraph/Location";
import {repoParam, repoPath, repoRev} from "sourcegraph/repo";
import "sourcegraph/repo/RepoBackend";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {Store} from "sourcegraph/Store";

interface Props {
	params: any;
	location: Location;
}

type State = any;

// withResolvedRepoRev reads the repo, rev, etc.,
export function withResolvedRepoRev(Component: any): React.ComponentClass<Props> {

	class WithResolvedRepoRev extends Container<Props, State> {
		stores(): Store<any>[] {
			return [RepoStore];
		}

		componentWillUnmount(): void {
			if (super.componentWillUnmount) {
				super.componentWillUnmount();
			}
		}

		reconcileState(state: State, props: Props): void {
			Object.assign(state, props);

			const repoSplat = repoParam(props.params.splat);
			state.repo = repoPath(repoSplat);
			state.rev = repoRev(repoSplat); // the original rev from the URL
		}

		render(): JSX.Element | null {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithResolvedRepoRev;
}
