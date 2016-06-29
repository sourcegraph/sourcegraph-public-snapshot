import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalSearchMain.css";
import {queryFromStateOrURL} from "sourcegraph/search/routes";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import SearchSettings from "sourcegraph/search/SearchSettings";

function GlobalSearchMain({location}: {location: Location}) {
	const q = queryFromStateOrURL(location) || "";
	return (
		<div>
			<SearchSettings styleName="settings-inner" location={location} showAlerts={true} />
			<GlobalSearch query={q} location={location} className={styles["results"]} resultClassName={styles["result"]} />
		</div>
	);
}

export default CSSModules(GlobalSearchMain, styles);
