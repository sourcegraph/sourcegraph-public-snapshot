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
			<div className="token-details">
				<div className="card">
					<header>
						<h1>
							{def.Found &&
								<a className="btn btn-toolbar btn-default go-to-def" href={def.URL} onClick={(def.Data && def.Data.Kind !== "package") && hotLink}>Go to definition</a>
							}
							<span className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
							<a className="close top-action" onClick={() => {
								Dispatcher.dispatch(new DefActions.SelectDef(null));
							}}>×</a>
						</h1>
					</header>
				</div>

				<section className="examples-card card">
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
