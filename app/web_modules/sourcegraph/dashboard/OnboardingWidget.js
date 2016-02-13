import React from "react";

import Component from "sourcegraph/Component";

import ListMenu from "sourcegraph/dashboard/ListMenu";
import SelectableList from "sourcegraph/dashboard/SelectableList";
import EntityTypes from "sourcegraph/dashboard/EntityTypes";

import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";

class OnboardingWidget extends Component {
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
			<div>
				<div className="list-picker">
					<div className="category-menu">
						<ListMenu label={this.state.menuLabel}
							categories={this.state.orgs}
							current={this.state.currentOrg} />
					</div>
					<div className="list">
						<SelectableList items={this.state.items.filter(item => item.org === this.state.currentOrg)}
							type={this.state.currentType}
							selectAll={this.state.selectAll}
							selections={this.state.selections}
							searchPlaceholderText={`Search GitHub ${this.state.currentType === EntityTypes.REPO ? "repositories" : "contacts"}`} />
					</div>
				</div>
				<div className="footer submit-form">
					<button className="btn btn-block btn-primary btn-lg"
						onClick={(e) => {
							if (this.state.currentType === EntityTypes.REPO) {
								Dispatcher.dispatch(new DashboardActions.WantAddRepos(this._getSelections()))
							} else if (this.state.currentType === EntityTypes.USER) {
								Dispatcher.dispatch(new DashboardActions.WantAddUsers(this._getSelections()))
							}
						}}>
						{`add${selected.length > 0 ? ` (${selected.length})` : ""}`}
					</button>
					<p>
						<a onClick={(e) => {
							e.preventDefault();
							Dispatcher.dispatch(new OnboardingActions.AdvanceProgressStep());
						}}>i'll do that later</a>
					</p>
				</div>
			</div>
		);
	}
}

OnboardingWidget.propTypes = {
	items: React.PropTypes.arrayOf(React.PropTypes.shape({
		org: React.PropTypes.string,
		name: React.PropTypes.string,
		index: React.PropTypes.number,
	})).isRequired,
	currentType: React.PropTypes.string.isRequired,
	currentOrg: React.PropTypes.string,
	orgs: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
	selections: React.PropTypes.object.isRequired,
	selectAll: React.PropTypes.bool.isRequired,
	menuLabel: React.PropTypes.string.isRequired,
};

export default OnboardingWidget;
