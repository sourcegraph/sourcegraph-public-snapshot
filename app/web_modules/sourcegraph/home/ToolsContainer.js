import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import ToolsHomeComponent from "./ToolsHomeComponent";
import Component from "sourcegraph/Component";

class ToolsContainer extends Component {
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
