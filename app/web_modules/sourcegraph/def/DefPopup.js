import React from "react";

import Component from "sourcegraph/Component";
import hotLink from "sourcegraph/util/hotLink";

class DefPopup extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.path = props.path || null;
	}

	render() {
		let def = this.state.def;
		let refsURL = `${this.state.def.URL}/-/refs`;
		return (
			<div className="sidebar-section token-details">
				<section>
					<p className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
				</section>

				<header className="usage-header">Used in {!this.state.refLocations && <i className="fa fa-circle-o-notch fa-spin"></i>}</header>
				{this.state.refLocations && this.state.refLocations.length === 0 &&
					<i>No usages found</i>
				}
				{this.state.refLocations && this.state.refLocations.length > 0 &&
					this.state.refLocations.map((repoRef, i) => (
						<div key={i} className="usages">
							<header><span className="badge">{repoRef.Count}</span> <a href={`${refsURL}?Repo=${repoRef.Repo}`} onClick={hotLink}>{repoRef.Repo}</a> </header>
							<div className="usage-category">
								{repoRef.Files.map((file, j) => (
									<div key={j} className={this.state.path === file.Path ? "current-file" : ""}>
										<span className="badge">{file.Count}</span> <a href={`${refsURL}?Repo=${repoRef.Repo}&Files=${file.Path}`} onClick={hotLink}>{file.Path}</a>
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

DefPopup.propTypes = {
	def: React.PropTypes.object,
	refLocations: React.PropTypes.array,
	annotations: React.PropTypes.object,
	activeDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
};

export default DefPopup;
