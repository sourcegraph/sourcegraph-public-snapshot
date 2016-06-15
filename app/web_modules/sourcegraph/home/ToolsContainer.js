import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import ToolsHomeComponent from "./ToolsHomeComponent";
import "sourcegraph/user/UserBackend"; // for side effects

class ToolsContainer extends React.Component {
	static propTypes = {
		location: React.PropTypes.object,
	};

	reconcileState(state, props, context) {
		Object.assign(state, props);
	}

	render() {
		return (<div>
			<ToolsHomeComponent location={this.props.location}/>
		</div>);
	}
}

export default CSSModules(ToolsContainer, styles);
