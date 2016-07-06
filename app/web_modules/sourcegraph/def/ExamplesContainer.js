import React from "react";
import Container from "sourcegraph/Container";
import RefsContainer from "sourcegraph/def/RefsContainer";
import DefStore from "sourcegraph/def/DefStore";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import "sourcegraph/blob/BlobBackend";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import base from "sourcegraph/components/styles/_base.css";
import typography from "sourcegraph/components/styles/_typography.css";
import {Panel, Heading, Loader} from "sourcegraph/components";
import "whatwg-fetch";

class ExamplesContainer extends Container {
	static propTypes = {
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
		def: React.PropTypes.string,
		defObj: React.PropTypes.object,
		className: React.PropTypes.string,
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
				<Heading level="7" className={base.mb3} styleName="cool-mid-gray">
					Usage Example{(refLocs && refLocs.RepoRefs && refLocs.RepoRefs.length > 1) ? "s" : ""}
				</Heading>
				<Panel
					hoverLevel="low"
					styleName="full-width-sm b--cool-pale-gray"
					className={base.ba}>
					<div className={this.props.className}>
						{!refLocs && <div className={typography.tc}> <Loader className={base.mv4} /></div>}
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
				</Panel>
			</div>
		);
	}
}

export default CSSModules(ExamplesContainer, styles, {allowMultiple: true});
