import React from "react";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import classNames from "classnames";

class UserList extends Component {
	constructor(props) {
		super(props);
		this._isCurrentUser = this._isCurrentUser.bind(this);
		this._existsLocally = this._existsLocally.bind(this);
		this._isInvited = this._isInvited.bind(this);
		this._hasEmail = this._hasEmail.bind(this);
		this._avatarURL = this._avatarURL.bind(this);
		this._name = this._name.bind(this);
		this._handleInviteUser = this._handleInviteUser.bind(this);
		this._handleInviteAll = this._handleInviteAll.bind(this);
		this._userSort = this._userSort.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
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

	_handleInviteUser(user) {
		Dispatcher.dispatch(new DashboardActions.WantInviteUsers([user.Email]));
	}

	_handleInviteAll() {
		const emails = this.state.users
			.filter(user => this._hasEmail(user) && !this._isInvited(user) && !this._existsLocally(user))
			.map(user => user.Email);
		if (emails.length > 0) Dispatcher.dispatch(new DashboardActions.WantInviteUsers(emails));
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
		return (
			<div className="panel panel-default">
				<div className="panel-heading">
					<div className="panel-heading-content">
						<h5>Team</h5>
						{!this.state.onboarding.linkGitHub &&
							<i className="btn-icon sg-icon-plus-box"
								onClick={this._handleInviteAll}
								data-tooltip={true}
								title="Invite all teammates" />
						}
					</div>
				</div>
				<div className="users-list panel-body">
					{this.state.users.length === 0 ?
						<div className="well empty-well">{"Link your GitHub account to add teammates."}</div> :
						<div className="list-group">
							{this.state.users.sort(this._userSort).map((user, i) => (
								<div className="list-group-item" key={i}>
									<img className="avatar avatar-sm" src={this._avatarURL(user) || "https://secure.gravatar.com/avatar?d=mm&f=y&s=128"} />
									<div className="user-name">{this._name(user)}</div>
									{!this._existsLocally(user) && !this._isInvited(user) &&
										<i className={classNames("sg-icon sg-icon-plus-box btn-icon", {"add-user-icon-disabled": !this._hasEmail(user)})}
											onClick={() => this._handleInviteUser(user)}
											data-tooltip={!this._hasEmail(user) ? "top" : null}
											title={!this._hasEmail(user) ? "No public email" : null} />
									}
								</div>
							))}
						</div>
					}
				</div>
			</div>
		);
	}
}

UserList.propTypes = {
	users: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	currentUser: React.PropTypes.object.isRequired,
	onboarding: React.PropTypes.object.isRequired,
};

export default UserList;
