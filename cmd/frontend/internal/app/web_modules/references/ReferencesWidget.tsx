import { CodeExcerpt } from "app/components/CodeExcerpt";
import { triggerReferences } from "app/references";
import { locKey, ReferencesState, store } from "app/references/store";
import { parseURL, urlToBlob } from "app/util";
import * as colors from "app/util/colors";
import { normalFontColor } from "app/util/colors";
import { Reference } from "app/util/types";
import * as csstips from "csstips";
import * as _ from "lodash";
import * as React from "react";
import { classes, style } from "typestyle";
import * as URI from "urijs";

namespace Styles {
	const border = `1px solid ${colors.borderColor}`;

	export const titleBar = style(csstips.horizontal, csstips.center, { borderBottom: border, padding: "10px", fontSize: "14px" });
	export const titleBarTitle = style(csstips.content, { maxWidth: "calc(50vw)", marginRight: "25px" });
	export const titleBarGroup = style(csstips.content, {
		textTransform: "uppercase",
		letterSpacing: "1px",
		textDecoration: "none",
		$nest: {
			"&:hover": { color: "white" },
		},
	});
	export const titleBarGroupActive = classes(style({ fontWeight: "bold !important", color: "white !important" }), titleBarGroup);

	export const badge = style(csstips.content, { backgroundColor: "#233043 !important", borderRadius: "20px", color: normalFontColor, marginLeft: "10px", marginRight: "25px", fontSize: "11px", padding: "3px 6px" });

	export const uriPathPart = style({ fontWeight: "bold", paddingRight: "15px" });
	export const pathPart = style({ paddingRight: "15px" });
	export const filePathPart = style({ color: "white", fontWeight: "bold", paddingRight: "15px" });

	export const refsGroup = style(csstips.horizontal, csstips.center, { borderBottom: border });
	export const refsList = style(csstips.horizontal, csstips.wrap, { borderLeft: border });
	export const ref = style(csstips.content, { borderRight: border, borderBottom: border, marginBottom: "-1px" /* prevent "double border" */, padding: "10px" });
}

interface Props {
	onDismiss: () => void;
}

interface State extends ReferencesState {
	docked: boolean;
	group: "all" | "local" | "external";
}

export class ReferencesWidget extends React.Component<Props, State> {
	subscription: any;
	hashWatcher: any;

	constructor(props: Props) {
		super(props);
		const onRefs = window.location.hash.indexOf("$references") !== -1;
		this.state = { ...store.getValue(), group: this.getRefsGroupFromUrl(window.location.href), docked: onRefs };
		if (onRefs) {
			const url = parseURL();
			const coords = window.location.hash.split("$references")[0].split("#L")[1].split(":");
			triggerReferences({
				loc: {
					uri: url.uri!,
					rev: url.rev!,
					path: url.path!,
					line: parseInt(coords[0], 10),
					char: parseInt(coords[1], 10),
				},
				word: "TBD",
			});
		}
		this.hashWatcher = window.addEventListener("hashchange", (e) => {
			const shouldShow = e!.newURL!.indexOf("$references") !== -1;
			if (shouldShow) {
				this.setState({ ...this.state, group: this.getRefsGroupFromUrl(e!.newURL!), docked: true });
			}
		});
	}

	getRefsGroupFromUrl(urlStr: string): "all" | "local" | "external" {
		if (urlStr.indexOf("$references:local") !== -1) {
			return "local";
		}
		if (urlStr.indexOf("$references:external") !== -1) {
			return "external";
		}
		return "all";
	}

	componentDidMount(): void {
		this.subscription = store.subscribe((state) => {
			this.setState({ ...state, group: this.state.group, docked: this.state.docked });
		});
	}

	componentWillUnmount(): void {
		if (this.subscription) {
			this.subscription.unsubscribe();
		}
		if (this.hashWatcher) {
			window.removeEventListener("hashchange", this.hashWatcher);
		}
	}

