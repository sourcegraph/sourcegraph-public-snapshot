import React from "react";
import Draggable from "react-draggable";

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
			<Draggable handle="header.toolbar">
				<div className="token-details">
					<div className="body">
						<header className="toolbar">
							{def.Found &&
								<a className="btn btn-toolbar btn-default go-to-def" href={def.URL} onClick={(def.Data && def.Data.Kind !== "package") && hotLink}>Go to definition</a>
							}

							<a className="close top-action" onClick={() => {
								Dispatcher.dispatch(new DefActions.SelectDef(null));
							}}>×</a>
						</header>

						<div>
							<section className="docHTML">
								<div className="header">
									<h1 className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
								</div>
								<section className="doc">
									{def.Found ? <div dangerouslySetInnerHTML={def.Data.DocHTML} /> : <div>Definition of <div dangerouslySetInnerHTML={def.QualifiedName} /> is not available.</div>}
								</section>
							</section>

							<ExampleView defURL={def.URL} examples={this.state.examples} highlightedDef={this.state.highlightedDef} />
						</div>
					</div>
				</div>
			</Draggable>
		);
	}
}

DefPopup.propTypes = {
	def: React.PropTypes.object,
	examples: React.PropTypes.object,
	highlightedDef: React.PropTypes.string,
};

export default DefPopup;
