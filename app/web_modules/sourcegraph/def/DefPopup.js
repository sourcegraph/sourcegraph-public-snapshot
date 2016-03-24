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
		let examplesURL = `${this.state.def.URL}/.examples`;
		return (
			<div className="sidebar-section token-details">
				<section>
					<p className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
				</section>

				<header className="usage-header">Usages {!this.state.refs && <i className="fa fa-circle-o-notch fa-spin"></i>}</header>
				{this.state.refs && this.state.refs.Total === 0 &&
					<i>No usages found</i>
				}
				{this.state.refs && this.state.refs.Files &&
					<div className="usages">
						<div><i className="fa fa-bookmark"></i> <a href={examplesURL} onClick={hotLink}>{def.Data.Repo}</a> ({this.state.refs.Total})</div>
						<div className="usage-category">
							{this.state.refs.Files.map((file, i) => (
								<div key={i}>
									<i className="fa fa-file-text-o"></i> <a href={`${examplesURL}?file=${file.Name}`} onClick={hotLink}>{this.state.path === file.Name ? "Current File" : file.Name}</a> ({file.RefCount})
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
