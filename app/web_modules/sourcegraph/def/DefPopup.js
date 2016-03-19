import React from "react";

import Component from "sourcegraph/Component";

class DefPopup extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.path = props.path || null;
	}

	render() {
		let def = this.state.def;
		return (
			<div className="sidebar-section token-details">
				<section>
					<p className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
				</section>

				<header className="usage-header">Usages</header>
				{this.state.refs && this.state.refs.Total === 0 &&
					<i>No usages found</i>
				}
				{this.state.refs && this.state.refs.Files &&
					<div>
						{this.state.refs.Files.map((file, i) => (
							<div key={i}>
								<i className="fa fa-file-text-o"></i> {this.state.path === file.Name ?
									"Current File" : file.Name} ({file.RefCount})
							</div>
						))}
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
