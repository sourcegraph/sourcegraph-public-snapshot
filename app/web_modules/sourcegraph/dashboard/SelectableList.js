import React from "react";

import Component from "sourcegraph/Component";

import Dispatcher from "sourcegraph/Dispatcher";
import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";

import update from "react/lib/update";

class SelectableList extends Component {
	constructor(props) {
		super(props);
		this.state = {
			searchQuery: "",
		};
		this._handleSearch = this._handleSearch.bind(this);
		this._handleSelectAll = this._handleSelectAll.bind(this);
		this._handleSelect = this._handleSelect.bind(this);
		this._showItem = this._showItem.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_handleSearch(e) {
		this.setState(update(this.state, {
			searchQuery: {$set: e.target.value},
			selectAll: {$set: false},
		}));
	}

	_handleSelectAll(e) {
		const items = this.state.items.filter(this._showItem);
		Dispatcher.dispatch(new OnboardingActions.SelectItems(items, this.state.type, e.target.checked));
	}

	_handleSelect(e, index) {
		Dispatcher.dispatch(new OnboardingActions.SelectItem(index, this.state.type, e.target.checked));
	}

	_showItem(item) {
		const query = this.state.searchQuery;
		return query === "" ? true : item.name.toLowerCase().indexOf(query.toLowerCase()) !== -1;
	}

	render() {
		return (
			<div className="add-repo-list">
				<div className="header">
					<div className="search-bar">
						<i className="fa fa-search search-icon"></i>
						<input className="form-control search-input"
							placeholder={this.state.searchPlaceholderText || "Search"}
							value={this.state.searchQuery}
							onChange={this._handleSearch}
							type="text" />
					</div>
					<div className="table-controls">
						<div className="select-all">
							<input className="select-all"
								type="checkbox"
								defaultChecked={false}
								checked={this.state.selectAll}
								onChange={this._handleSelectAll} />
						</div>
						<span className="name">NAME</span>
					</div>
				</div>
				<div className="body">
					<div className="list-group">
						{this.state.items.filter(this._showItem).map(item =>
							<div className="table-row" key={item.index}>
								<div className="select">
									<input
										type="checkbox"
										name={`${this.state.formItemName}[]`}
										checked={this.state.selections[item.index]}
										onChange={e => this._handleSelect(e, item.index)}
										value={item.name} />
								</div>
								<span className="name">{item.name}</span>
							</div>
						)}
					</div>
				</div>
			</div>
		);
	}
}

SelectableList.propTypes = {
	items: React.PropTypes.arrayOf(React.PropTypes.shape({
		index: React.PropTypes.number,
		name: React.PropTypes.string,
	})).isRequired,
	// type identifies the entity type of the items which populate the list
	type: React.PropTypes.string.isRequired,
	// selections is a object which identifies which items are currently selected {index: isSelected}
	selections: React.PropTypes.object.isRequired,
	// selectAll identifies if the "select all" aggregator is toggled
	selectAll: React.PropTypes.bool.isRequired,
	searchPlaceholderText: React.PropTypes.string,
};

export default SelectableList;
