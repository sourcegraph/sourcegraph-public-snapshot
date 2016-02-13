import React from "react";

import Component from "sourcegraph/Component";

import Dispatcher from "sourcegraph/Dispatcher";
import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";

class ListMenu extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		return (
			<div>
				<p className="category-menu-label"><strong>{this.state.label}:</strong></p>
				<div>
					{this.state.categories.map((category, i) =>
						<div className="category-label"
							key={i}>
							<a className={category === this.state.current ? "category-active" : null}
								onClick={(e) => {
									e.preventDefault();
									Dispatcher.dispatch(new OnboardingActions.SelectCategory(category));
								}}>{category}</a>
						</div>
					)}
				</div>
			</div>
		);
	}
}

ListMenu.propTypes = {
	label: React.PropTypes.string.isRequired,
	categories: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
	current: React.PropTypes.string,
};

export default ListMenu;
