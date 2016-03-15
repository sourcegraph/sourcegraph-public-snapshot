import React from "react";

const pushstateEvent = "LocationAdaptor.pushstate";

class LocationAdaptor extends React.Component {
	componentDidMount() {
		if (typeof window !== "undefined") {
			window.addEventListener("popstate", this._locationChanged.bind(this));
			window.addEventListener(pushstateEvent, this._locationChanged.bind(this));
		}
	}

	componentWillUnmount() {
		if (typeof window !== "undefined") {
			window.removeEventListener("popstate", this._locationChanged.bind(this));
			window.removeEventListener(pushstateEvent, this._locationChanged.bind(this));
		}
	}

	_locationChanged() {
		this.forceUpdate(); // this is necessary because the component uses external state (window.location)
	}

	render() {
		let other = Object.assign({}, this.props);
		delete other.component;

		// Don't use initially provided location prop in the browser, where it
		// should always reflect the current URL.
		if (typeof window !== "undefined") other.location = window.location.href;

		const navigate = typeof window === "undefined" ? null : (uri) => {
			window.history.pushState(null, "", uri);
			let event = new CustomEvent(pushstateEvent);
			window.dispatchEvent(event);
		};

		return (
				<this.props.component navigate={navigate} {...other} />
		);
	}
}

LocationAdaptor.propTypes = {
	location: React.PropTypes.string,
};

export default LocationAdaptor;
