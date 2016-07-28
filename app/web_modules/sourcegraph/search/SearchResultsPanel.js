// @flow

import * as React from "react";
import {Panel} from "sourcegraph/components";
import SearchSettings from "sourcegraph/search/SearchSettings";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import CSSModules from "react-css-modules";
import styles from "./styles/SearchResultsPanel.css";


class SearchResultsPanel extends React.Component {
	static propTypes = {
		repo: React.PropTypes.string,
		location: React.PropTypes.object.isRequired,
		query: React.PropTypes.string.isRequired,
	};

	render() {
		const {repo, location, query} = this.props;
		return (
			<Panel hoverLevel="low" styleName="search-panel">
				<SearchSettings styleName="search-settings" innerClassName={styles["search-settings-inner"]} location={location} repo={repo} />
				{query && <GlobalSearch styleName="search-results" query={query} repo={repo} location={location} resultClassName={styles["search-result"]} />}
			</Panel>
		);
	}
}

export default CSSModules(SearchResultsPanel, styles);