	render(): JSX.Element | null {
		if (!this.state.context) {
			return null;
		}
		const loc = locKey(this.state.context.loc);
		const refs = this.state.refsByLoc.get(loc);

		// References by fully qualified URI, like git://github.com/gorilla/mux?rev#mux.go
		const refsByUri = _.groupBy(refs, (ref) => ref.uri);

		const localPrefix = "git://" + this.state.context.loc.uri;
		const [localRefs, externalRefs] = _(refsByUri).keys().partition((uri) => uri.startsWith(localPrefix)).value();

		return <div>
			<div className={Styles.titleBar}>
				<div className={Styles.titleBarTitle}>
					{this.state.context.word}
				</div>
				<a className={this.state.group === "all" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={urlToBlob({ ...this.state.context.loc, refs: "all" })}>
					All References
				</a>
				<div className={Styles.badge}>{localRefs.length + externalRefs.length}</div>
				<a className={this.state.group === "local" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={urlToBlob({ ...this.state.context.loc, refs: "local" })}>
					Local
				</a>
				<div className={Styles.badge}>{localRefs.length}</div>
				<a className={this.state.group === "external" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={urlToBlob({ ...this.state.context.loc, refs: "external" })}>
					Global
				</a>
				<div className={Styles.badge}>{externalRefs.length}</div>
				<div className={style(csstips.flex)} />
				<div style={{ float: "right" }} onClick={() => this.props.onDismiss()}>X</div>
			</div>
			<div>
				{
					(this.state.group === "all" || this.state.group === "local") && localRefs.sort().map((uri, i) => {
						const parsed = URI.parse(uri);
						return <ReferencesGroup key={i} uri={parsed.hostname + parsed.path} path={parsed.fragment} isLocal={true} refs={refsByUri[uri]} />;
					})
				}
			</div>
			<div>
				{
					(this.state.group === "all" || this.state.group === "external") && externalRefs.map((uri, i) => { /* don't sort, to avoid jerky UI as new repo results come in */
						const parsed = URI.parse(uri);
						return <ReferencesGroup key={i} uri={parsed.hostname + parsed.path} path={parsed.fragment} isLocal={false} refs={refsByUri[uri]} />;
					})
				}
			</div>
		</div>;
	}
}

function getRangeString(ref: Reference): string {
	// return `${ref.range.start.line + 1}:${ref.range.start.character + 1}-${ref.range.end.line + 1}:${ref.range.end.character + 1}`;
	return `${ref.range.start.line + 1}:${ref.range.start.character + 1}`;
}

function getRefURL(ref: Reference): string {
	const uri = URI.parse(ref.uri);
	return `http://localhost:3080/${uri.hostname}/${uri.path}@${uri.query}/-/blob/${uri.fragment}#L${ref.range.start.line + 1}`;
}

class ReferencesGroup extends React.Component<{ uri: string, path: string, refs: Reference[], isLocal: boolean }, {}> {
	render(): JSX.Element | null {
		const uriSplit = this.props.uri.split("/");
		const uriStr = uriSplit.length > 1 ? uriSplit.slice(1).join("/") : this.props.uri;
		const pathSplit = this.props.path.split("/");
		const filePart = pathSplit.pop();
		return <div className={Styles.refsGroup}>
			<div className={Styles.uriPathPart}>{uriStr}</div>
			<div className={Styles.pathPart}>{pathSplit.join("/")}</div>
			<div className={Styles.filePathPart}>{filePart}</div>
			<div className={Styles.refsList}>
				{
					this.props.refs.sort((a, b) => {
						if (a.uri < b.uri) { return -1; }
						if (a.uri === b.uri) { return 0; }
						return 1;
					}).map((ref, i) => {
						const uri = URI.parse(ref.uri);
						return <div key={i} className={Styles.ref}>
							<a href={getRefURL(ref)}>
								<CodeExcerpt uri={uri.hostname + uri.path} rev={uri.query} path={uri.fragment} line={ref.range.start.line} />
							</a>
						</div>;
					})
				}
			</div>
		</div>;
	}
}
