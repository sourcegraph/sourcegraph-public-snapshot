import React from "react";
import update from "react/lib/update";

import Container from "sourcegraph/Container";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import OnboardingStore from "sourcegraph/dashboard/OnboardingStore";

import DashboardUsers from "sourcegraph/dashboard/DashboardUsers";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import AddReposModal from "sourcegraph/dashboard/AddReposModal";

class DashboardContainer extends Container {
	constructor(props) {
		super(props);
		this.state = {
			showModal: false,
		};
		this._openModal = this._openModal.bind(this);
		this._dismissModal = this._dismissModal.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
	}

	componentWillUnmount() {
		super.componentWillUnmount();
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repos = DashboardStore.repos;
		state.users = DashboardStore.users;
		state.isOnboarding = true;
	}

	stores() { return [DashboardStore, OnboardingStore]; }

	_openModal() {
		this.setState(update(this.state, {
			showModal: {$set: true},
		}));
	}

	_dismissModal() {
		this.setState(update(this.state, {
			showModal: {$set: false},
		}));
	}

	render() {
		if (this.state.isOnboarding) return null;
		return (
			<div className="dashboard-container dashboard">
				{this.state.showModal ? <AddReposModal dismissModal={this._dismissModal} /> : null}
				<div className="dash-repos">
					<div className="dash-repos-header">
						<h3 className="your-repos">Your Repositories</h3>
						<button className="btn btn-primary btn-block add-repo-btn"
							onClick={this._openModal}>
							<div className="plus-btn">
								<span className="plus">+</span>
							</div>
							<span className="add-repo-label">Add New</span>
						</button>
					</div>
					<div>
						<DashboardRepos repos={this.state.repos} />
					</div>
				</div>
				<div className="dash-users">
					<DashboardUsers users={this.state.users} />
				</div>
			</div>
		);
	}
}

DashboardContainer.propTypes = {
};

export default DashboardContainer;
