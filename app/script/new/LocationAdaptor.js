import React from "react";

const pushstateEvent = "LocationAdaptor.pushstate";

class LocationAdaptor extends React.Component {
	componentDidMount() {
		window.addEventListener("popstate", this._locationChanged.bind(this));
		window.addEventListener(pushstateEvent, this._locationChanged.bind(this));
	}

	componentWillUnmount() {
		window.removeEventListener("popstate", this._locationChanged.bind(this));
		window.removeEventListener(pushstateEvent, this._locationChanged.bind(this));
	}

	_locationChanged() {
		this.forceUpdate(); // this is necessary because the component uses external state (window.location)
	}

	render() {
		return (
			<this.props.component location={window.location.href} navigate={(uri) => {
				window.history.pushState(null, "", uri);
				let event = new CustomEvent(pushstateEvent);
				window.dispatchEvent(event);
			}} />
		);
	}
}

export default LocationAdaptor;
