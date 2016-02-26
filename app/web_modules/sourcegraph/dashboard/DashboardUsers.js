import React from "react";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

class UserList extends Component {
	constructor(props) {
		super(props);
		this._canAdd = this._canAdd.bind(this);
		this._reasonCannotAdd = this._reasonCannotAdd.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_selectUpdateUser() {

	}

	_updateUserPermissions() {

	}

	_getUserPermissionString(user) {
		if (user.Admin) return "Admin";
		if (user.Write) return "Write";
		return "Read";
	}

	_canAdd(user) {
		if (user.hasOwnProperty("LocalAccount")) return false;
		if (!user.hasOwnProperty("Email")) return false;
		return true;
	}

	_reasonCannotAdd(user) {
		if (!user.Email) return `User does not have a public Email`;
	}

	render() {
		const userSort = (a, b) => {
			if (a.hasOwnProperty("LocalAccount")) {
				if (a.LocalAccount.UID === window.currentUser.UID) return -100;
			}
			if (b.hasOwnProperty("LocalAccount")) {
				if (b.LocalAccount.UID === window.currentUser.UID) return 100;
			}
			if (a.hasOwnProperty("LocalAccount") && !b.hasOwnProperty("LocalAccount")) return -1;
			if (!this._canAdd(a) && this._canAdd(b)) return 1;
			if (this._canAdd(a) && !this._canAdd(b)) return -1;
			return -1;
		};

		return (
			<div className="panel panel-default">
				<div className="panel-heading">
					<h5>Team</h5>
					{!this.state.isMothership &&
						<button className="btn btn-primary add-user-btn"
							onClick={() => Dispatcher.dispatch(new DashboardActions.OpenAddUsersModal())} >
							<i className="fa fa-user-plus"></i>
						</button>
					}
				</div>
				<div className="users-list panel-body">
					<div className="list-group">
						{this.state.users.sort(userSort).map((user, i) => (
							<div className="list-group-item" key={i}>
								<img className="avatar-sm" src={user.RemoteAccount.AvatarURL || "https://secure.gravatar.com/avatar?d=mm&f=y&s=128"} />
								<span className="user-name">{user.RemoteAccount.Name || user.RemoteAccount.Login}{user.RemoteAccount.IsInvited ? " (pending)" : ""}</span>
								{this._canAdd(user) &&
								<button className="btn btn-primary add-user-btn"
									onClick={() => Dispatcher.dispatch(new DashboardActions.OpenAddUsersModal())} >
									<i className="fa fa-user-plus"></i>
								</button>
								}
								{this.state.allowStandaloneUsers &&
									<a className="user-permissions">{this._getUserPermissionString(user)}</a>
								}
							</div>
						))}
					</div>
				</div>
			</div>
		);
	}
}

UserList.propTypes = {
	users: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	allowStandaloneUsers: React.PropTypes.bool.isRequired,
	isMothership: React.PropTypes.bool.isRequired,
};

export default UserList;
