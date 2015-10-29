import React from "react";
import moment from "moment";

import Component from "./Component";

// TODO merge small and large layouts
export default class DiscussionsList extends Component {
	updateState(state, props) {
		state.discussions = props.discussions;
		state.onViewDiscussion = props.onViewDiscussion;
		state.small = props.small;
	}

	render() {
		return (
			<ul className="list">
				{this.state.discussions.map((d) =>
					<li className="discussion" onClick={() => { this.state.onViewDiscussion(d); }} key={d.ID}>
						{this.state.small ? (
							<div>
								<a className="title truncate">{d.Title}</a>
								<div className="stats">
									<span className="octicon octicon-comment-discussion" /> {d.Comments ? d.Comments.length : 0}
								</div>
								<p className="body truncate">{d.Description}</p>
							</div>
						) : (
							<div>
								<header>
									<h1>
										<div className="contents">
											<a>{d.Title}</a> <span className="id">#{d.ID}</span>
										</div>
									</h1>
									<div className="stats">
										<span className="octicon octicon-comment-discussion" /> {d.Comments ? d.Comments.length : 0}
									</div>
									<div className="subtitle">
										<span className="author"><a>@{d.Author.Login}</a></span>
										<span className="date"> {moment(d.CreatedAt).fromNow()}</span>
									</div>
								</header>
								<p className="body">{d.Description && d.Description.slice(0, 250)+(d.Description.length > 250 ? "..." : "")}</p>
							</div>
						)}
					</li>
				)}
			</ul>
		);
	}
}

DiscussionsList.propTypes = {
	discussions: React.PropTypes.array,
	onViewDiscussion: React.PropTypes.func,
	small: React.PropTypes.bool,
};
