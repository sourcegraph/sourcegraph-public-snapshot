var React = require("react");
var moment = require("moment");

var TimeAgo = React.createClass({
	render() {
		return <time title={moment(this.props.time).calendar()}>{moment(this.props.time).fromNow()}</time>;
	},
});

module.exports = TimeAgo;
