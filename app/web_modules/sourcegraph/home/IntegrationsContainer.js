import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Integrations.css";
import Integrations from "./Integrations";
import "sourcegraph/user/UserBackend"; // for side effects

class IntegrationsContainer extends React.Component {
	static propTypes = {
		location: React.PropTypes.object,
	};

	reconcileState(state, props, context) {
		Object.assign(state, props);
	}

	render() {
		return (<div>
			<Integrations location={this.props.location}/>
		</div>);
	}
}

export default CSSModules(IntegrationsContainer, styles);
