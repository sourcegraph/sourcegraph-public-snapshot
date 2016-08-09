// tslint:disable

import * as React from "react";
import * as styles from "./styles/Integrations.css";
import {Integrations} from "./Integrations";
import "sourcegraph/user/UserBackend"; // for side effects

type Props = {
	location?: any,
};

export class IntegrationsContainer extends React.Component<Props, any> {
	reconcileState(state, props, context) {
		Object.assign(state, props);
	}

	render(): JSX.Element | null {
		return (<div>
			<Integrations location={this.props.location}/>
		</div>);
	}
}
