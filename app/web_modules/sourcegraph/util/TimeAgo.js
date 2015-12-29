import moment from "moment";
import React from "react";

import Component from "sourcegraph/Component";

class TimeAgo extends Component {
	render() {
		return <time title={moment(this.props.time).calendar()}>{moment(this.props.time).fromNow()}</time>;
	}
}
TimeAgo.propTypes = {
	time: React.PropTypes.string.isRequired,
};

export default TimeAgo;
