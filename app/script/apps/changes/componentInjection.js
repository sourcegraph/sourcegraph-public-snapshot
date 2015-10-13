var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");

var CodeReview = require("./components/CodeReview");

document.addEventListener("DOMContentLoaded", () => {
	var el = $("#CodeReviewView");
	if (el.length > 0) {
		ReactDOM.render(
			<CodeReview data={window.preloadedReviewData||null} />,
			el[0]
		);
	}
});
