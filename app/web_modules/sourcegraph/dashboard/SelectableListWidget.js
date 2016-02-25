import React from "react";

import Component from "sourcegraph/Component";

import ListMenu from "sourcegraph/dashboard/ListMenu";
import SelectableList from "sourcegraph/dashboard/SelectableList";

class SelectableListWidget extends Component {
	constructor(props) {
		super(props);
		this._getSelections = this._getSelections.bind(this);
		this._getCategoryNumbers = this._getCategoryNumbers.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_getSelections() {
		return this.state.items.filter(item => this.state.selections[item.key]);
	}

	_getCategoryNumbers(selected) {
		let categoryNumbers = {};
		for (let cat of this.state.menuCategories) {
			categoryNumbers[categoryNumbers[cat]] = 0;
		}
		selected.map(repo => { categoryNumbers[repo.Owner.Login] += 1; });
		return categoryNumbers;
	}
	render() {
		const selected = this._getSelections();
		return (
			<div className="selectable-list-widget">
				<div className="widget-body">
					<div className="category-menu">
						<ListMenu label={this.state.menuLabel}
							categories={this.state.menuCategories}
							onMenuClick={this.state.onMenuClick}
							current={this.state.currentCategory}
							categoryNumbers={this._getCategoryNumbers(selected)} />
					</div>
					<div className="list-control">
						<SelectableList items={this.state.items}
							unselectableItems={this.state.unselectableItems}
							selectAll={this.state.selectAll}
							selections={this.state.selections}
							onSelect={this.state.onSelect}
							onSelectAll={this.state.onSelectAll}
							searchPlaceholderText={this.state.searchPlaceholderText || "search"} />
					</div>
				</div>
				<div className="widget-footer">
					<button className="btn btn-block btn-primary btn-lg"
						onClick={() => this.state.onSubmit(selected)} disabled={this.state.showLoading}>
						<div>
						{this.state.showLoading &&
							<i className="fa fa-spinner fa-spin"></i>
						}
						{!this.state.showLoading &&
							(selected.length > 0 ? `add (${selected.length}) ${selected.length === 1 ? "repository" : "repositories"}` : "add repositories")
						}
						</div>
					</button>
				</div>
			</div>
		);
	}
}

SelectableListWidget.propTypes = {
	items: React.PropTypes.arrayOf(React.PropTypes.shape({
		name: React.PropTypes.string,
		key: React.PropTypes.string,
	})).isRequired,
	unselectableItems: React.PropTypes.arrayOf(React.PropTypes.shape({
		name: React.PropTypes.string,
		key: React.PropTypes.string,
		reason: React.PropTypes.string,
	})).isRequired,
	allItems: React.PropTypes.arrayOf(React.PropTypes.object),
	currentCategory: React.PropTypes.string,
	menuCategories: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
	onMenuClick: React.PropTypes.func.isRequired,
	onSelect: React.PropTypes.func.isRequired,
	onSelectAll: React.PropTypes.func.isRequired,
	selections: React.PropTypes.object.isRequired,
	selectAll: React.PropTypes.bool.isRequired,
	menuLabel: React.PropTypes.string.isRequired,
	onSubmit: React.PropTypes.func.isRequired,
	searchPlaceholderText: React.PropTypes.string,
	showLoading: React.PropTypes.bool.isRequired,
};

export default SelectableListWidget;
