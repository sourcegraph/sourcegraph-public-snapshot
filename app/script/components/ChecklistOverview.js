var React = require("react");

var ChecklistOverview = React.createClass({
	render() {
		var numTodo = this.props.checklist.Todo;
		var numDone = this.props.checklist.Done;
		var total = numTodo + numDone;
		return (
			<div className="pull-checklist-overview">
				<i className="fa fa-check-square-o"/> {numDone} of {total} tasks completed
			</div>
		);
	},
});

module.exports = ChecklistOverview;
