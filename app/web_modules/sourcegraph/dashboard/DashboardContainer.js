import React from "react";
import update from "react/lib/update";

import Container from "sourcegraph/Container";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import OnboardingStore from "sourcegraph/dashboard/OnboardingStore";

import DashboardUsers from "sourcegraph/dashboard/DashboardUsers";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import AddReposModal from "sourcegraph/dashboard/AddReposModal";
import AddUsersModal from "sourcegraph/dashboard/AddUsersModal";

class DashboardContainer extends Container {
	constructor(props) {
		super(props);
		this.state = {
			showReposModal: false,
			showUsersModal: false,
		};
		this._openReposModal = this._openReposModal.bind(this);
		this._dismissReposModal = this._dismissReposModal.bind(this);
		this._openUsersModal = this._openUsersModal.bind(this);
		this._dismissUsersModal = this._dismissUsersModal.bind(this);
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
		state.allowStandaloneRepos = !DashboardStore.isMothership;
		state.allowGitHubMirrors = DashboardStore.allowMirrors;
	}

	stores() { return [DashboardStore, OnboardingStore]; }

	_openReposModal() {
		this.setState(update(this.state, {
			showReposModal: {$set: true},
		}));
	}

	_dismissReposModal() {
		this.setState(update(this.state, {
			showReposModal: {$set: false},
		}));
	}

	_openUsersModal() {
		this.setState(update(this.state, {
			showUsersModal: {$set: true},
		}));
	}

	_dismissUsersModal() {
		this.setState(update(this.state, {
			showUsersModal: {$set: false},
		}));
	}

	render() {
		return (
			<div className="dashboard-container">
				{this.state.showReposModal ? <AddReposModal
					dismissModal={this._dismissReposModal}
					allowStandaloneRepos={this.state.allowStandaloneRepos}
					allowGitHubMirrors={this.state.allowGitHubMirrors} /> : null}
				{this.state.showUsersModal ? <AddUsersModal
					dismissModal={this._dismissUsersModal} /> : null}
				<div className="dash-repos">
					<div className="dash-repos-header">
						<h3 className="your-repos">Your Repositories</h3>
						<button className="btn btn-primary btn-block add-repo-btn"
							onClick={this._openReposModal}>
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
					<DashboardUsers users={this.state.users} openUsersModal={this._openUsersModal} />
				</div>
			</div>
		);
	}
}

DashboardContainer.propTypes = {
};

export default DashboardContainer;
