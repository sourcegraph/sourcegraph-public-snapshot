// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/components/styles/avatar.css";
import * as classNames from "classnames";

const PLACEHOLDER_IMAGE = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";

interface Props {
	img?: string;
	size: string;
	className?: string;
}

export function Avatar({className, size, img}: Props) {
	return (
		<img className={classNames(className, sizeClasses[size] || styles.small)} src={img || PLACEHOLDER_IMAGE} />
	);
}

const sizeClasses = {
	"tiny": styles.tiny,
	"small": styles.small,
	"medium": styles.medium,
	"large": styles.large,
};
