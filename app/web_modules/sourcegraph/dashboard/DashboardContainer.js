import React from "react";

import Container from "sourcegraph/Container";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import GitHubReposStore from "sourcegraph/dashboard/GitHubReposStore";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import OnboardingStore from "sourcegraph/dashboard/OnboardingStore";
import ModalStore from "sourcegraph/dashboard/ModalStore";

import DashboardUsers from "sourcegraph/dashboard/DashboardUsers";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import AddReposModal from "sourcegraph/dashboard/AddReposModal";
import AddUsersModal from "sourcegraph/dashboard/AddUsersModal";

import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

class DashboardContainer extends Container {
	constructor(props) {
		super(props);
	}

	componentDidMount() {
		super.componentDidMount();
	}

	componentWillUnmount() {
		super.componentWillUnmount();
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repos = (DashboardStore.repos || []).concat(GitHubReposStore.reposByOrg.getMirrored());
		state.users = (DashboardStore.users || []).concat(GitHubUsersStore.users.getAdded());
		state.showReposModal = ModalStore.showReposModal;
		state.showUsersModal = ModalStore.showUsersModal;
		state.allowStandaloneRepos = !DashboardStore.isMothership;
		state.allowGitHubMirrors = DashboardStore.allowMirrors;
		state.allowStandaloneUsers = !DashboardStore.isMothership;
		state.allowGitHubUsers = DashboardStore.allowMirrors;
	}

	stores() { return [DashboardStore, ModalStore, GitHubReposStore, GitHubUsersStore, OnboardingStore]; }

	render() {
		return (
			<div className="dashboard-container">
				{this.state.showReposModal ? <AddReposModal
					allowStandaloneRepos={this.state.allowStandaloneRepos}
					allowGitHubMirrors={this.state.allowGitHubMirrors} /> : null}
				{this.state.showUsersModal ? <AddUsersModal
					allowStandaloneUsers={this.state.allowStandaloneUsers}
					allowGitHubUsers={this.state.allowGitHubUsers} /> : null}
				<div className="dash-repos">
					<div className="dash-repos-header">
						<h3 className="your-repos">Your Repositories</h3>
						<button className="btn btn-primary add-repo-btn"
							onClick={() => Dispatcher.dispatch(new DashboardActions.OpenAddReposModal())}>
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
					<DashboardUsers users={this.state.users} allowStandaloneUsers={this.state.allowStandaloneUsers} />
				</div>
			</div>
		);
	}
}

DashboardContainer.propTypes = {
};

export default DashboardContainer;
