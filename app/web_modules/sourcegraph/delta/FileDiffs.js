import React from "react";
import styles from "sourcegraph/delta/styles/FileDiffs.css";
import CSSModules from "react-css-modules";
import BlobStore from "sourcegraph/blob/BlobStore";
import Container from "sourcegraph/Container";
import * as DefActions from "sourcegraph/def/DefActions";
import DefStore from "sourcegraph/def/DefStore";
import DefTooltip from "sourcegraph/def/DefTooltip";
import DiffFileList from "sourcegraph/delta/DiffFileList";
import Dispatcher from "sourcegraph/Dispatcher";
import FileDiff from "sourcegraph/delta/FileDiff";
import {isExternalLink} from "sourcegraph/util/externalLink";
import {routeParams as defRouteParams} from "sourcegraph/def";

class FileDiffs extends Container {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.annotations = BlobStore.annotations;

		state.defs = DefStore.defs;
		state.highlightedDef = DefStore.highlightedDef || null;

		if (state.highlightedDef && !isExternalLink(state.highlightedDef)) {
			let {repo, rev, def} = defRouteParams(state.highlightedDef);
			state.highlightedDefObj = DefStore.defs.get(repo, rev, def);
		} else {
			state.highlightedDefObj = null;
		}
	}

	onStateTransition(prevState, nextState) {
		if (nextState.highlightedDef && prevState.highlightedDef !== nextState.highlightedDef) {
			if (!isExternalLink(nextState.highlightedDef)) { // kludge to filter out external def links
				let {repo, rev, def, err} = defRouteParams(nextState.highlightedDef);
				if (err) {
					console.err(err);
				} else {
					Dispatcher.Backends.dispatch(new DefActions.WantDef(repo, rev, def));
				}
			}
		}
	}

	stores() {
		return [BlobStore, DefStore];
	}

	render() {
		return (
			<div styleName="container">
				<DiffFileList files={this.props.files} stats={this.props.stats} />
				{this.props.files.map((fd, i) => (
					<div styleName="file-diff" key={fd.OrigName + fd.NewName}><FileDiff
						id={`F${i}`}
						diff={fd}
						baseRepo={this.props.baseRepo}
						baseRev={this.props.baseRev}
						headRepo={this.props.headRepo}
						headRev={this.props.headRev}
						highlightedDef={this.state.highlightedDef}
						highlightedDefObj={this.state.highlightedDefObj}
						annotations={this.state.annotations} /></div>
				))}
				{this.state.highlightedDefObj && !this.state.highlightedDefObj.Error && <DefTooltip currentRepo={this.state.baseRepo} def={this.state.highlightedDefObj} />}
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
export default CSSModules(FileDiffs, styles);
