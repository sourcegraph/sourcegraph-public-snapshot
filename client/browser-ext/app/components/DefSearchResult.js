import React from "react";

import CSSModules from "react-css-modules";
import styles from "./App.css";
import EventLogger from "../analytics/EventLogger"

@CSSModules(styles)
export default class DefSearchResult extends React.Component {
	static propTypes = {
		href: React.PropTypes.string,
		qualifiedNameAndType: React.PropTypes.array,
		query: React.PropTypes.string,
	};

	render() {
		return (
			<tr styleName="def-result-row" className="js-navigation-item tree-browser-result sg-search-result">
				<td styleName="octicon-chevron" className="icon">
					<svg aria-hidden='true' styleName='octicon-chevron-right' className='octicon octicon-chevron-right' height='16' role='img' version='1.1' viewBox='0 0 8 16' width='8'>
						<path d='M7.5 8L2.5 13l-1.5-1.5 3.75-3.5L1 4.5l1.5-1.5 5 5z'></path>
					</svg>
				</td>
				<td styleName='page-icon' className="icon">
					<svg aria-hidden='true' className='octicon octicon-file-text' height='16' role='img' version='1.1' viewBox='0 0 12 16' width='12'>
						<path d='M6 5H2v-1h4v1zM2 8h7v-1H2v1z m0 2h7v-1H2v1z m0 2h7v-1H2v1z m10-7.5v9.5c0 0.55-0.45 1-1 1H1c-0.55 0-1-0.45-1-1V2c0-0.55 0.45-1 1-1h7.5l3.5 3.5z m-1 0.5L8 2H1v12h10V5z'></path>
					</svg>
				</td>
				<td>
					<a onClick={() => EventLogger.logEvent("GitHubSearchItemSelected", {query: this.props.query})} href={this.props.href + "?utm_source=browser-ext&browser_type=chrome"}>{this.props.qualifiedNameAndType}</a>
				</td>
			</tr>
		);
	}
}
