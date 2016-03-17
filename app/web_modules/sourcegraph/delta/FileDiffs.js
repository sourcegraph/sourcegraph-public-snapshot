import React from "react";

import BlobStore from "sourcegraph/blob/BlobStore";
import Container from "sourcegraph/Container";
import * as DefActions from "sourcegraph/def/DefActions";
import DefStore from "sourcegraph/def/DefStore";
import DefTooltip from "sourcegraph/def/DefTooltip";
import DiffFileList from "sourcegraph/delta/DiffFileList";
import Dispatcher from "sourcegraph/Dispatcher";
import FileDiff from "sourcegraph/delta/FileDiff";

// TODO(sqs): FileDiffs does not yet support multiple-defs (when a single
// ref links to multiple defs, like Go embedded fields linking to both the
// type and the field). We could copy over the implementation from
// BlobContainer, but that is going to be factored out soon, and let's
// keep it clean.

class FileDiffs extends Container {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.annotations = BlobStore.annotations;

		state.defs = DefStore.defs;
		state.highlightedDef = DefStore.highlightedDef || null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			Dispatcher.Backends.dispatch(new DefActions.WantDef(nextState.highlightedDef));
		}
	}

	stores() {
		return [BlobStore, DefStore];
	}

	render() {
		const highlightedDefData = this.state.highlightedDef && this.state.defs.get(this.state.highlightedDef);

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
						annotations={this.state.annotations}
						defs={this.state.defs} />
				))}
				{highlightedDefData && !this.state.defOptionsURLs && <DefTooltip def={highlightedDefData} />}
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
