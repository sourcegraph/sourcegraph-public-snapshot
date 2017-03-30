import { ApolloError } from "apollo-client";
import * as autobind from "autobind-decorator";
import * as React from "react";
import { gql, graphql } from "react-apollo";
import { IRange } from "vs/editor/common/editorCommon";

import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { RouteProps, Router, pathFromRouteParams, repoRevFromRouteParams } from "sourcegraph/app/router";
import "sourcegraph/blob/styles/Monaco.css";
import { Heading, PageTitle } from "sourcegraph/components";
import { ChromeExtensionToast } from "sourcegraph/components/ChromeExtensionToast";
import { whitespace } from "sourcegraph/components/utils";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { repoPath, repoRev } from "sourcegraph/repo";
import { CloningRefresher, RepoMain } from "sourcegraph/repo/RepoMain";
import { Paywall, TrialEndingWarning, needsPayment } from "sourcegraph/user/Paywall";
import { InfoPanelLifecycle } from "sourcegraph/workbench/info/sidebar";
import { WorkbenchShell } from "sourcegraph/workbench/shell";

interface Props extends RouteProps {
	id: string | null; // symbol id, if defined
	mode: string | null; // mode for symbol, if defined
	repo: string | null;
	rev: string | null;
	isSymbolUrl: boolean;
	selection: IRange | null;
	loading?: boolean;
	refetch?: () => void;
	root?: GQL.IRoot;
	error?: ApolloError;
}

// WorkbenchComponent loads the VSCode workbench shell. To learn about VSCode and the
// way we use it, read README.vscode.md.
@autobind
class WorkbenchComponent extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	componentWillMount(): void {
		document.title = "Sourcegraph";
	}

	shouldComponentUpdate(nextProps: Props): boolean {
		return !nextProps.loading;
	}

	render(): JSX.Element | null {
		let repository: GQL.IRepository;
		let selection: IRange | null;
		let path: string;
		if (this.props.error) {
			switch (this.props.error.graphQLErrors[0].message) {
				case "revision not found":
					return <Error
						code="404"
						desc="Revision not found" />;
			}
		}
		if (this.props.loading || !this.props.root) {
			return null; // data not yet fetched
		}
		if (this.props.isSymbolUrl) {
			if (!this.props.root.symbols || this.props.root.symbols.length === 0) {
				return <Error
					code="404"
					desc="Symbol not found" />;
			}
			// Assume that there is only one symbol for now
			const symbol = this.props.root.symbols[0];
			repository = symbol.repository;
			path = symbol.path;
			selection = RangeOrPosition.fromLSPPosition(symbol).toMonacoRangeAllowEmpty();
		} else {
			if (!this.props.root.repository) {
				return <Error
					code="404"
					desc="Repository not found" />;
			}
			repository = this.props.root.repository;
			selection = this.props.selection;
			path = pathFromRouteParams(this.props.params);
		}

		if (repository.revState.cloneInProgress) {
			return <CloningRefresher refetch={this.props.refetch!} />;
		}

		if (needsPayment(repository.expirationDate)) {
			return <Paywall repo={repository.uri} />;
		}

		const commitID = repository.revState.zapRev ? repository.revState.zapRev.base : repository.revState.commit!.sha1;
		return <div style={{ display: "flex", height: "100%" }}>
			<BlobPageTitle repo={this.props.repo} path={path} />
			{/* TODO(john): repo main takes the commit state for the 'y' hotkey, assume for now revState can be passed. */}
			<RepoMain {...this.props} commit={repository.revState}>
				<ChromeExtensionToast location={this.props.location} layout={() => this.forceUpdate()} />
				<TrialEndingWarning layout={() => this.forceUpdate()} expirationDate={repository.expirationDate} />
				<WorkbenchShell
					location={this.props.location}
					routes={this.props.routes}
					params={this.props.params}
					repo={repository.uri}
					rev={this.props.rev}
					commitID={commitID}
					zapRev={repository.revState.zapRev && this.props.rev ? this.props.rev : undefined}
					zapRef={repository.revState.zapRev ? repository.revState.zapRev.ref : undefined}
					branch={repository.revState.zapRev ? repository.revState.zapRev.branch : undefined}
					path={path}
					selection={selection} />
				<InfoPanelLifecycle fileEventProps={{ repo: repository.uri, rev: commitID, path: path }} />
			</RepoMain>
		</div>;
	}
}

function Error({ code, desc }: { code: string, desc: string }): JSX.Element {
	return <div style={{ textAlign: "center", marginTop: whitespace[8] }}>
		<Heading level={2}>{code}</Heading>
		<Heading level={4} color="gray">{desc}</Heading>
	</div>;
}

function BlobPageTitle({ repo, path }: { repo: string | null, path: string }): JSX.Element {
	const base = path.split("/").pop() || path;
	if (!repo) {
		return <PageTitle title={base} />;
	}
	repo = repo.replace(/^github.com\//, "");
	let title = base === "/" ? repo : `${base} · ${repo}`;
	return <PageTitle title={title} />;
}

function mapRouteProps(props: RouteProps): Props {
	let rangeOrPosition: RangeOrPosition = RangeOrPosition.fromOneIndexed(1);
	if (props.location && props.location.hash && props.location.hash.startsWith("#L")) {
		rangeOrPosition = RangeOrPosition.parse(props.location.hash.replace(/^#L/, "")) || rangeOrPosition;
	}
	const isSymbolUrl = getRoutePattern(props.routes) === abs.goSymbol;
	let id: string | null = null;
	let mode: string | null = null;
	let repo: string | null = null;
	let rev: string | null = null;
	if (isSymbolUrl) {
		id = props.params.splat as string;
		mode = "go";
	} else {
		const repoRevString = repoRevFromRouteParams(props.params);
		repo = repoPath(repoRevString);
		rev = repoRev(repoRevString);
	}

	return { ...props, id, mode, repo, rev, isSymbolUrl, selection: rangeOrPosition.toMonacoRangeAllowEmpty() };
}

// TODO(john): use saved graphql queries and fragments.
const repoFields = `
	uri
	description
	defaultBranch
	expirationDate
	revState(rev: $rev) {
		zapRev {
			ref
			base
			branch
		}
		commit {
			sha1
		}
		cloneInProgress
	}
`;

export const Workbench = graphql(gql`
	query Workbench($id: String, $repo: String!, $rev: String, $isSymbolUrl: Boolean) {
		root {
			symbols(id: $id, mode: $mode) @include(if: $isSymbolUrl) {
				path
				line
				character
				repository {
					${repoFields}
				}
			}
			repository(uri: $repo) @skip(if: $isSymbolUrl) {
				${repoFields}
			}
		}
	}`, {
		props: props => ({ ...mapRouteProps(props.ownProps), ...props.data }),
		options: props => {
			const mappedProps = mapRouteProps(props);
			return {
				variables: {
					id: mappedProps.id || "",
					mode: mappedProps.mode || "",
					repo: mappedProps.repo,
					rev: mappedProps.rev || "",
					isSymbolUrl: mappedProps.isSymbolUrl,
				}
			};
		},
	})(WorkbenchComponent);
