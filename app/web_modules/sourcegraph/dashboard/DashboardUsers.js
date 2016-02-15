import React from "react";

import Component from "sourcegraph/Component";

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

	render() {
		return (
			<div className="panel panel-default">
				<div className="panel-heading">
					<h5>Team</h5>
					<button className="btn btn-primary btn-block add-user-btn">
						<i className="fa fa-user-plus"></i>
					</button>
				</div>
				<div className="users-list panel-body">
					<div className="list-group">
						{this.state.users.map((user, i) => (
							<div className="list-group-item" key={i}>
								<img className="avatar-sm" src="http://placekitten.com/g/24/24" />
								<span className="user-name">Pete Nichols</span>
								<a className="user-permissions" onClick={this._selectUpdateUser(1)}>Admin</a>
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
};

export default UserList;
