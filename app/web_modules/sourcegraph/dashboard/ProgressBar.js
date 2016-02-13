import React from "react";

import Component from "sourcegraph/Component";

class ProgressBar extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		const dots = Array.apply(null, Array(this.state.numSteps)).map((_, i) =>
			<div className="dot-wrapper" key={i}>
				<div className={`${i === this.state.currentStep - 1 ? "animated rubberBand" : ""} dot dot-${i} ${i >= this.state.currentStep ? "dot-incomplete" : ""}`}>
					{i > this.state.currentStep ? <div className={`dot-animation-target smallest-dot dot-${i} dot-${i}-inner`} /> : null}
					{i === this.state.currentStep ? <div className={`dot-animation-target smaller-dot dot-${i} dot-${i}-inner`} /> : null}
				</div>
				{i === this.state.currentStep ? <div className="animated fadeInLeft dot-carat"></div> : null}
			</div>
		);
		return (
			<div className="onboarding-progress-bar">
				<div className="dot-container">
					{dots}
				</div>
				<div className={`bar bar-${this.state.currentStep}`}></div>
			</div>
		);
	}
}

ProgressBar.propTypes = {
	numSteps: React.PropTypes.number.isRequired,
	currentStep: React.PropTypes.number.isRequired,
};

export default ProgressBar;
