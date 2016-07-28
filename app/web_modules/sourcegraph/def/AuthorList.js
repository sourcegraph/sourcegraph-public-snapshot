// @flow

import * as React from "react";
import TimeAgo from "sourcegraph/util/TimeAgo";
import {Avatar} from "sourcegraph/components";
import {PencilIcon} from "sourcegraph/components/Icons";
import type {DefAuthor} from "sourcegraph/def";
import CSSModules from "react-css-modules";
import styles from "./styles/AuthorList.css";

export default CSSModules(function AuthorList({
	authors,
	horizontal = false,
	className,
}: {
	authors: Array<DefAuthor>,
	horizontal: bool,
	className?: string,
}) {
	return (
		<div className={className}>
			{authors && authors.length > 0 &&
				<ol styleName={`list${horizontal ? "-horizontal" : ""}`}>
					{horizontal && <PencilIcon title="Authors" styleName="pencil-icon" />}
					{authors.map((a, i) => (
						<li key={i} styleName={`person${horizontal ? "-horizontal" : ""}`}>
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
}, styles);

