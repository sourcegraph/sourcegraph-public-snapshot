import * as React from "react";
import * as CloseIcon from "react-icons/lib/md/close";
import * as colors from "sourcegraph/util/colors";
import { style } from "typestyle";

namespace Styles {
	export const closeIcon = style({ cursor: "pointer", fontSize: "18px", color: colors.normalFontColor, marginRight: 16, $nest: { "&:hover": { color: colors.white } } });
}

interface Props {
	onClick: () => void;
}

export class TreeHeader extends React.Component<Props, {}> {

	render(): JSX.Element | null {
		return <div style={{ marginTop: 15, marginBottom: 10, display: "flex", alignItems: "center", alignContent: "center" }} >
			<div style={{ paddingLeft: 16, display: "flex", alignItems: "center", justifyContent: "center", width: "100%", fontSize: 12 }}>
				<div className="js-selected-navigation-item header-navlink" style={{ paddingLeft: 5, color: "white", alignItems: "center", alignContent: "center", display: "inline-block", marginTop: 0, flex: 1 }}>FILES</div>
				<CloseIcon className={Styles.closeIcon} onClick={this.props.onClick} />
			</div>
		</div>;
	}
}
