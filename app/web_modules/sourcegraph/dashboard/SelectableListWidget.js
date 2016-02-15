import React from "react";

import Component from "sourcegraph/Component";

import ListMenu from "sourcegraph/dashboard/ListMenu";
import SelectableList from "sourcegraph/dashboard/SelectableList";

import Dispatcher from "sourcegraph/Dispatcher";

class SelectableListWidget extends Component {
	constructor(props) {
		super(props);
		this._getSelections = this._getSelections.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_getSelections() {
		return this.state.items.filter(item => this.state.selections[item.index]);
	}

	render() {
		const selected = this._getSelections();
		return (
			<div className="selectable-list-widget list-picker">
				<div className="list-picker">
					<div className="category-menu">
						<ListMenu label={this.state.menuLabel}
							categories={this.state.menuCategories}
							current={this.state.currentCategory} />
					</div>
					<div className="list">
						<SelectableList items={this.state.items.filter(item => item.org === this.state.currentCategory)}
							selectAll={this.state.selectAll}
							selections={this.state.selections}
							searchPlaceholderText="TODO" />
					</div>
				</div>
				<div className="footer submit-form">
					<button className="btn btn-block btn-primary btn-lg"
						onClick={() => this.state.onSubmit(selected)}>
						{`add${selected.length > 0 ? ` (${selected.length})` : ""}`}
					</button>
					{/*<p>
						<a onClick={(e) => {
							e.preventDefault();
							Dispatcher.dispatch(new OnboardingActions.AdvanceProgressStep());
						}}>i'll do that later</a>
					</p>*/}
				</div>
			</div>
		);
	}
}

SelectableListWidget.propTypes = {
	items: React.PropTypes.arrayOf(React.PropTypes.shape({
		org: React.PropTypes.string,
		name: React.PropTypes.string,
		index: React.PropTypes.number,
	})).isRequired,
	currentCategory: React.PropTypes.string,
	menuCategories: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
	selections: React.PropTypes.object.isRequired,
	selectAll: React.PropTypes.bool.isRequired,
	menuLabel: React.PropTypes.string.isRequired,
	onSubmit: React.PropTypes.func.isRequired,
};

export default SelectableListWidget;
