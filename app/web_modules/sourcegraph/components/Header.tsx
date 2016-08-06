// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/header.css";
import Loader from "./Loader";

class Header extends React.Component<any, any> {
	static propTypes = {
		title: React.PropTypes.string.isRequired,
		subtitle: React.PropTypes.string,
		loading: React.PropTypes.bool,
	};

	render(): JSX.Element | null {
		return (
			<div styleName="container">
				<div styleName="cloning_title">{this.props.title}</div>
				<div styleName="cloning_subtext">{this.props.loading ? <Loader /> : this.props.subtitle}</div>
			</div>
		);
	}
}

export default CSSModules(Header, styles);
