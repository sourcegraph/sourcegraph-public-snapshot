import React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/icon.css";

export const FileIcon = iconWrapper(require("react-icons/lib/go/file-text"));
export const FolderIcon = iconWrapper(require("react-icons/lib/go/file-directory"));
export const GitHubIcon = iconWrapper(require("react-icons/lib/go/mark-github"));
export const TriangleUpIcon = iconWrapper(require("react-icons/lib/go/triangle-up"));
export const TriangleDownIcon = iconWrapper(require("react-icons/lib/go/triangle-down"));
export const TriangleLeftIcon = iconWrapper(require("react-icons/lib/go/triangle-left"));
export const TriangleRightIcon = iconWrapper(require("react-icons/lib/go/triangle-right"));
export const PencilIcon = iconWrapper(require("react-icons/lib/go/pencil"));
export const CheckIcon = iconWrapper(require("react-icons/lib/go/check"));

// iconWrapper lets you pass a style directly to any of the exported components, e.g.
// <RepoIcon styleName="foo" />
function iconWrapper(Component) {
	const C = ({className, title}) => <div className={className} title={title} styleName="icon"><Component /></div>; // eslint-disable-line react/jsx-key
	C.propTypes = {
		className: React.PropTypes.string,
		title: React.PropTypes.string,
	};
	return CSSModules(C, styles);
}
