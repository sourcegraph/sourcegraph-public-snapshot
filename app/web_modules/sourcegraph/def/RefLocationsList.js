// @flow

import React from "react";
import {Link} from "react-router";
import {urlToDefRefs} from "sourcegraph/def/routes";
import s from "sourcegraph/def/styles/Def.css";

export default class RefLocationsList extends React.Component {
	static propTypes = {
		def: React.PropTypes.object.isRequired,
		refLocations: React.PropTypes.array,

		// Current repo and path, so that they can be highlighted.
		repo: React.PropTypes.string.isRequired,
		path: React.PropTypes.string.isRequired,
	};

	render() {
		let def = this.props.def;
		let refLocs = this.props.refLocations;

		if (!refLocs || refLocs.length === 0) return null;

		return (
			<div>
				{refLocs.filter((r) => r && r.Files).map((repoRef, i) => (
					<div key={i} className={s.allRefs}>
						<header>
							<span className={s.refsCount}>{repoRef.Count}</span> <Link to={urlToDefRefs(def, repoRef.Repo)}>{repoRef.Repo}</Link>
						</header>
						<div className={s.refsGroup}>
							{repoRef.Files.map((file, j) => (
								<div key={j} className={this.props.repo === repoRef.Repo && this.props.path === file.Path ? s.currentFileRefs : ""}>
									<span className={s.refsCount}>{file.Count}</span> <Link to={urlToDefRefs(def, repoRef.Repo, file.Path)}>{file.Path}</Link>
								</div>
							))}
						</div>
					</div>
				))}
			</div>
		);
	}
}
