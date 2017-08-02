import { CodeExcerpt } from "app/components/CodeExcerpt";
import { triggerReferences } from "app/references";
import { locKey, ReferencesState, refsFetchKey, store } from "app/references/store";
import { events } from "app/tracking/events";
import { normalFontColor, white } from "app/util/colors";
import * as colors from "app/util/colors";
import { pageVars } from "app/util/pageVars";
import { Reference } from "app/util/types";
import * as url from "app/util/url";
import * as csstips from "csstips";
import * as _ from "lodash";
import * as React from "react";
import * as RepoIcon from "react-icons/lib/go/repo";
import * as CloseIcon from "react-icons/lib/md/close";
import * as GlobeIcon from "react-icons/lib/md/language";
import { classes, style } from "typestyle";
import * as URI from "urijs";

namespace Styles {
	const border = `1px solid ${colors.borderColor}`;

	export const icon = style({ fontSize: "18px", marginLeft: "15px" });

	export const titleBar = style(csstips.horizontal, csstips.center, { backgroundColor: colors.referencesBackgroundColor, borderBottom: border, padding: "0px 16px", fontSize: "14px", height: "32px", width: "100vw", position: "sticky", top: "0px" });
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

	export const emptyState = style({ padding: "10px 16px", fontFamily: "system", fontSize: "14px" });

	export const uriPathPart = style({ paddingLeft: "25px", paddingRight: "15px" });
	export const pathPart = style({});
	export const filePathPart = style({ color: "white", fontWeight: "bold", paddingRight: "15px" });

	export const refsGroup = style({ fontSize: "12px", fontFamily: "system", color: normalFontColor });
	export const closeIcon = style({ cursor: "pointer", fontSize: "18px", color: colors.normalFontColor, $nest: { "&:hover": { color: white } } });
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
		let u = url.parseBlob();
		const onRefs = Boolean(u.path && u.modal && u.modal === "references");
		this.state = { ...store.getValue(), group: this.getRefsGroupFromUrl(window.location.href), docked: onRefs };
		if (onRefs) {
			const rev = pageVars.ResolvedRev;
			triggerReferences({
				loc: {
					uri: u.uri!,
					rev: rev,
					path: u.path!,
					line: u.line!,
					char: u.char!,
				},
				word: "", // TODO: derive the correct word from somewhere
			});
		}
		this.hashWatcher = window.addEventListener("hashchange", (e) => {
			u = url.parseBlob(e!.newURL!);
			const shouldShow = Boolean(u.path && u.modal && u.modal === "references");
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

	isLoading(group: "all" | "local" | "external"): boolean {
		if (!this.state.context) {
			return false;
		}

		const state = store.getValue();
		const loadingRefs = state.fetches.get(refsFetchKey(this.state.context.loc, true)) === "pending";
		const loadingXRefs = state.fetches.get(refsFetchKey(this.state.context.loc, false)) === "pending";

		switch (group) {
			case "all":
				return loadingRefs || loadingXRefs;
			case "local":
				return loadingRefs;
			case "external":
				return loadingXRefs;
		}

		return false;
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

		const isEmptyGroup = () => {
			switch (this.state.group) {
				case "all":
					return localRefs.length === 0 && externalRefs.length === 0;
				case "local":
					return localRefs.length === 0;
				case "external":
					return externalRefs.length === 0;
			}
			return false;
		};

		return <div>
			<div className={Styles.titleBar}>
				<div className={Styles.titleBarTitle}>
					{this.state.context.word}
				</div>
				<a className={this.state.group === "all" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={url.toBlob({ ...this.state.context.loc, modalMode: "", modal: "references" })} onClick={() => events.ShowAllRefsButtonClicked.log()}>
					All References
				</a>
				<div className={Styles.badge}>{localRefs.length + externalRefs.length}</div>
				<a className={this.state.group === "local" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={url.toBlob({ ...this.state.context.loc, modalMode: "local", modal: "references" })} onClick={() => events.ShowLocalRefsButtonClicked.log()}>
					Local
				</a>
				<div className={Styles.badge}>{localRefs.length}</div>
				<a className={this.state.group === "external" ? Styles.titleBarGroupActive : Styles.titleBarGroup} href={url.toBlob({ ...this.state.context.loc, modalMode: "external", modal: "references" })} onClick={() => events.ShowExternalRefsButtonClicked.log()}>
					Global
				</a>
				<div className={Styles.badge}>{externalRefs.length}</div>
				<div className={style(csstips.flex)} />
				<CloseIcon className={Styles.closeIcon} onClick={() => this.props.onDismiss()} />
			</div>
			{
				isEmptyGroup() && <div className={Styles.emptyState}>
					{this.isLoading(this.state.group) ? "Working..." : "No results"}
				</div>
			}
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
	return `/${uri.hostname}/${uri.path}@${uri.query}/-/blob/${uri.fragment}#L${ref.range.start.line + 1}`;
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
							<div onClick={() => {
								if (this.props.isLocal) {
									events.GoToLocalRefClicked.log();
								} else {
									events.GoToExternalRefClicked.log();
								}
								window.location.href = getRefURL(ref);
							}}>
								<CodeExcerpt uri={uri.hostname + uri.path} rev={uri.query} path={uri.fragment} line={ref.range.start.line} char={ref.range.start.character} highlightLength={ref.range.end.character - ref.range.start.character} previewWindowExtraLines={1} />
							</div>
						</div>;
					})
				}
			</div>
		</div>;
	}
}
