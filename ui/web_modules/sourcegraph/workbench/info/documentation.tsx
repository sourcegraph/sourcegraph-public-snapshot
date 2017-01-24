import * as autobind from "autobind-decorator";
import * as React from "react";
import { Link } from "react-router";
import { marked } from "vs/base/common/marked/marked";
import URI from "vs/base/common/uri";

import { urlToBlobRange } from "sourcegraph/blob/routes";
import { Button, FlexContainer } from "sourcegraph/components";
import { ArrowRight, List } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { URIUtils } from "sourcegraph/core/uri";
import { DefinitionData } from "sourcegraph/util/RefsBackend";
import { RouterContext, prettifyRev } from "sourcegraph/workbench/utils";

interface Props {
	defData: DefinitionData;
}

interface State {
	showingFullDocString: boolean;
}

const DocStringLength = 200;

@autobind
export class DefinitionDocumentationHeader extends React.Component<Props, State> {

	constructor(props: Props) {
		super(props);
		this.state = {
			showingFullDocString: false,
		};
	}

	private onToggleExpand(e: React.MouseEvent<HTMLDivElement>): void {
		if ((e.target as any).dataset.toggle === undefined) {
			return;
		}
		this.setState({
			showingFullDocString: !this.state.showingFullDocString,
		});
	}

	render(): JSX.Element | null {
		const { defData } = this.props;
		const uri = URI.parse(defData.definition.uri);
		let { repo, rev, path } = URIUtils.repoParams(uri);
		rev = prettifyRev(rev);
		const url = urlToBlobRange(repo, rev, path, defData.definition.range);
		const fullDocString = marked(defData.docString, { sanitize: true });
		let renderedDocString = fullDocString;
		if (fullDocString.length >= DocStringLength) {
			if (this.state.showingFullDocString) {
				renderedDocString = renderedDocString + `<a data-toggle style="display: inline-block; padding-left: 5px;">   Hide...</a>`;
			} else {
				renderedDocString = renderedDocString.substr(0, DocStringLength);
				renderedDocString = renderedDocString + `<a data-toggle style="display: inline-block; padding-left: 5px;">   More...</a>`;
			}
		}
		return <RouterContext><div style={Object.assign({
			color: colors.text(),
			padding: whitespace[3],
			paddingTop: 0,
		}, typography.small)}>
			<div onClick={this.onToggleExpand} style={Object.assign({}, {
				maxHeight: "40vh",
				overflowY: "scroll",
				color: colors.blueGrayD1(),
			}, typography[2])} dangerouslySetInnerHTML={{ __html: renderedDocString }}>
			</div>
			<p style={{ color: colors.blueGray(), paddingTop: 0 }}>
				Defined in
				<Link to={`/${repo}`} style={{ paddingTop: whitespace[2], paddingBottom: whitespace[2] }}>
					<List width={20} style={{ marginLeft: 4 }} />
					{repo.replace(/^github.com\//, "")}
				</Link>
			</p>
			<FlexContainer content="stretch" justify="between" items="center">
				<Link style={{ flex: "1 0" }} to={url}>
					<Button color="blueGray" outline={true} style={{ width: "100%" }}>
						Jump to definition <ArrowRight width={22} style={{ top: 0 }} />
					</Button>
				</Link>
			</FlexContainer>
		</div ></RouterContext>;
	}
}
