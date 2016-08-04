// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/avatar.css";

const PLACEHOLDER_IMAGE = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";

function Avatar({className, size, img}: {className?: any, size: any, img: any}) {
	return (
		<img className={className || ""} styleName={size || "small"} src={img || PLACEHOLDER_IMAGE} />
	);
}
(Avatar as any).propTypes = {
	img: React.PropTypes.string,
	size: React.PropTypes.string,
	className: React.PropTypes.string,
};

export default CSSModules(Avatar, styles);
