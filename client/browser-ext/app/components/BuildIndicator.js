import React from "react";
import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import * as Actions from "../actions";
import {SourcegraphIcon} from "../components/Icons";
import {keyFor} from "../reducers/helpers";
import * as utils from "../utils";


@connect(
	(state) => ({
		srclibDataVersion: state.srclibDataVersion,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
export default class BuildIndicator extends React.Component {
	static propTypes = {
		path: React.PropTypes.string.isRequired,
		srclibDataVersion: React.PropTypes.object.isRequired,
		actions: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._updateIntervalID = null;
		this.state = utils.parseURL();
	}

	render() {
		return (<span>
			{/*<SourcegraphIcon style={{marginTop: "-2px", paddingLeft: "5px", fontSize: "16px"}} />*/}
			{/*<span id="sourcegraph-build-indicator-text" style={{paddingLeft: "5px"}}>{indicatorText}</span>*/}
		</span>);
	}
}
