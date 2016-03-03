import React from "react";
import update from "react/lib/update";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import classNames from "classnames";
import emailValidator from "email-validator";

class UserList extends Component {
	constructor(props) {
		super(props);
		this.state = {
			userEmail: "", // for the user create input
			userPermission: "read", // for the user create input
		};
		this._isCurrentUser = this._isCurrentUser.bind(this);
		this._existsLocally = this._existsLocally.bind(this);
		this._isInvited = this._isInvited.bind(this);
		this._hasEmail = this._hasEmail.bind(this);
		this._avatarURL = this._avatarURL.bind(this);
		this._name = this._name.bind(this);
		this._handleUserEmailTextInput = this._handleUserEmailTextInput.bind(this);
		this._handleUserPermissionInput = this._handleUserPermissionInput.bind(this);
		this._handleAddButton = this._handleAddButton.bind(this);
		this._handleAddUser = this._handleAddUser.bind(this);
		this._handleInviteUser = this._handleInviteUser.bind(this);
		this._userSort = this._userSort.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_getUserPermissionString(user) {
		if (user.LocalAccount.Admin) return "Admin";
		if (user.LocalAccount.Write) return "Write";
		return "Read";
	}

	_isCurrentUser(user) {
		return this._existsLocally(user) && user.LocalAccount.UID === this.state.currentUser.UID;
	}

	_existsLocally(user) {
		return Boolean(user.LocalAccount);
	}

	_isInvited(user) {
		return Boolean(user.IsInvited);
	}

	_hasEmail(user) {
		return Boolean(user.Email);
	}

	_avatarURL(user) {
		if (this._existsLocally(user)) return user.LocalAccount.AvatarURL;
		return user.RemoteAccount.AvatarURL;
	}

	_name(user) {
		if (this._existsLocally(user)) return user.LocalAccount.Name || user.LocalAccount.Login;
		return user.RemoteAccount.Name || user.RemoteAccount.Login;
	}

	_handleUserEmailTextInput(e) {
		this.setState(update(this.state, {
			userEmail: {$set: e.target.value},
		}));
	}

	_handleUserPermissionInput(e) {
		this.setState(update(this.state, {
			userPermission: {$set: e.target.value},
		}));
	}

	_handleAddUser() {
		if (!emailValidator.validate(this.state.userEmail)) return;
		Dispatcher.dispatch(new DashboardActions.WantInviteUser(this.state.userEmail, this.state.userPermission));
		this.setState({showCreateUserWell: false, userEmail: "", userPermission: "read"});
	}

	_handleInviteUser(user) {
		Dispatcher.dispatch(new DashboardActions.WantInviteUsers([user.Email]));
	}

	_handleAddButton() {
		if (this.state.allowStandaloneUsers) {
			this.setState({showCreateUserWell: !this.state.showCreateUserWell});
		} else {
			const emails = this.state.users
				.filter(user => this._hasEmail(user) && !this._isInvited(user) && !this._existsLocally(user))
				.map(user => user.Email);
			if (emails.length > 0) Dispatcher.dispatch(new DashboardActions.WantInviteUsers(emails));
		}
	}

	_userSort(a, b) {
		if (this._isCurrentUser(a)) return -1;
		if (this._isCurrentUser(b)) return 1;
		if (this._existsLocally(a) && !this._existsLocally(b)) return -1;
		if (!this._existsLocally(a) && this._existsLocally(b)) return 1;
		if (this._isInvited(a) && !this._isInvited(b)) return -1;
		if (!this._isInvited(a) && this._isInvited(b)) return 1;
		if (this._hasEmail(a) && !this._hasEmail(b)) return -1;
		if (!this._hasEmail(a) && this._hasEmail(b)) return 1;
		return this._name(a) < this._name(b) ? -1 : 1;
	}

	render() {
		const emptyStateLabel = this.state.allowGitHubUsers ?
			"Link your GitHub account to add teammates." : "No teammates.";

		if (this.state.allowStandaloneUsers && !this.state.currentUser.Admin) {
			return <div />;
		}

		return (
			<div className="panel panel-default">
				<div className="panel-heading">
					<h5>Team</h5>
					{!this.state.onboarding.linkGitHub &&
						<button className="btn btn-primary add-user-btn" onClick={() => this._handleAddButton()} >
							<i className={classNames("fa button-icon", {
								"fa-users": this.state.allowGitHubUsers,
								"fa-user-plus": !this.state.allowGitHubUsers,
							})}></i>
							{!this.state.allowGitHubUsers && (!this.state.showCreateUserWell ? "ADD NEW" : "CANCEL")}
							{this.state.allowGitHubUsers && "INVITE ALL"}
						</button>
					}
					{this.state.showCreateUserWell && <div className="add-user-well">
						<div className="well">
							<div className={classNames("form-group", {
								"has-error": this.state.userEmail !== "" && !emailValidator.validate(this.state.userEmail),
							})}>
								<input className="form-control create-repo-input"
									placeholder="Email"
									type="email"
									value={this.state.userEmail}
									onKeyPress={(e) => { if ((e.keyCode || e.which) === 13) this._handleAddUser(); }}
									onChange={this._handleUserEmailTextInput} />
							</div>
							<div className="form-group user-permissions">
								<div className="radio">
									<label>
									<input type="radio"
										value="read"
										onChange={this._handleUserPermissionInput}
										checked={this.state.userPermission === "read"}/>Read
									</label>
								</div>
								<div className="radio">
									<label><input type="radio"
										value="write"
										onChange={this._handleUserPermissionInput}
										checked={this.state.userPermission === "write"} />Write
									</label>
								</div>
								<div className="radio">
									<label>
									<input type="radio"
										value="admin"
										onChange={this._handleUserPermissionInput}
										checked={this.state.userPermission === "admin"} />Admin
									</label>
								</div>
							</div>
							<button type="submit" className={classNames("btn btn-primary create-repo-btn", {
								disabled: !emailValidator.validate(this.state.userEmail),
							})}
								onClick={this._handleAddUser}>SEND INVITATION</button>
						</div>
					</div>}
				</div>
				<div className="users-list panel-body">
					{this.state.users.length === 0 ? <div className="well empty-well">{emptyStateLabel}</div> : <div className="list-group">
						{this.state.users.sort(this._userSort).map((user, i) => (
							<div className="list-group-item" key={i}>
								<img className="avatar-sm" src={this._avatarURL(user) || "https://secure.gravatar.com/avatar?d=mm&f=y&s=128"} />
								<span className="user-name">{this._name(user)}</span>
								{!this._existsLocally(user) && !this._isInvited(user) &&
									<i className={classNames("fa fa-plus-square-o add-user-icon", {"add-user-icon-disabled": !this._hasEmail(user)})}
										onClick={() => this._handleInviteUser(user)}
										data-tooltip={!this._hasEmail(user) ? "top" : null}
										title={!this._hasEmail(user) ? "No public email" : null} />
								}
								{this.state.allowStandaloneUsers &&
									<span className="user-permissions">{this._getUserPermissionString(user)}</span>
								}
							</div>
						))}
					</div>}
				</div>
			</div>
		);
	}
}

UserList.propTypes = {
	users: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	currentUser: React.PropTypes.object.isRequired,
	onboarding: React.PropTypes.object.isRequired,
	allowStandaloneUsers: React.PropTypes.bool.isRequired,
	allowGitHubUsers: React.PropTypes.bool.isRequired,
};

export default UserList;
