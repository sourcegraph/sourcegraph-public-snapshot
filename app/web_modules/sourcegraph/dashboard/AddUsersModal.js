import React from "react";
import update from "react/lib/update";

import Component from "sourcegraph/Component";
import ImportGitHubUsersMenu from "sourcegraph/dashboard/ImportGitHubUsersMenu";

class AddUsersModal extends Component {
	constructor(props) {
		super(props);
		this.state = {
			email: "",
		};
		this._handleTextInput = this._handleTextInput.bind(this);
		this._handleCreate = this._handleCreate.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
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
									<ImportGitHubUsersMenu />
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
};

export default AddUsersModal;
