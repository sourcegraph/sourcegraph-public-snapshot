import { CodeExcerpt } from "app/components/CodeExcerpt";
import { triggerReferences } from "app/references";
import { locKey, ReferencesState, store } from "app/references/store";
import * as url from "app/util/url";
import * as colors from "app/util/colors";
import { normalFontColor } from "app/util/colors";
import { Reference } from "app/util/types";
import * as csstips from "csstips";
import * as _ from "lodash";
import * as React from "react";
import { classes, style } from "typestyle";
import * as URI from "urijs";
import * as GlobeIcon from "react-icons/lib/md/language";
import * as RepoIcon from "react-icons/lib/go/repo";

namespace Styles {
	const border = `1px solid ${colors.borderColor}`;

	export const icon = style({ fontSize: "18px", marginLeft: "15px" });

	export const titleBar = style(csstips.horizontal, csstips.center, { backgroundColor: colors.referencesBackgroundColor, borderBottom: border, padding: "10px", fontSize: "14px", height: "32px", width: "100vw", position: "sticky", top: "0px" });
	export const titleBarTitle = style(csstips.content, { maxWidth: "calc(50vw)", marginRight: "25px" });
	export const titleBarGroup = style(csstips.content, {
		textTransform: "uppercase",
		letterSpacing: "1px",
		textDecoration: "none",
		$nest: {
			"&:hover": { color: "white" },
		},
	});
	export const titleBarGroupActive = classes(style({ fontWeight: "bold !important", color: "white !important" } as any), titleBarGroup);

	export const badge = style(csstips.content, { backgroundColor: "#233043 !important", borderRadius: "20px", color: normalFontColor, marginLeft: "10px", marginRight: "25px", fontSize: "11px", padding: "3px 6px", fontFamily: "system" });

	export const uriPathPart = style({ paddingLeft: "25px", paddingRight: "15px" });
	export const pathPart = style({});
	export const filePathPart = style({ color: "white", fontWeight: "bold", paddingRight: "15px" });

	export const refsGroup = style({ fontSize: "12px", fontFamily: "system", color: normalFontColor });
	export const refsGroupTitle = style(csstips.horizontal, csstips.center, { backgroundColor: "#233043", height: "32px" });
	export const refsList = style({ backgroundColor: colors.referencesBackgroundColor });
	export const ref = style({ fontFamily: "Menlo", borderBottom: border, padding: "10px" });
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
			const pageVars = (window as any).pageVars;
			if (!pageVars || !pageVars.ResolvedRev) {
				throw new TypeError("expected window.pageVars to exist, but it does not");
			}
			const rev = pageVars.ResolvedRev;
			const u = url.parseBlob();
			const coords = window.location.hash.split("$references")[0].split("#L")[1].split(":");
			triggerReferences({
				loc: {
					uri: u.uri!,
					rev: rev,
					path: u.path!,
					line: parseInt(coords[0], 10),
					char: parseInt(coords[1], 10),
				},
				word: "", // TODO: derive the correct word from somewhere
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
				<a className={this.state.group === "all" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={url.toBlob({ ...this.state.context.loc, modalMode: "", modal: "references" })}>
					All References
				</a>
				<div className={Styles.badge}>{localRefs.length + externalRefs.length}</div>
				<a className={this.state.group === "local" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={url.toBlob({ ...this.state.context.loc, modalMode: "local", modal: "references" })}>
					Local
				</a>
				<div className={Styles.badge}>{localRefs.length}</div>
				<a className={this.state.group === "external" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={url.toBlob({ ...this.state.context.loc, modalMode: "external", modal: "references" })}>
					Global
				</a>
				<div className={Styles.badge}>{externalRefs.length}</div>
				<div className={style(csstips.flex)} />
				<div style={{ float: "right", marginRight: "20px" }} onClick={() => this.props.onDismiss()}>X</div>
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

function getRefURL(ref: Reference): string {
	const uri = URI.parse(ref.uri);
	return `http://localhost:3080/${uri.hostname}/${uri.path}@${uri.query}/-/blob/${uri.fragment}#L${ref.range.start.line + 1}`;
}

export class ReferencesGroup extends React.Component<{ uri: string, path: string, refs: Reference[], isLocal: boolean }, {}> {
	render(): JSX.Element | null {
		const uriSplit = this.props.uri.split("/");
		const uriStr = uriSplit.length > 1 ? uriSplit.slice(1).join("/") : this.props.uri;
		const pathSplit = this.props.path.split("/");
		const filePart = pathSplit.pop();
		return <div className={Styles.refsGroup}>
			<div className={Styles.refsGroupTitle}>
				{this.props.isLocal ? <RepoIcon className={Styles.icon} /> : <GlobeIcon className={Styles.icon} />}
				<div className={Styles.uriPathPart}>{uriStr}</div>
				<div className={Styles.pathPart}>{pathSplit.join("/")}{pathSplit.length > 0 ? "/" : ""}</div>
				<div className={Styles.filePathPart}>{filePart}</div>
			</div>
			<div className={Styles.refsList}>
				{
					this.props.refs.sort((a, b) => {
						if (a.range.start.line < b.range.start.line) { return -1; }
						if (a.range.start.line === b.range.start.line) {
							if (a.range.start.character < b.range.start.character) {
								return -1;
							}
							if (a.range.start.character === b.range.start.character) {
								return 0;
							}
							return 1;
						}
						return 1;
					}).map((ref, i) => {
						const uri = URI.parse(ref.uri);
						return <div key={i} className={Styles.ref}>
							<div onClick={() => window.location.href = getRefURL(ref)}>
								<CodeExcerpt uri={uri.hostname + uri.path} rev={uri.query} path={uri.fragment} line={ref.range.start.line} char={ref.range.start.character} highlightLength={ref.range.end.character - ref.range.start.character} previewWindowExtraLines={1} />
							</div>
						</div>;
					})
				}
			</div>
		</div>;
	}
}
