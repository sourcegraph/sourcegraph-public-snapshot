import React from "react";

import CodeStore from "sourcegraph/code/CodeStore";
import Container from "sourcegraph/Container";
import DiffFileList from "sourcegraph/delta/DiffFileList";
import FileDiff from "sourcegraph/delta/FileDiff";

class FileDiffs extends Container {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.annotations = CodeStore.annotations;
	}

	stores() {
		return [CodeStore];
	}

	render() {
		return (
			<div>
				<DiffFileList files={this.props.files} stats={this.props.stats} />
				{this.props.files.map((fd, i) => (
					<FileDiff
						key={fd.OrigName + fd.NewName}
						id={`F${i}`}
						diff={fd}
						baseRepo={this.props.baseRepo}
						baseRev={this.props.baseRev}
						headRepo={this.props.headRepo}
						headRev={this.props.headRev}
						annotations={this.state.annotations} />
				))}
			</div>
		);
	}
}
FileDiffs.propTypes = {
	files: React.PropTypes.arrayOf(React.PropTypes.object),
	stats: React.PropTypes.object.isRequired,
	baseRepo: React.PropTypes.string.isRequired,
	baseRev: React.PropTypes.string.isRequired,
	headRepo: React.PropTypes.string.isRequired,
	headRev: React.PropTypes.string.isRequired,
};
export default FileDiffs;
