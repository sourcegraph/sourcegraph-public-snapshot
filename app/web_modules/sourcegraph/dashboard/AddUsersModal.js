import React from "react";
import update from "react/lib/update";

import Container from "sourcegraph/Container";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import SelectableListWidget from "sourcegraph/dashboard/SelectableListWidget";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";

class AddUsersModal extends Container {
	constructor(props) {
		super(props);
		this.state = {
			email: "",
		};
		this._handleTextInput = this._handleTextInput.bind(this);
		this._handleCreate = this._handleCreate.bind(this);
		this._handleAddMirrors = this._handleAddMirrors.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.selectedUsers = GitHubUsersStore.selectedUsers;
		state.currentOrg = GitHubUsersStore.currentOrg;
		state.orgs = GitHubUsersStore.orgs;
		state.selectAll = GitHubUsersStore.selectAll;
		state.items = GitHubUsersStore.usersByOrg.get(state.currentOrg).map((user) => ({
			name: user.RemoteAccount.Name ? `${user.RemoteAccount.Login} (${user.RemoteAccount.Name})` : user.RemoteAccount.Login,
			key: user.RemoteAccount.Login,
		}));
	}

	_handleTextInput(e) {
		this.setState(update(this.state, {
			email: {$set: e.target.value},
		}));
	}

	_handleCreate() {
		// TODO:
		// Dispatcher.dispatch(new DashboardActions.WantAddUsers());
		this.state.dismissModal();
	}

	_handleAddMirrors(users) {
		// TODO:
		// Dispatcher.dispatch(new DashboardActions.WantAddUsers());
		this.state.dismissModal();
	}

	stores() { return [GitHubUsersStore]; }

	render() {
		return (
			<div className="modal add-users-widget"
				style={{display: "block"}}
				tabIndex="-1"
				role="dialog" >
				<div className="modal-dialog">
					<div className="modal-content">
						<div className="modal-header">
							<button type="button"
								className="close"
								data-dismiss="modal"
								aria-label="Close"
								onClick={this.state.dismissModal}>
								<span aria-hidden="true">&times;</span>
							</button>
							<h4 className="modal-title">Invite People to join Sourcegraph</h4>
						</div>
						<div className="modal-body">
							<ul className="nav nav-tabs" role="tablist">
								<li role="presentation" className="active">
									<a href="#email-invite" role="tab" data-toggle="tab">Invite by Email</a>
								</li>
								<li role="presentation">
									<a href="#github-invite" role="tab" data-toggle="tab">Invite from GitHub</a>
								</li>
							</ul>

							<div className="tab-content">
								<div role="tabpanel" className="tab-pane active" id="email-invite">
									<div className="widget-body">
										<p className="add-repo-label">REPO NAME:</p>
										<input className="form-control"
											type="text"
											value={this.state.email}
											placeholder="Type Name here"
											onChange={this._handleTextInput}/>
									</div>
									<div className="widget-footer">
										<button className="btn btn-block btn-primary btn-lg"
											onClick={this._handleCreate}>
											CREATE
										</button>
									</div>
								</div>
								<div role="tabpanel" className="tab-pane" id="github-invite">
									<SelectableListWidget items={this.state.mirrorUsers}
										currentCategory={this.state.currentOrg}
										menuCategories={this.state.orgs}
										onMenuClick={(org) => Dispatcher.dispatch(new DashboardActions.SelectUserOrg(org))}
										onSelect={(userKey, select) => Dispatcher.dispatch(new DashboardActions.SelectUser(userKey, select))}
										onSelectAll={(users, selectAll) => Dispatcher.dispatch(new DashboardActions.SelectUsers(users, selectAll))}
										selections={this.state.selectedUsers}
										selectAll={this.state.selectAll}
										menuLabel="Organizations"
										searchPlaceholderText="Search GitHub contacts"
										onSubmit={this._handleAddMirrors} />
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	}
}

AddUsersModal.propTypes = {
	dismissModal: React.PropTypes.func.isRequired,
	allowStandaloneUsers: React.PropTypes.bool.isRequired,
	allowGitHubMirrors: React.PropTypes.bool.isRequired,
};

export default AddUsersModal;
