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
		let refsURL = `${this.state.def.URL}/.refs`;
		return (
			<div className="sidebar-section token-details">
				<section>
					<p className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
				</section>

				<header className="usage-header">Used in {!this.state.refs && <i className="fa fa-circle-o-notch fa-spin"></i>}</header>
				{this.state.refs && this.state.refs.Total === 0 &&
					<i>No usages found</i>
				}
				{this.state.refs && this.state.refs.Files && this.state.refs.Total > 0 &&
					<div className="usages">
						<header><span className="badge">{this.state.refs.Total}</span> <a href={refsURL} onClick={hotLink}>{def.Repo}</a> </header>
						<div className="usage-category">
							{this.state.refs.Files.map((file, i) => (
								<div key={i} className={this.state.path === file.Name ? "current-file" : ""}>
									<span className="badge">{file.RefCount}</span> <a href={`${refsURL}?Files=${file.Name}`} onClick={hotLink}>{file.Name}</a>
								</div>
							))}
						</div>
					</div>
				}
			</div>
		);
	}
}

DefPopup.propTypes = {
	def: React.PropTypes.object,
	refs: React.PropTypes.object,
	annotations: React.PropTypes.object,
	activeDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
};

export default DefPopup;
