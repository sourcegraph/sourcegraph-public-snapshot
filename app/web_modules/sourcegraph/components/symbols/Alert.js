// @flow

import React from "react";

class Alert extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		height: React.PropTypes.number,
		width: React.PropTypes.number,
		style: React.PropTypes.object,
	};

	static defaultProps = {
		width: 16,
	};

	render() {
		let style = this.props.style;
		style.verticalAlign = "middle";
		return <svg className={this.props.className} width={`${this.props.width}px`} style={style} viewBox="2 76 16 14" xmlns="http://www.w3.org/2000/svg"><path d="M17.2 87.6l-6-10.5c-.2 0-.3-.2-.4-.3-.6-.6-1.6-.6-2.2 0l-.3.4-6 10.6-.2.4c-.2 1 .3 1.7 1.2 2h12.4c.7 0 1.3-.2 1.6-.8.3-.5.3-1 0-1.6zm-6.8 0c-.2.3-.4.4-.7.4-.3 0-.5 0-.8-.3-.3-.2-.4-.5-.4-.7 0-.3 0-.6.3-.8.2-.2.4-.3.7-.3.3 0 .5 0 .7.2.2.2.3.5.3.8 0 .2 0 .5-.3.7zm.3-6l-.2.8-.3 1.2-.3 1.7h-.5c0-.7-.2-1.3-.3-1.7 0-.5-.2-1-.3-1.2l-.3-.8V81c0-.3 0-.6.2-.8 0-.2.4-.3.7-.3.3 0 .5 0 .7.2.2.2.3.5.3.7v.5z" /></svg>;
	}
}

export default Alert;
