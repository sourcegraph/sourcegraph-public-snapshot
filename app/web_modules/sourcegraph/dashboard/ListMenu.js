import React from "react";

import Component from "sourcegraph/Component";

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
									this.state.onMenuClick(category);
								}}>
								{category}
							</a>
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
	onMenuClick: React.PropTypes.func.isRequired,
};

export default ListMenu;
