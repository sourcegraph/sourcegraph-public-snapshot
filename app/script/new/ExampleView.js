import React from "react";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import CodeListing from "./CodeListing";

class ExampleView extends React.Component {
	constructor(props) {
		super(props);
		this.state = {};
	}

	componentWillMount() {
		this._resetExamples();
	}

	componentWillReceiveProps(nextProps) {
		this._updateState(nextProps);
		if (nextProps.defURL !== this.props.defURL) {
			this._resetExamples();
		}
	}

	shouldComponentUpdate(nextProps, nextState) {
		return nextProps.highlightedDef !== this.props.highlightedDef ||
			nextState.selectedIndex !== this.state.selectedIndex ||
			nextState.displayedIndex !== this.state.displayedIndex ||
			nextState.displayedExample !== this.state.displayedExample;
	}

	_resetExamples() {
		this.setState({
			selectedIndex: 0,
			displayedIndex: -1,
			displayedExample: null,
		}, () => {
			this._requestData();
			this._updateState(this.props);
		});
	}

	_updateState(props) {
		let count = props.examples.getCount(this.props.defURL);
		if (this.state.selectedIndex >= count) {
			this.setState({
				selectedIndex: count - 1,
			}, () => {
				this._updateState(props);
			});
			return;
		}

		let example = props.examples.get(props.defURL, this.state.selectedIndex);
		if (example !== null) {
			this.setState({
				displayedIndex: this.state.selectedIndex,
				displayedExample: example,
			});
		}
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
				this._updateState(this.props);
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
								selectedDef={this.props.defURL}
								highlightedDef={this.props.highlightedDef} />
						</div>
					}
				</div>

				<footer>
					<a target="_blank" href={`${this.props.defURL}/.examples`} className="pull-right">
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

export default ExampleView;
