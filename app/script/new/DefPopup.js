import React from "react";
import Draggable from "react-draggable";

import Component from "./Component";
import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import ExampleView from "./ExampleView";
import DiscussionsList from "./DiscussionsList";

export default class DefPopup extends Component {
	constructor(props) {
		super(props);
		this.state = {
			viewAllDiscussions: false,
		};
	}

	updateState(state, props) {
		state.def = props.def;
		state.examples = props.examples;
		state.examplesGeneration = props.examples.generation;
		state.highlightedDef = props.highlightedDef;
		state.discussions = props.discussions;
	}

	_renderBody() {
		let def = this.state.def;

		if (this.state.viewAllDiscussions) {
			return (
				<div className="discussions-list discussions">
					<div className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
					<div className="container">
						<div className="padded-form">
							<DiscussionsList discussions={this.state.discussions} />
						</div>
					</div>
					<footer>
						<a ref="createBtn"><i className="fa fa-comment" /> New</a>
					</footer>
				</div>
			);
		}

		return (
			<div>
				<section className="docHTML">
					<div className="header">
						<h1 className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
					</div>
					<section className="doc" dangerouslySetInnerHTML={def.Data && def.Data.DocHTML} />
				</section>

				<ExampleView defURL={def.URL} examples={this.state.examples} highlightedDef={this.state.highlightedDef} />

				{this.state.discussions &&
					<div className="code-discussions">
						{this.state.discussions.length === 0 ? (
							<div className="no-discussions"><a ref="createBtn"><i className="octicon octicon-plus" /> Start a code discussion</a></div>
						) : (
							<div className="contents">
								<DiscussionsList discussions={this.state.discussions.slice(0, 4)} small={true} />
								<footer>
									<a ref="listBtn" onClick={() => { this.setState({viewAllDiscussions: true}); }}><i className="fa fa-eye" /> View all</a>
									<a ref="createBtn"><i className="fa fa-comment" /> New</a>
								</footer>
							</div>
						)}
					</div>
				}
			</div>
		);
	}

	render() {
		let def = this.state.def;
		return (
			<Draggable handle="header.toolbar">
				<div className="token-details">
					<div className="body">
						<header className="toolbar">
							{this.state.viewAllDiscussions &&
								<a key="back-to-main" className="btn btn-toolbar btn-default" onClick={() => { this.setState({viewAllDiscussions: false}); }}>
									<span className="octicon octicon-arrow-left" /> Back to token
								</a>
							}

							<a className="btn btn-toolbar btn-default go-to-def" href={def.URL} onClick={(event) => {
								event.preventDefault();
								Dispatcher.dispatch(new DefActions.GoToDef(def.URL));
							}}>Go to definition</a>

							<a className="close top-action" onClick={() => {
								Dispatcher.dispatch(new DefActions.SelectDef(null));
							}}>×</a>
						</header>

						{this._renderBody()}
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
