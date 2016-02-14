import React from "react";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import ExampleView from "sourcegraph/def/ExampleView";

class DefPopup extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		let def = this.state.def;
		return (
			<div className="token-details card">
				<div className="body">
					<header className="docHTML">
						<div className="header">
							<h1>
								<span className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
								<a className="close top-action" onClick={() => {
									Dispatcher.dispatch(new DefActions.SelectDef(null));
								}}>×</a>
							</h1>
						</div>
					</header>
					<header className="examples-header">Uses</header>

					<ExampleView defURL={def.URL} examples={this.state.examples} highlightedDef={this.state.highlightedDef} />
				</div>
			</div>
		);
	}
}

DefPopup.propTypes = {
	def: React.PropTypes.object,
	examples: React.PropTypes.object,
	highlightedDef: React.PropTypes.string,
};

export default DefPopup;
