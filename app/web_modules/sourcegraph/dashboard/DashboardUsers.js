import React from "react";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

class UserList extends Component {
	constructor(props) {
		super(props);
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

	render() {
		return (
			<div className="panel panel-default">
				<div className="panel-heading">
					<h5>Team</h5>
					<button className="btn btn-primary add-user-btn"
						onClick={() => Dispatcher.dispatch(new DashboardActions.OpenAddUsersModal())} >
						<i className="fa fa-user-plus"></i>
					</button>
				</div>
				<div className="users-list panel-body">
					<div className="list-group">
						{this.state.users.map((user, i) => (
							<div className="list-group-item" key={i}>
								<img className="avatar-sm" src={user.AvatarURL || "https://secure.gravatar.com/avatar?d=mm&f=y&s=128"} />
								<span className="user-name">{user.Name || user.Login}{user.IsInvited ? " (pending)" : ""}</span>
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
};

export default UserList;
