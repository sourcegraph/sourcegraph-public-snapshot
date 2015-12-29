import moment from "moment";
import React from "react";

import Component from "sourcegraph/Component";
import TimeAgo from "sourcegraph/util/TimeAgo";

class BuildHeader extends Component {
	reconcileState(state, props) {
		if (state.build !== props.build) {
			state.build = props.build;
		}

		if (state.commit !== props.commit) {
			state.commit = props.commit;
		}
	}

	render() {
		return (
			<div>
						<div className="commit single media repo-build">
							<a className="pull-left">
								<img className="media-object avatar img-rounded" src={this.state.commit.AuthorPerson.AvatarURL} />
							</a>
							<div className="media-body">
								<h4 className="media-heading commit-title">
									<a href={`/${this.state.build.Repo}/.commits/${this.state.commit.ID}`}>{this.state.commit.Message.slice(0, 70)}</a>
								</h4>
								<p className="author committer">
									<span className="date">authored <TimeAgo time={this.state.commit.Author.Date} /></span>
									{this.state.commit.Committer ? <span className="date">, committed <TimeAgo time={this.state.commit.Committer.Date} /></span> : null}
									<a href={`/${this.state.build.Repo}/.commits/${this.state.commit.ID}`}>
										<tt className="commit-id pull-right">{this.state.commit.ID.substring(0, 6)}</tt>
									</a>
								</p>
					</div>
				</div>
				<div className="row">
					<div className="col-md-12">
						<div className="media">
							<header className="repo-build pull-left">
								<h1 className={`label label-${buildClass(this.state.build)}`}>
									<span className="number">#{this.state.build.ID}</span>
									<span className="status">{buildStatus(this.state.build)}</span>
								</h1>
							</header>
							<div className="media-body">
								<table className="table repo-build table-condensed">
									<tbody>
										{(!this.state.build.StartedAt && !this.state.build.EndedAt) ? <tr><th>Created</th><td><TimeAgo time={this.state.build.CreatedAt} /></td></tr> : null}
										{(this.state.build.StartedAt && !this.state.build.EndedAt) ? <tr><th>Started</th><td><TimeAgo time={this.state.build.StartedAt} /></td></tr> : null}
										{this.state.build.EndedAt ? <tr><th>Ended</th><td><TimeAgo time={this.state.build.EndedAt} /></td></tr> : null}
										{this.state.build.StartedAt ? <tr><th>Elapsed</th><td>{moment.duration(moment(this.state.build.EndedAt).diff(this.state.build.StartedAt)).as("seconds")}s</td></tr> : null}
									</tbody>
								</table>
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	}
}

BuildHeader.propTypes = {
	build: React.PropTypes.object.isRequired,
	commit: React.PropTypes.object.isRequired,
};

// buildStatus returns a textual status description for the build.
function buildStatus(b) {
	if (b.Failure) {
		return "Failed";
	}
	if (b.Success) {
		return "Succeeded";
	}
	if (b.StartedAt !== null && !b.EndedAt) {
		return "In progress";
	}
	return "Queued";
}

// buildClass returns the CSS class for the build.
function buildClass(b) {
	switch (buildStatus(b)) {
	case "Failed":
		return "danger";
	case "Succeeded":
		return "success";
	case "In progress":
		return "info";
	}
	return "default";
}

export default BuildHeader;
