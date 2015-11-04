import React from "react";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import CodeListing from "./CodeListing";

export default class ExampleView extends Component {
	reconcileState(state, props) {
		// reset examples if showing a different definition
		if (state.defURL !== props.defURL) {
			state.defURL = props.defURL;
			state.selectedIndex = 0;
			state.displayedIndex = -1;
			state.displayedExample = null;
		}

		// fix selected index if not enough examples
		state.count = props.examples.getCount(props.defURL);
		if (state.selectedIndex >= state.count) {
			state.selectedIndex = Math.max(state.count - 1, 0);
		}

		// update displayed example if example was fetched
		let example = props.examples.get(props.defURL, state.selectedIndex);
		if (example !== null) {
			state.displayedIndex = state.selectedIndex;
			state.displayedExample = example;
		}

		state.highlightedDef = props.highlightedDef;
	}

	requestData(prevState, nextState) {
		if (prevState.defURL !== nextState.defURL || prevState.selectedIndex !== nextState.selectedIndex) {
			Dispatcher.dispatch(new DefActions.WantExample(nextState.defURL, nextState.selectedIndex));
		}
	}

	_changeExample(delta) {
		return () => {
			let newIndex = this.state.selectedIndex + delta;
			if (newIndex < 0 || newIndex >= this.state.count) {
				return;
			}
			this.setState({selectedIndex: newIndex});
		};
	}

	render() {
		let example = this.state.displayedExample;
		let loading = this.state.selectedIndex !== this.state.displayedIndex && this.state.count !== 0;
		return (
			<div className="example">
				<header>
					<div className="pull-right">{example && example.Repo}</div>
					<nav>
						<a className={`fa fa-chevron-circle-left btnNav ${this.state.selectedIndex === 0 ? "disabled" : ""}`} onClick={this._changeExample(-1)}></a>
						<a className={`fa fa-chevron-circle-right btnNav ${this.state.selectedIndex >= this.state.count - 1 ? "disabled" : ""}`} onClick={this._changeExample(+1)}></a>
					</nav>
					{example && <a>{example.File}:{example.StartLine}-{example.EndLine}</a>}
					{loading && <i className="fa fa-spinner fa-spin"></i>}
					{this.state.count === 0 && "No example available"}
				</header>

				<div className="body">
					{example &&
						<div style={{opacity: loading ? 0.5 : 1}}>
							<CodeListing
								lines={example.SourceCode.Lines}
								selectedDef={this.state.defURL}
								highlightedDef={this.state.highlightedDef} />
						</div>
					}
				</div>

				<footer>
					<a target="_blank" href={`${this.state.defURL}/.examples`} className="pull-right">
						<i className="fa fa-eye" /> View all
					</a>
				</footer>
			</div>
		);
	}
}

ExampleView.propTypes = {
	defURL: React.PropTypes.string,
	examples: React.PropTypes.object,
	highlightedDef: React.PropTypes.string,
};
