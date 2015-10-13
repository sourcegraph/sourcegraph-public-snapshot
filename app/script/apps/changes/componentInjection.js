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
