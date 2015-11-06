import React from "react";

export default class LocationAdaptor extends React.Component {
	componentDidMount() {
		window.addEventListener("popstate", this._locationChanged.bind(this));
	}

	componentWillUnmount() {
		window.removeEventListener("popstate", this._locationChanged.bind(this));
	}

	_locationChanged() {
		this.forceUpdate(); // this is necessary because the component uses external state (window.location)
	}

	render() {
		return (
			<this.props.component location={window.location.href} navigate={(uri) => {
				window.history.pushState(null, "", uri);
				this._locationChanged();
			}} />
		);
	}
}
