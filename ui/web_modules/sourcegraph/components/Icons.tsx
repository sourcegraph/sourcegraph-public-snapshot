// tslint:disable: no-var-requires
import * as classNames from "classnames";
import * as React from "react";

import * as styles from "sourcegraph/components/styles/icon.css";

export const FileIcon = iconWrapper(require("react-icons/lib/go/file-text"));
export const FolderIcon = iconWrapper(require("react-icons/lib/go/file-directory"));
export const GitHubIcon = iconWrapper(require("react-icons/lib/go/mark-github"));
export const GoogleIcon = iconWrapper(require("react-icons/lib/fa/google"));
export const CloudDownloadIcon = iconWrapper(require("react-icons/lib/go/cloud-download"));
export const TriangleUpIcon = iconWrapper(require("react-icons/lib/go/triangle-up"));
export const TriangleDownIcon = iconWrapper(require("react-icons/lib/go/triangle-down"));
export const TriangleLeftIcon = iconWrapper(require("react-icons/lib/go/triangle-left"));
export const TriangleRightIcon = iconWrapper(require("react-icons/lib/go/triangle-right"));
export const PencilIcon = iconWrapper(require("react-icons/lib/go/pencil"));
export const CheckIcon = iconWrapper(require("react-icons/lib/go/check"));
export const GlobeIcon = iconWrapper(require("react-icons/lib/fa/globe"));
export const LanguageIcon = iconWrapper(require("react-icons/lib/fa/language"));
export const MagnifyingGlassIcon = iconWrapper(require("react-icons/lib/fa/search"));
export const CloseIcon = iconWrapper(require("react-icons/lib/fa/close"));
export const EllipsisHorizontal = iconWrapper(require("react-icons/lib/fa/ellipsis-v"));
export const FaAngleDown = iconWrapper(require("react-icons/lib/fa/angle-down"));
export const FaAngleRight = iconWrapper(require("react-icons/lib/fa/angle-right"));
export const FaChevronDown = iconWrapper(require("react-icons/lib/fa/chevron-down"));
export const PlayIcon = iconWrapper(require("react-icons/lib/fa/play-circle"));
export const ToolsIcon = iconWrapper(require("react-icons/lib/go/tools"));
export const FaThumbsUp = iconWrapper(require("react-icons/lib/fa/thumbs-up"));
export const FaThumbsDown = iconWrapper(require("react-icons/lib/fa/thumbs-down"));

// iconWrapper lets you pass a style directly to any of the exported components, e.g.
// <RepoIcon className={styles.foo} />
function iconWrapper<P>(Component: React.ComponentClass<P>): wrapper {
	return function({className, style, title}: Props): JSX.Element {
		return <div className={classNames(className, styles.icon)} style={style} title={title}><Component /></div>;
	};
}

type Props = {
	className: string,
	style: React.CSSProperties,
	title: string
};

type wrapper = (props: Props) => JSX.Element;
