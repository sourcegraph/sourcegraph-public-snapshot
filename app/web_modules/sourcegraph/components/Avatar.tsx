// tslint:disable

import * as React from "react";
import * as styles from "./styles/avatar.css";

const PLACEHOLDER_IMAGE = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";

function Avatar({className, size, img}: {className?: any, size: any, img: any}) {
	return (
		<img className={`${className || ""} ${sizeClasses[size] || styles.small}`} src={img || PLACEHOLDER_IMAGE} />
	);
}
(Avatar as any).propTypes = {
	img: React.PropTypes.string,
	size: React.PropTypes.string,
	className: React.PropTypes.string,
};

const sizeClasses = {
	"tiny": styles.tiny,
	"small": styles.small,
	"medium": styles.medium,
	"large": styles.large,
};

export default Avatar;
