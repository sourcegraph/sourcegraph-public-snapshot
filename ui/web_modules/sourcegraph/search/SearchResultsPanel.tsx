import * as React from "react";
import {Panel} from "sourcegraph/components";
import {GlobalSearch} from "sourcegraph/search/GlobalSearch";
import * as styles from "sourcegraph/search/styles/SearchResultsPanel.css";

interface Props {
	location: any;
	query: string;
};

export class SearchResultsPanel extends React.Component<Props, {}> {
	render(): JSX.Element | null {
		const {location, query} = this.props;
		return (
			<Panel hoverLevel="low" className={styles.search_panel}>
				{query && <GlobalSearch className={styles.search_results} query={query} location={location} resultClassName={styles.search_result} />}
			</Panel>
		);
	}
}
