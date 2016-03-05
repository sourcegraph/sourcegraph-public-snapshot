import React from "react";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import ExampleView from "sourcegraph/def/ExampleView";
import hotLink from "sourcegraph/util/hotLink";

class DefPopup extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		let def = this.state.def;
		return (
			<div className="token-details card">
				<header>
					{def.Found &&
						<a className="go-to-def" href={def.URL} onClick={(def.Data && def.Data.Kind !== "package") && hotLink}><i className="fa fa-caret-square-o-right"></i>&nbsp; Go to definition</a>
					}
				</header>
				<section>
					<p className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
				</section>

				<header>Usage Examples</header>
				<section className="examples-card">
					<ExampleView defURL={def.URL} examples={this.state.examples} annotations={this.state.annotations} highlightedDef={this.state.highlightedDef} />
				</section>
			</div>
		);
	}
}

DefPopup.propTypes = {
	def: React.PropTypes.object,
	examples: React.PropTypes.object,
	annotations: React.PropTypes.object,
	highlightedDef: React.PropTypes.string,
};

export default DefPopup;
