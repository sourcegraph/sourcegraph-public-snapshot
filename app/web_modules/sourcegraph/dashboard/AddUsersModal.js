import React from "react";
import update from "react/lib/update";

import Component from "sourcegraph/Component";
import ImportGitHubUsersMenu from "sourcegraph/dashboard/ImportGitHubUsersMenu";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";

class AddUsersModal extends Component {
	constructor(props) {
		super(props);
		this.state = {
			email: "",
			permission: "read",
		};
		this._handleTextInput = this._handleTextInput.bind(this);
		this._handleInvite = this._handleInvite.bind(this);
		this._handlePermissionChange = this._handlePermissionChange.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_handleTextInput(e) {
		this.setState(update(this.state, {
			email: {$set: e.target.value},
		}));
	}

	_handlePermissionChange(e) {
		this.setState(update(this.state, {
			permission: {$set: e.target.value},
		}));
	}

	_handleInvite() {
		Dispatcher.dispatch(new DashboardActions.WantInviteUser(this.state.email, this.state.permission));
		Dispatcher.dispatch(new DashboardActions.DismissUsersModal());
	}

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
								onClick={() => Dispatcher.dispatch(new DashboardActions.DismissUsersModal())}>
								<span aria-hidden="true">&times;</span>
							</button>
							<h4 className="modal-title">Invite People to join Sourcegraph</h4>
						</div>
						<div className="modal-body">
							<ul className="nav nav-tabs" role="tablist">
								{this.state.allowStandaloneUsers &&
									<li role="presentation" className={this.state.allowStandaloneUsers ? "active" : ""}>
										<a href="#email-invite" role="tab" data-toggle="tab">Invite by Email</a>
									</li>
								}
								{this.state.allowGitHubUsers &&
									<li role="presentation" className={!this.state.allowStandaloneUsers ? "active" : ""}>
										<a href="#github-invite" role="tab" data-toggle="tab">Invite from GitHub</a>
									</li>
								}
							</ul>

							<div className="tab-content">
								{this.state.allowStandaloneUsers &&
									<div role="tabpanel" className={`tab-pane ${this.state.allowStandaloneUsers ? "active" : ""}`} id="email-invite">
										<div className="widget-body">
											<p className="add-repo-label">EMAIL:</p>
											<div className="form-inline invite-user-form">
												<input className="form-control"
													type="text"
													value={this.state.email}
													placeholder="Type email here"
													onChange={this._handleTextInput}/>
												<select className="form-control"
													onChange={this._handlePermissionChange}>
													<option value="read">Can Read</option>
													<option value="write">Can Write</option>
													<option value="admin">Admin</option>
												</select>
											</div>
										</div>
										<div className="widget-footer">
											<button className="btn btn-block btn-primary btn-lg"
												onClick={this._handleInvite}>
												CREATE
											</button>
										</div>
									</div>
								}
								{this.state.allowGitHubUsers &&
									<div role="tabpanel" className={`tab-pane ${!this.state.allowStandaloneUsers ? "active" : ""}`} id="github-invite">
										<ImportGitHubUsersMenu />
									</div>
								}
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	}
}

AddUsersModal.propTypes = {
	allowStandaloneUsers: React.PropTypes.bool.isRequired,
	allowGitHubUsers: React.PropTypes.bool.isRequired,
};

export default AddUsersModal;
