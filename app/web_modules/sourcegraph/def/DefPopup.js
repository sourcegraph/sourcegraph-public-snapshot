import React from "react";

import Component from "sourcegraph/Component";
import ExampleView from "sourcegraph/def/ExampleView";

class DefPopup extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		let def = this.state.def;
		return (
			<div className="token-details card">
				<section>
					<p className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
				</section>

				<header>Usage Examples</header>
				<section className="examples-card">
					<ExampleView
						defURL={def.URL}
						examples={this.state.examples}
						annotations={this.state.annotations}
						activeDef={this.state.activeDef}
						highlightedDef={this.state.highlightedDef} />
				</section>
			</div>
		);
	}
}

DefPopup.propTypes = {
	def: React.PropTypes.object,
	examples: React.PropTypes.object,
	annotations: React.PropTypes.object,
	activeDef: React.PropTypes.string,
	highlightedDef: React.PropTypes.string,
};

export default DefPopup;
