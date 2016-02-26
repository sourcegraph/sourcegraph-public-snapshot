import React from "react";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import classNames from "classnames";

class UserList extends Component {
	constructor(props) {
		super(props);
		this._handleInviteAllUsers = this._handleInviteAllUsers.bind(this);
		this._handleInviteUser = this._handleInviteUser.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_getUserPermissionString(user) {
		if (user.Admin) return "Admin";
		if (user.Write) return "Write";
		return "Read";
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
		if (this._existsLocally(user)) {
			return user.LocalAccount.AvatarURL || "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";
		}
		return user.RemoteAccount.AvatarURL || "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";
	}

	_name(user) {
		if (this._existsLocally(user)) {
			return user.LocalAccount.Name || user.LocalAccount.Login;
		}
		return user.RemoteAccount.Name || user.RemoteAccount.Login;
	}

	_handleInviteUser(user) {
		let email = [];
		email.push(user.Email);
		Dispatcher.dispatch(new DashboardActions.WantInviteUsers(email));
	}

	_handleInviteAllUsers() {
		const emails = this.state.users
			.filter(user => this._hasEmail(user) && !this._isInvited(user) && !this._existsLocally(user))
			.map(user => user.Email);
		if (emails.length > 0) Dispatcher.dispatch(new DashboardActions.WantInviteUsers(emails));
	}


	render() {
		const emptyStateLabel = this.state.allowGitHubUsers ?
			"Link your GitHub account to add teammates." : "No teammates.";

		const userSort = (a, b) => {
			if (this._existsLocally(a) && a.LocalAccount.UID === window.currentUser.UID) return -1;
			if (this._existsLocally(b) && b.LocalAccount.UID === window.currentUser.UID) return 1;
			if (this._existsLocally(a) && !this._existsLocally(b)) return -1;
			if (!this._existsLocally(a) && this._existsLocally(b)) return 1;
			if (this._isInvited(a) && !this._isInvited(b)) return -1;
			if (!this._isInvited(a) && this._isInvited(b)) return 1;
			if (this._hasEmail(a) && !this._hasEmail(b)) return -1;
			if (!this._hasEmail(a) && this._hasEmail(b)) return 1;
			return this._name(a) < this._name(b) ? -1 : 1;
		};

		return (
			<div className="panel panel-default">
				<div className="panel-heading">
					<h5>Team</h5>
					{this.state.allowGitHubUsers && !this.state.onboarding.linkGitHub &&
						<button className="btn btn-default add-user-btn" data-tooltip="top" title="Invite all teammates"
							onClick={() => this._handleInviteAllUsers()} >
							<i className="fa fa-users"></i>
						</button>
					}
					{!this.state.allowGitHubUsers &&
						<button className="btn btn-primary add-user-btn"
							onClick={() => Dispatcher.dispatch(new DashboardActions.OpenAddUsersModal())} >
							<i className="fa fa-user-plus"></i>
						</button>
					}
				</div>
				<div className="users-list panel-body">
					{this.state.users.length === 0 ? <div className="well empty-well">{emptyStateLabel}</div> : <div className="list-group">
						{this.state.users.sort(userSort).map((user, i) => (
							<div className="list-group-item" key={i}>
								<img className="avatar-sm" src={this._avatarURL(user)} />
								<span className="user-name">
									{this._name(user)}{this._isInvited(user) ? " (pending)" : ""}
								</span>
								{!this._existsLocally(user) && !this._isInvited(user) &&
									<i className={classNames("fa fa-plus-square-o add-user-icon", {"add-user-icon-disabled": !this._hasEmail(user)})}
										onClick={() => this._handleInviteUser(user)}
										data-tooltip={!this._hasEmail(user) ? "top" : null}
										title={!this._hasEmail(user) ? "No public email" : null} />
								}
								{this.state.allowStandaloneUsers &&
									<a className="user-permissions">{this._getUserPermissionString(user)}</a>
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
	onboarding: React.PropTypes.object.isRequired,
	allowStandaloneUsers: React.PropTypes.bool.isRequired,
	isMothership: React.PropTypes.bool.isRequired,
	allowGitHubUsers: React.PropTypes.bool.isRequired,
};

export default UserList;
