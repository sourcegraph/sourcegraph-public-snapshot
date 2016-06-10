import React from "react";
import DiffStatScale from "sourcegraph/delta/DiffStatScale";
import {isDevNull} from "sourcegraph/delta/util";
import styles from "sourcegraph/delta/styles/DiffFileList.css";
import CSSModules from "react-css-modules";
import {TriangleDownIcon} from "sourcegraph/components/Icons";
import {Panel, Menu, Popover} from "sourcegraph/components";

class DiffFileList extends React.Component {
	static propTypes = {
		files: React.PropTypes.arrayOf(React.PropTypes.object),
		stats: React.PropTypes.object.isRequired,
	};

	state = {closed: false};

	render() {
		return (
			<Panel styleName="container">
				<div styleName="header">
				<Popover popoverClassName={styles.popover}>
						<div styleName="label">
							<DiffStatScale Stat={this.props.stats} />
							<span styleName="count">{this.props.files.length}</span> changed files
							<span styleName="overall-stats">
								<span styleName="stat added">+{this.props.stats.Added}</span>
								<span styleName="stat deleted">&ndash;{this.props.stats.Deleted}</span>
							</span>
							&nbsp;<TriangleDownIcon />
						</div>
						<Menu className={styles["popover-menu"]}>
							{this.props.files.map((fd, i) => (
								<div key={fd.OrigName + fd.NewName} styleName="file-item">
									<a href={`#F${i}`} styleName="file">
										{isDevNull(fd.OrigName) ? <code styleName="change-type added">+</code> : null}
										{isDevNull(fd.NewName) ? <code styleName="change-type deleted">&ndash;</code> : null}
										{!isDevNull(fd.OrigName) && !isDevNull(fd.NewName) ? <code styleName="change-type changed">&bull;</code> : null}
										{!isDevNull(fd.OrigName) && !isDevNull(fd.NewName) && fd.OrigName !== fd.NewName ? (
											<span>{fd.OrigName} &rarr;&nbsp;</span>
										) : null}
										{isDevNull(fd.NewName) ? fd.OrigName : fd.NewName}
									</a>
									<span styleName="file-stats">
										<span styleName="stat added">+{fd.Stats.Added}</span>
										<span styleName="stat deleted">&ndash;{fd.Stats.Deleted}</span>
									</span>
								</div>
							))}
						</Menu>
					</Popover>
				</div>
			</Panel>
		);
	}
}

export default CSSModules(DiffFileList, styles, {allowMultiple: true});
