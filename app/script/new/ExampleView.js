import React from "react";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import CodeListing from "./CodeListing";

export default class ExampleView extends React.Component {
	constructor(props) {
		super(props);
		this.state = {};
		this.state = this._calculateState(props);
	}

	componentWillMount() {
		this._requestData();
	}

	componentWillReceiveProps(nextProps) {
		this.setState(this._calculateState(nextProps));
	}

	shouldComponentUpdate(nextProps, nextState) {
		return nextState.highlightedDef !== this.state.highlightedDef ||
			nextState.selectedIndex !== this.state.selectedIndex ||
			nextState.displayedIndex !== this.state.displayedIndex ||
			nextState.displayedExample !== this.state.displayedExample;
	}

	_calculateState(props) {
		let selectedIndex = this.state.selectedIndex;
		let displayedIndex = this.state.displayedIndex;
		let displayedExample = this.state.displayedExample;

		if (this.state.defURL !== props.defURL) {
			selectedIndex = 0;
			displayedIndex = -1;
			displayedExample = null;
		}

		let count = props.examples.getCount(props.defURL);
		if (selectedIndex >= count) {
			selectedIndex = count - 1;
		}

		let example = props.examples.get(props.defURL, selectedIndex);
		if (example !== null) {
			displayedIndex = selectedIndex;
			displayedExample = example;
		}

		return {
			defURL: props.defURL,
			highlightedDef: props.highlightedDef,
			selectedIndex: selectedIndex,
			displayedIndex: displayedIndex,
			displayedExample: displayedExample,
		};
	}

	_requestData(props) {
		setTimeout(() => {
			Dispatcher.dispatch(new DefActions.WantExample(this.props.defURL, this.state.selectedIndex));
		}, 0);
	}

	_changeExample(delta) {
		return () => {
			let newIndex = this.state.selectedIndex + delta;
			if (newIndex < 0 || newIndex >= this.props.examples.getCount(this.props.defURL)) {
				return;
			}
			this.setState({selectedIndex: newIndex}, () => {
				this.setState(this._calculateState(this.props));
				this._requestData();
			});
		};
	}

	render() {
		let example = this.state.displayedExample;
		let loading = this.state.selectedIndex !== this.state.displayedIndex;
		return (
			<div className="example">
				<header>
					<div className="pull-right">{example && example.Repo}</div>
					<nav>
						<a className={`fa fa-chevron-circle-left btnNav ${this.state.selectedIndex === 0 ? "disabled" : ""}`} onClick={this._changeExample(-1)}></a>
						<a className={`fa fa-chevron-circle-right btnNav ${this.state.selectedIndex === this.props.examples.getCount(this.props.defURL) - 1 ? "disabled" : ""}`} onClick={this._changeExample(+1)}></a>
					</nav>
					{example && <a>{example.File}:{example.StartLine}-{example.EndLine}</a>}
					{loading && <i className="fa fa-spinner fa-spin"></i>}
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
