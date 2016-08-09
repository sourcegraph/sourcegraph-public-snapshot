// tslint:disable

import * as React from "react";
import {Panel} from "sourcegraph/components/index";
import {SearchSettings} from "sourcegraph/search/SearchSettings";
import {GlobalSearch} from "sourcegraph/search/GlobalSearch";
import * as styles from "./styles/SearchResultsPanel.css";


type Props = {
	repo: string | null,
	location: any,
	query: string,
};

export class SearchResultsPanel extends React.Component<Props, any> {
	render(): JSX.Element | null {
		const {repo, location, query} = this.props;
		return (
			<Panel hoverLevel="low" className={styles.search_panel}>
				<SearchSettings className={styles.search_settings} innerClassName={styles.search_settings_inner} location={location} repo={repo} />
				{query && <GlobalSearch className={styles.search_results} query={query} repo={repo} location={location} resultClassName={styles.search_result} />}
			</Panel>
		);
	}
}
