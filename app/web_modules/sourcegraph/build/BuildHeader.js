import * as React from "react";

import Component from "sourcegraph/Component";
import TimeAgo from "sourcegraph/util/TimeAgo";
import {buildStatus, buildClass, elapsed} from "sourcegraph/build/Build";
import CSSModules from "react-css-modules";
import styles from "./styles/Build.css";

class BuildHeader extends Component {
	reconcileState(state, props) {
		if (state.build !== props.build) {
			state.build = props.build;
		}
	}

	render() {
		return (
			<header styleName={`header ${buildClass(this.state.build)}`}>
				<div styleName="number">#{this.state.build.ID}</div>
				<div styleName="status">{buildStatus(this.state.build)}</div>
				<div styleName="date">
					<TimeAgo time={this.state.build.EndedAt || this.state.build.StartedAt || this.state.build.CreatedAt} />
				</div>
				<div styleName="elapsed">{elapsed(this.state.build)}</div>
			</header>
		);
	}
}

BuildHeader.propTypes = {
	build: React.PropTypes.object.isRequired,
};

export default CSSModules(BuildHeader, styles, {allowMultiple: true});
