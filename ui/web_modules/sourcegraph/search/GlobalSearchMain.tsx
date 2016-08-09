// tslint:disable

import * as React from "react";
import * as styles from "./styles/GlobalSearchMain.css";
import {queryFromStateOrURL} from "sourcegraph/search/routes";
import {GlobalSearch} from "sourcegraph/search/GlobalSearch";
import {SearchSettings} from "sourcegraph/search/SearchSettings";

export function GlobalSearchMain({location}: {location: HistoryModule.Location}) {
	const q = queryFromStateOrURL(location) || "";
	return (
		<div>
			<SearchSettings className={styles.search_settings} innerClassName={styles.search_settings_inner} location={location} />
			<GlobalSearch repo={null} query={q} location={location} className={styles.results} resultClassName={styles.result} />
		</div>
	);
}
