// tslint:disable

import * as React from "react";
import TimeAgo from "sourcegraph/util/TimeAgo";
import {Avatar} from "sourcegraph/components/index";
import {PencilIcon} from "sourcegraph/components/Icons";
import {DefAuthor} from "sourcegraph/def/index";
import CSSModules from "react-css-modules";
import * as styles from "./styles/AuthorList.css";

export default CSSModules(function AuthorList({
	authors,
	horizontal = false,
	className,
}: {
	authors: Array<DefAuthor>,
	horizontal?: boolean,
	className?: string,
}) {
	return (
		<div className={className}>
			{authors && authors.length > 0 &&
				<ol className={horizontal ? styles.list_horizontal : styles.list}>
					{horizontal && <PencilIcon title="Authors" className={styles.pencil_icon} />}
					{authors.map((a, i) => (
						<li key={i} className={horizontal ? styles.person_horizontal : styles.person}>
							<div className={styles.badge_wrapper}>
								<span className={styles.badge}>{Math.round(100 * a.BytesProportion) || "< 1"}%</span>
							</div>
							<Avatar className={styles.avatar} size="tiny" img={a.AvatarURL} />
							{a.Email}
							<TimeAgo time={a.LastCommitDate} className={styles.timestamp} />
						</li>
					))}
				</ol>
			}
		</div>
	);
}, styles);

