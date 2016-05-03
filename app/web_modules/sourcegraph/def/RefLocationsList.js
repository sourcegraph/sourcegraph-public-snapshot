// @flow

import React from "react";
import {Link} from "react-router";
import {urlToDefInfo} from "sourcegraph/def/routes";
import s from "sourcegraph/def/styles/Def.css";

export default class RefLocationsList extends React.Component {
	static propTypes = {
		def: React.PropTypes.object.isRequired,
		refLocations: React.PropTypes.array,

		// Current repo and path info, so that they can be highlighted.
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string.isRequired,
		path: React.PropTypes.string,
	};

	render() {
		let def = this.props.def;
		let refLocs = this.props.refLocations;

		if (!refLocs) return null;

		return (
			<div>
				{refLocs.map((repoRef, i) => (
					<div key={i} className={s.allRefs}>
						<header className={this.props.repo === repoRef.Repo ? s.activeGroupHeader : ""}>
							<span className={s.refsCount}>{repoRef.Count}</span> <Link to={urlToDefInfo(def, this.props.rev)}>{repoRef.Repo}</Link>
						</header>
						<div className={s.refsGroup}>
							{repoRef.Files && repoRef.Files.map((file, j) => (
								<div key={j} className={`${s.refFilename} ${this.props.repo === repoRef.Repo && this.props.path === file.Path ? s.currentFileRefs : ""}`}>
									<span className={s.refsCount}>{file.Count}</span> <Link title={file.Path} to={urlToDefInfo(def, this.props.rev)}>{file.Path}</Link>
								</div>
							))}
						</div>
					</div>
				))}
			</div>
		);
	}
}
