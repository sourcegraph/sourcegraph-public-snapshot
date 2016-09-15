// tslint:disable: typedef ordered-imports

import { Location } from "history";
import * as React from "react";
import Helmet from "react-helmet";
import { Blob } from "sourcegraph/blob/Blob";
import "sourcegraph/blob/BlobBackend";
import * as Style from "sourcegraph/blob/styles/Blob.css";
import { trimRepo } from "sourcegraph/repo";
import { httpStatusCode } from "sourcegraph/util/httpStatusCode";
import { Header } from "sourcegraph/components/Header";

interface Props {
	repo: string;
	rev: string | null;
	commitID?: string;
	path: string;
	blob?: any;
	startLine?: number;
	startCol?: number;
	endLine?: number;
	endCol?: number;
	location: Location;
}

type State = any;

export class BlobMain extends React.Component<Props, State> {
	componentDidMount(): void {
		document.body.style.overflowY = "hidden";
	}

	componentWillUnmount(): void {
		document.body.style.overflowY = "auto";
	}

	render(): JSX.Element | null {
		if (this.props.blob && this.props.blob.Error) {
			let msg;
			switch (this.props.blob.Error.response.status) {
				case 413:
					msg = "Sorry, this file is too large to display.";
					break;
				default:
					msg = "File is not available.";
			}
			return (
				<Header
					title={`${httpStatusCode(this.props.blob.Error)}`}
					subtitle={msg} />
			);
		}

		let title = trimRepo(this.props.repo);
		const pathParts = this.props.path ? this.props.path.split("/") : null;
		if (pathParts) {
			title = `${pathParts[pathParts.length - 1]} · ${title}`;
		}
		return (
			<div className={Style.container}>
				<Helmet title={title} />
				{this.props.blob && typeof this.props.blob.ContentsString === "string" && <Blob
					repo={this.props.repo}
					rev={this.props.rev}
					path={this.props.path}
					contents={this.props.blob.ContentsString}
					startLine={this.props.startLine}
					endLine={this.props.endLine} />}
			</div>
		);
	}
}
