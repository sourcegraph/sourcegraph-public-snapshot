import React from "react";
import update from "react/lib/update";

import Component from "sourcegraph/Component";
import ImportGitHubReposMenu from "sourcegraph/dashboard/ImportGitHubReposMenu";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";

class AddReposWidget extends Component {
	constructor(props) {
		super(props);
		this.state = {
			repoName: "",
		};
		this._handleTextInput = this._handleTextInput.bind(this);
		this._handleCreate = this._handleCreate.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_handleTextInput(e) {
		this.setState(update(this.state, {
			repoName: {$set: e.target.value},
		}));
	}

	_handleCreate() {
		Dispatcher.dispatch(new DashboardActions.WantCreateRepo(this.state.repoName));
		this.state.dismissModal();
	}

	render() {
		console.log("and re-rendering add repos widget");
		return (
			<div className="modal add-repos-widget"
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
							<h4 className="modal-title">Add a new Repository</h4>
						</div>
						<div className="modal-body">
							<ul className="nav nav-tabs" role="tablist">
								<li role="presentation" className="active">
									<a href="#new-repo" role="tab" data-toggle="tab">Create New</a>
								</li>
								<li role="presentation">
									<a href="#github-mirror" role="tab" data-toggle="tab">Import from GitHub</a>
								</li>
							</ul>

							<div className="tab-content">
								{this.state.allowStandaloneRepos &&
									<div role="tabpanel" className="tab-pane active" id="new-repo">
										<div className="widget-body">
											<p className="add-repo-label">REPO NAME:</p>
											<input className="form-control"
												type="text"
												value={this.state.repoName}
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
								}
								{this.state.allowGitHubMirrors &&
									<div role="tabpanel" className="tab-pane" id="github-mirror">
										<ImportGitHubReposMenu />
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

AddReposWidget.propTypes = {
	dismissModal: React.PropTypes.func.isRequired,
	allowStandaloneRepos: React.PropTypes.bool.isRequired,
	allowGitHubMirrors: React.PropTypes.bool.isRequired,
};

export default AddReposWidget;
