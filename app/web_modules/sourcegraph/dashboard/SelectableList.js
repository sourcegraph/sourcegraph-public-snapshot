import React from "react";

import Component from "sourcegraph/Component";

import update from "react/lib/update";

class SelectableList extends Component {
	constructor(props) {
		super(props);
		this.state = {
			searchQuery: "",
		};
		this._handleSearch = this._handleSearch.bind(this);
		this._handleSelectAll = this._handleSelectAll.bind(this);
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
		this.state.onSelectAll(items, e.target.checked);
	}

	_showItem(item) {
		const query = this.state.searchQuery;
		return query === "" ? true : item.name.toLowerCase().indexOf(query.toLowerCase()) !== -1;
	}

	render() {
		return (
			<div className="selectable-list">
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
							<div className="table-row" key={item.key}>
								<div className="select">
									<input
										type="checkbox"
										name={`${this.state.formItemName}[]`}
										checked={this.state.selections[item.key]}
										onChange={e => this.state.onSelect(item.key, e.target.checked)}
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
		key: React.PropTypes.number,
		name: React.PropTypes.string,
	})).isRequired,
	// type identifies the entity type of the items which populate the list
	// selections is a object which identifies which items are currently selected {key: isSelected}
	selections: React.PropTypes.object.isRequired,
	// selectAll identifies if the "select all" aggregator is toggled
	selectAll: React.PropTypes.bool.isRequired,
	searchPlaceholderText: React.PropTypes.string,
	onSelect: React.PropTypes.func.isRequired,
	onSelectAll: React.PropTypes.func.isRequired,
};

export default SelectableList;
