// @flow

import React from "react";

class Repository extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		width: React.PropTypes.number, // appended by "px"
	};

	static defaultProps = {
		width: 16,
	};

	render() {
		return <svg xmlns="http://www.w3.org/2000/svg" className={this.props.className} width={`${this.props.width}px`} viewBox="0 0 24 24" xmlns:xlink="http://www.w3.org/1999/xlink"><defs><path id="a" d="M0 3c0-1.7 1.3-3 3-3h18c1.7 0 3 1.3 3 3v18c0 1.7-1.3 3-3 3H3c-1.7 0-3-1.3-3-3V3z"/></defs><g fill="none" fill-rule="evenodd"><path stroke="#D5E5F2" stroke-opacity=".6" d="M12.5 4.5v78.2" stroke-linecap="square"/><mask id="b" fill="#fff"><use xlink:href="#a"/></mask><use fill="#D5E5F2" xlink:href="#a"/><path fill="#7793AE" fill-opacity=".5" d="M-5 7h33v1H-5zm0 9h33v1H-5z" mask="url(#b)"/><ellipse cx="4" cy="4" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="4" cy="12" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="4" cy="20" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="7" cy="4" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="7" cy="12" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="7" cy="20" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/></g></svg>;
	}
}

export default Repository;
