import React from "react";
import Container from "sourcegraph/Container";
import "sourcegraph/build/BuildBackend";
import BuildStore from "sourcegraph/build/BuildStore";
import BuildIndicator from "sourcegraph/build/BuildIndicator";

// BuildIndicatorContainer is for standalone BuildIndicators that need to
// be able to respond to changes in BuildStore.builds independently.
class BuildIndicatorContainer extends Container {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.builds = BuildStore.builds;
	}

	stores() { return [BuildStore]; }

	render() {
		let childProps = this.props;
		delete childProps.buildStore;
		return <BuildIndicator builds={BuildStore.builds} {...childProps} />;
	}
}

BuildIndicatorContainer.propTypes = {
	// All of the same properties as BuildIndicator, minus builds.
};

export default BuildIndicatorContainer;
