// @flow weak

import React from "react";
import {Link} from "react-router";
import {urlToDefRefs} from "sourcegraph/def/routes";
import s from "sourcegraph/def/styles/Def.css";

class DefPopup extends React.Component {
	static propTypes = {
		def: React.PropTypes.object.isRequired,
		refs: React.PropTypes.object,
		path: React.PropTypes.string.isRequired,
	};

	render() {
		let def = this.props.def;
		let refs = this.props.refs;
		return (
			<div className={s.marginBox}>
				<header className={s.boxTitle} dangerouslySetInnerHTML={def.QualifiedName} />

				<header className={s.sectionTitle}>Used in {!refs && <i className="fa fa-circle-o-notch fa-spin"></i>}</header>
				{refs && refs.Total === 0 &&
					<i>No usages found</i>
				}
				{refs && refs.Files && refs.Total > 0 &&
					<div className={s.allRefs}>
						<header><span className={s.refsCount}>{refs.Total}</span> <Link to={urlToDefRefs(def)}>{def.Repo}</Link></header>
						<div className={s.refsGroup}>
							{refs.Files.map((file, i) => (
								<div key={i} className={this.props.path === file.Name ? s.currentFileRefs : ""}>
									<span className={s.refsCount}>{file.RefCount}</span> <Link to={urlToDefRefs(def, file.Name)}>{file.Name}</Link>
								</div>
							))}
						</div>
					</div>
				}
			</div>
		);
	}
}

export default DefPopup;
