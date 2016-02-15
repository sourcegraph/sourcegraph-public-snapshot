import React from "react";

import Container from "sourcegraph/Container";
import GitHubReposStore from "sourcegraph/dashboard/GitHubReposStore";
import SelectableListWidget from "sourcegraph/dashboard/SelectableListWidget";

class AddReposWidget extends Container {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.mirrorRepos = GitHubReposStore.mirrorRepos;
		state.selectedRepos = GitHubReposStore.selectedRepos;
		state.currentOrg = GitHubReposStore.currentOrg;
		state.orgs = GitHubReposStore.orgs;
		state.selectAll = GitHubReposStore.selectAll;
	}

	stores() { return [GitHubReposStore]; }

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
								<div role="tabpanel" className="tab-pane active" id="new-repo">

								</div>
								<div role="tabpanel" className="tab-pane" id="github-mirror">
									<SelectableListWidget items={this.state.mirrorRepos}
										currentCategory={this.state.currentOrg}
										menuCategories={this.state.orgs}
										selections={this.state.selectedRepos}
										selectAll={this.state.selectAll}
										menuLabel="Organizations"
										onSubmit={(items) => console.log("submitted", items)} />
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	}
}

AddReposWidget.propTypes = {
	dismissModal: React.PropTypes.func,
};

export default AddReposWidget;
