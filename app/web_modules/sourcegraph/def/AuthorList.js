// @flow weak

import React from "react";
import TimeAgo from "sourcegraph/util/TimeAgo";
import {Avatar} from "sourcegraph/components";
import styles from "sourcegraph/def/styles/Def.css";

class AuthorList extends React.Component {
	static propTypes = {
		authors: React.PropTypes.object.isRequired,
	};

	render() {
		let authors = this.props.authors ? this.props.authors.DefAuthors || [] : null;

		return (
			<div>
				{authors && authors.length === 0 &&
					<i>No authors found</i>
				}
				{authors && authors.length > 0 &&
					<ol className={styles.personList}>
						{authors.map((a, i) => (
							<li key={i} className={styles.author}>
								<span className={styles.badgeMinWidthWrapper}><span className={styles.bytesProportion}>{Math.round(100 * a.BytesProportion)}%</span></span> <Avatar size="tiny" img={a.AvatarURL} /> {a.Email} <TimeAgo time={a.LastCommitDate} className={styles.timestamp} />
							</li>
						))}
					</ol>
				}
			</div>
		);
	}
}

export default AuthorList;
