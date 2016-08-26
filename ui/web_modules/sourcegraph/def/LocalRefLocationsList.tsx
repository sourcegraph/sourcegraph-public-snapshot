import * as React from "react";
import {Link} from "react-router";
import {LocalRefLocationsKey} from "sourcegraph/def/index";
import {urlToRepoBlob} from "sourcegraph/def/routes";
import * as styles from "sourcegraph/def/styles/Def.css";

type Props = {
	refLocations?: LocalRefLocationsKey,
	showMax?: number,

	// Current repo and path info, so that they can be highlighted.
	repo: string,
	rev: string | null,
	path?: string,
}

export class LocalRefLocationsList extends React.Component<Props, any> {

	render(): JSX.Element | null {
		let refLocs = this.props.refLocations;

		if (!refLocs) {
			return null;
		}

		return (
			<div>
				{refLocs.Files && refLocs.Files.map((fileRef, i) => (
					this.props.showMax && i >= this.props.showMax ? null : <div key={i}>
						<Link to={urlToRepoBlob(this.props.repo, this.props.rev, fileRef.Path)}>
							<header className={this.props.path === fileRef.Path ? styles.b : ""}>
								<span className={styles.refs_count}>{fileRef.Count}</span> <span>{fileRef.Path}</span>
							</header>
						</Link>
					</div>
				))}
			</div>
		);
	}
}
