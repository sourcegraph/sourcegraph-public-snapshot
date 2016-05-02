import React from "react";

import CSSModules from "react-css-modules";
import styles from "./App.css";

@CSSModules(styles)
export default class TextSearchResult extends React.Component {
	static propTypes = {
		repo: React.PropTypes.string,
		file: React.PropTypes.string,
		match: React.PropTypes.string,
		startLine: React.PropTypes.number,
		endLine: React.PropTypes.number,
		query: React.PropTypes.string
	};

	constructor(props) {
		super(props);
		this.state = {
			match: window.atob(this.props.match).split('\n'),
		};
	}

	render() {
		return (
		<div className="code-list-item">
			<span className='language'>{this.props.file.split('.')[1]}</span>
			<div>
				<p styleName="text-result-title" className="repo-name">
					<a href={`https://sourcegraph.com/${this.props.repo}@master/-/blob/${this.props.file}?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext`}>{this.props.file}</a>
				</p>
			</div>
			<div styleName="file-box" className="file-box blob-wrapper">
				<table>
					<tbody>
						{this.state.match.map((item, i) =>
						<tr key={i}>
							<td className="blob-num">
								<a href={`https://sourcegraph.com/${this.props.repo}@master/-/blob/${this.props.file}?utm_source=chromeext&utm_medium=chromeext&utm_campaign=chromeext#L${this.props.startLine + i}`}>{this.props.startLine + i}</a>
							</td>
							<td className="blob-code">
								{item}
							</td>
						</tr>
						)}
					</tbody>
				</table>
			</div>
		</div>

		);
	}
}
