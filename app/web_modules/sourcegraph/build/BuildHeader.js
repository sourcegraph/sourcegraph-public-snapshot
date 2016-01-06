import React from "react";

import Component from "sourcegraph/Component";
import TimeAgo from "sourcegraph/util/TimeAgo";
import {buildStatus, buildClass, elapsed} from "sourcegraph/build/Build";

class BuildHeader extends Component {
	reconcileState(state, props) {
		if (state.build !== props.build) {
			state.build = props.build;
		}
	}

	render() {
		return (
			<header className="repo-build">
				<h1 className={`label label-${buildClass(this.state.build)}`}>
					<div className="number">#{this.state.build.ID}</div>
					<div className="status">{buildStatus(this.state.build)}</div>
					<div className="date">
						<TimeAgo time={this.state.build.EndedAt || this.state.build.StartedAt || this.state.build.CreatedAt} />
					</div>
					{elapsed(this.state.build)}
				</h1>
			</header>
		);
	}
}

BuildHeader.propTypes = {
	build: React.PropTypes.object.isRequired,
};

export default BuildHeader;
