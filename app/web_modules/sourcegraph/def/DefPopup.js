// @flow weak

import React from "react";
import {Link} from "react-router";
import {urlToDefRefs} from "sourcegraph/def/routes";
import s from "sourcegraph/def/styles/Def.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";

class DefPopup extends React.Component {
	static propTypes = {
		def: React.PropTypes.object.isRequired,
		refLocations: React.PropTypes.array,
		path: React.PropTypes.string.isRequired,
	};

	render() {
		let def = this.props.def;
		let refLocs = this.props.refLocations;
		return (
			<div className={s.marginBox}>
				<header className={s.boxTitle}>{qualifiedNameAndType(def)}</header>
				<header className={s.sectionTitle}>Used in {!refLocs && <i className="fa fa-circle-o-notch fa-spin"></i>}</header>
				{refLocs && refLocs.length === 0 &&
					<i>No usages found</i>
				}
				{refLocs && refLocs.length > 0 &&
					refLocs.map((repoRef, i) => (
						<div key={i} className={s.allRefs}>
							<header><span className={s.refsCount}>{repoRef.Count}</span> <Link to={urlToDefRefs(def, repoRef.Repo)}>{repoRef.Repo}</Link></header>
							<div className={s.refsGroup}>
								{repoRef.Files.map((file, j) => (
									<div key={j} className={this.props.path === file.Path ? s.currentFileRefs : ""}>
										<span className={s.refsCount}>{file.Count}</span> <Link to={urlToDefRefs(def, repoRef.Repo, file.Path)}>{file.Path}</Link>
									</div>
								))}
							</div>
						</div>
					))
				}
			</div>
		);
	}
}

export default DefPopup;
