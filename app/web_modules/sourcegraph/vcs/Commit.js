import React from "react";

import Component from "sourcegraph/Component";
import TimeAgo from "sourcegraph/util/TimeAgo";

import {Avatar} from "sourcegraph/components";

import CSSModules from "react-css-modules";
import styles from "./styles/Commit.css";

class Commit extends Component {
	reconcileState(state, props) {
		if (state.commit !== props.commit) {
			state.commit = props.commit;
		}
		if (state.repo !== props.repo) {
			state.repo = props.repo;
		}
	}

	render() {
		return (
			<div styleName="container">
				<div styleName="avatar">
					<Avatar img={this.state.commit.AuthorPerson ? this.state.commit.AuthorPerson.AvatarURL : ""} size="large" />
				</div>
				<div styleName="body">
					<div styleName="title">
						<code styleName="sha">{this.state.commit.ID.substring(0, 6)}</code>
						{this.state.commit.Message.slice(0, 70)}
					</div>
					<div styleName="text">
						<span>authored <TimeAgo time={this.state.commit.Author.Date} /></span>
						{this.state.commit.Committer ? <span>, committed <TimeAgo time={this.state.commit.Committer.Date} /></span> : null}
					</div>
				</div>
			</div>
		);
	}
}

Commit.propTypes = {
	commit: React.PropTypes.object.isRequired,
	repo: React.PropTypes.string.isRequired,
};

export default CSSModules(Commit, styles);
