import React from "react";

import Component from "sourcegraph/Component";
import s from "sourcegraph/blob/styles/Blob.css";

class BlobToolbar extends Component {
	reconcileState(state, props) {
		state.repo = props.repo;
		state.rev = props.rev;
		state.path = props.path || null;
	}

	render() {
		return (
			<div className={s.toolbar}>
				<div className="actions">
				</div>
			</div>
		);
	}
}

BlobToolbar.propTypes = {
	repo: React.PropTypes.string.isRequired,
	rev: React.PropTypes.string.isRequired,
	path: React.PropTypes.string,
};

export default BlobToolbar;
