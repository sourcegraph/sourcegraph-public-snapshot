// @flow weak

import React from "react";
import TimeAgo from "sourcegraph/util/TimeAgo";
import {Avatar} from "sourcegraph/components";
import {PencilIcon} from "sourcegraph/components/Icons";

import CSSModules from "react-css-modules";
import styles from "./styles/AuthorList.css";

class AuthorList extends React.Component {
	static propTypes = {
		authors: React.PropTypes.object.isRequired,
		horizontal: React.PropTypes.bool,
		className: React.PropTypes.string,
	};

	render() {
		let authors = this.props.authors ? this.props.authors.DefAuthors || [] : null;

		return (
			<div className={this.props.className}>
				{authors && authors.length === 0 &&
					<i>No authors found</i>
				}
				{authors && authors.length > 0 &&
					<ol styleName={`list${this.props.horizontal ? "-horizontal" : ""}`}>
						{this.props.horizontal && <PencilIcon title="Authors" styleName="pencil-icon" />}
						{authors.map((a, i) => (
							<li key={i} styleName={`person${this.props.horizontal ? "-horizontal" : ""}`}>
								<div styleName="badge-wrapper">
									<span styleName="badge">{Math.round(100 * a.BytesProportion) || "< 1"}%</span>
								</div>
								<Avatar styleName="avatar" size="tiny" img={a.AvatarURL} />
								{a.Email}
								<TimeAgo time={a.LastCommitDate} styleName="timestamp" />
							</li>
						))}
					</ol>
				}
			</div>
		);
	}
}

export default CSSModules(AuthorList, styles);
