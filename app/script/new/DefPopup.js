import React from "react";
import Draggable from "react-draggable";

import Dispatcher from "./Dispatcher";
import * as CodeActions from "./CodeActions";
import * as DefActions from "./DefActions";

class DefPopup extends React.Component {
	shouldComponentUpdate(nextProps, nextState) {
		return nextProps.def !== this.props.def;
	}

	render() {
		let def = this.props.def;
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
								Dispatcher.dispatch(new CodeActions.SelectDef(null));
							}}>×</a>
						</header>
						<section className="docHTML">
							<div className="header">
								<h1 className="qualified-name" dangerouslySetInnerHTML={def.QualifiedName} />
							</div>
							<section className="doc" dangerouslySetInnerHTML={def.Data && def.Data.DocHTML} />
						</section>
					</div>
				</div>
			</Draggable>
		);
	}
}

DefPopup.propTypes = {
	def: React.PropTypes.object,
};

export default DefPopup;
