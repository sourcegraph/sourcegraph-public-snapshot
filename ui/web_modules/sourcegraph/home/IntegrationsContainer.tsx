import {Integrations} from "./Integrations";
import * as React from "react";
import "sourcegraph/user/UserBackend"; // for side effects

interface Props {
	location?: any;
}

type State = any;

export class IntegrationsContainer extends React.Component<Props, State> {
	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
	}

	render(): JSX.Element | null {
		return (<div>
			<Integrations location={this.props.location}/>
		</div>);
	}
}
