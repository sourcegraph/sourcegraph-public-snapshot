import * as React from "react";
import { rel } from "sourcegraph/app/routePatterns";
import {Component} from "sourcegraph/Component";

interface Props {
	repo: string;
	rev: string | null;
	params: any;
	path: string;
	route?: any;
}

type State = any;

export function withPath(InnerComponent: any): React.ComponentClass<Props> {
	class WithPath extends Component<Props, State> {

		reconcileState(state: State, props: Props): void {
			Object.assign(state, props);

			state.path = props.route && props.route.path && props.route.path.startsWith(rel.blob) ? props.params.splat[1] : state.path;
			if (!state.path) {
				state.path = null;
			}
		}

		render(): JSX.Element | null {
			return <InnerComponent {...this.props} {...this.state} />;
		}
	}

	return WithPath;
}
