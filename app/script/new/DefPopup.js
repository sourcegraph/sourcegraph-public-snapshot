import React from "react";
import Draggable from "react-draggable";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import ExampleView from "./ExampleView";
import DiscussionsView from "./DiscussionsView";

export default class DefPopup extends Component {
	updateState(state, props) {
		state.def = props.def;
		state.examples = props.examples;
		state.examplesGeneration = props.examples.generation;
		state.highlightedDef = props.highlightedDef;
		state.discussions = props.discussions;
	}

	render() {
		let def = this.state.def;
		return (
			<Draggable handle="header.toolbar">
				<div className="token-details">
					<div className="body">
						<header className="toolbar">
							<a className="btn btn-toolbar btn-default go-to-def" href={def.URL} onClick={(event) => {
								event.preventDefault();
								Dispatcher.dispatch(new DefActions.GoToDef(def.URL));
							}}>Go to definition</a>
							<a className="close top-action" onClick={() => {
								Dispatcher.dispatch(new DefActions.SelectDef(null));
							}}>×</a>
						</header>

						<section className="docHTML">
							<div className="header">
								<h1 className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
							</div>
							<section className="doc" dangerouslySetInnerHTML={def.Data && def.Data.DocHTML} />
						</section>

						<ExampleView defURL={def.URL} examples={this.state.examples} highlightedDef={this.state.highlightedDef} />

						{this.state.discussions && <DiscussionsView discussions={this.state.discussions.slice(0, 4)} />}
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
	discussions: React.PropTypes.array,
};
