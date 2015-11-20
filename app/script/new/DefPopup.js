import React from "react";
import Draggable from "react-draggable";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import ExampleView from "./ExampleView";
import hotLink from "./util/hotLink";

export default class DefPopup extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.examplesGeneration = props.examples.generation;
	}

	render() {
		let def = this.state.def;
		return (
			<Draggable handle="header.toolbar">
				<div className="token-details">
					<div className="body">
						<header className="toolbar">
							{def.Found &&
								<a className="btn btn-toolbar btn-default go-to-def" href={def.URL} onClick={hotLink}>Go to definition</a>
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
									{def.Found ? <span dangerouslySetInnerHTML={def.Data.DocHTML} /> : <span>Definition of <span dangerouslySetInnerHTML={def.QualifiedName} /> is not available.</span>}
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
