import React from "react";
import Container from "sourcegraph/Container";
import RefsContainer from "sourcegraph/def/RefsContainer";
import DefStore from "sourcegraph/def/DefStore";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import "sourcegraph/blob/BlobBackend";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import "whatwg-fetch";

class ExamplesContainer extends Container {
	static propTypes = {
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
		def: React.PropTypes.string,
		defObj: React.PropTypes.object,
	};

	constructor(props) {
		super(props);
	}

	stores() {
		return [DefStore];
	}

	reconcileState(state, props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.defRepos = props.defRepos || [];
		state.sorting = props.sorting || null;
		state.examples = state.def ? DefStore.getExamples({
			repo: state.repo, commitID: state.commitID, def: state.def,
		}) : null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantExamples({
				repo: nextState.repo, commitID: nextState.commitID, def: nextState.def,
			}));
		}
	}

	render() {
		let refLocs = this.state.examples;

		const expandedSnippets = 3;
		return (
			<div>
				<div styleName="section-label">
					{refLocs && refLocs.RepoRefs && `${refLocs.RepoRefs.length} ` || ""}
					Usage Example{(refLocs && refLocs.RepoRefs && refLocs.RepoRefs.length > 1) ? "s" : ""}
				</div>
				<hr style={{marginTop: 0, clear: "both"}}/>
				{!refLocs && <i>Loading...</i>}
				{refLocs && !refLocs.RepoRefs && <i>No examples found</i>}
				{refLocs && refLocs.RepoRefs && refLocs.RepoRefs.map((repoRefs, i) => <RefsContainer
					key={i}
					repo={this.props.repo}
					rev={this.props.rev}
					commitID={this.props.commitID}
					def={this.props.def}
					defObj={this.props.defObj}
					repoRefs={repoRefs}
					prefetch={i === 0}
					initNumSnippets={expandedSnippets}
					rangeLimit={2}
					fileCollapseThreshold={5} />)}
			</div>
		);
	}
}

export default CSSModules(ExamplesContainer, styles);
