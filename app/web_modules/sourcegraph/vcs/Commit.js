import React from "react";

import Component from "sourcegraph/Component";
import TimeAgo from "sourcegraph/util/TimeAgo";

class Commit extends Component {
	reconcileState(state, props) {
		if (state.commit !== props.commit) {
			state.commit = props.commit;
		}
	}

	render() {
		return (
			<div className="commit single media repo-build">
				<a className="pull-left">
					<img className="media-object avatar img-rounded" src={this.state.commit.AuthorPerson.AvatarURL} />
				</a>
				<div className="media-body">
					<h4 className="media-heading commit-title">
						<a href={`/${this.state.commit.RepoURI}/.commits/${this.state.commit.ID}`}>
							{this.state.commit.Message.slice(0, 70)}
						</a>
					</h4>
					<p className="author committer">
						<span className="date">authored <TimeAgo time={this.state.commit.Author.Date} /></span>
						{this.state.commit.Committer ? <span className="date">, committed <TimeAgo time={this.state.commit.Committer.Date} /></span> : null}
						<a href={`/${this.state.commit.RepoURI}/.commits/${this.state.commit.ID}`}>
							<tt className="commit-id pull-right">{this.state.commit.ID.substring(0, 6)}</tt>
						</a>
					</p>
				</div>
			</div>
		);
	}
}

Commit.propTypes = {
	commit: React.PropTypes.object.isRequired,
};

export default Commit;
