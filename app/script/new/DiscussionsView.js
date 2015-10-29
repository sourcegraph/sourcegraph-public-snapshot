import React from "react";

import Component from "./Component";

export default class DiscussionsView extends Component {
	updateState(state, props) {
		state.discussions = props.discussions;
	}

	render() {
		return (
			<div className="code-discussions">
				{this.state.discussions.length === 0 ? (
					<div className="no-discussions"><a ref="createBtn"><i className="octicon octicon-plus" /> Start a code discussion</a></div>
				) : (
					<div className="contents">
						<ul className="list">
							{this.state.discussions.map((d) =>
								<li className="discussion" key={`snippet-d-${d.ID}`}>
									<a className="title truncate">{d.Title}</a>
									<div className="stats">
										<span className="octicon octicon-comment-discussion" /> {d.Comments ? d.Comments.length : 0}
									</div>
									<p className="body truncate">{d.Description}</p>
								</li>
							)}
						</ul>
						<footer>
							<a ref="listBtn"><i className="fa fa-eye" /> View all</a>
							<a ref="createBtn"><i className="fa fa-comment" /> New</a>
						</footer>
					</div>
				)}
			</div>
		);
	}
}

DiscussionsView.propTypes = {
	discussions: React.PropTypes.array,
};
