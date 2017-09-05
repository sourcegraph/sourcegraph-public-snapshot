import * as _ from 'lodash';
import * as React from 'react';
import * as DownIcon from 'react-icons/lib/fa/angle-down';
import * as RightIcon from 'react-icons/lib/fa/angle-right';
import * as RepoIcon from 'react-icons/lib/go/repo';
import * as CloseIcon from 'react-icons/lib/md/close';
import * as GlobeIcon from 'react-icons/lib/md/language';
import { Link } from 'react-router-dom';
import { CodeExcerpt } from 'sourcegraph/components/CodeExcerpt';
import { triggerReferences } from 'sourcegraph/references';
import { locKey, ReferencesState, refsFetchKey, store } from 'sourcegraph/references/store';
import { events } from 'sourcegraph/tracking/events';
import { pageVars } from 'sourcegraph/util/pageVars';
import { Reference } from 'sourcegraph/util/types';
import * as url from 'sourcegraph/util/url';
import * as URI from 'urijs';

interface ReferenceGroupProps {
    uri: string;
    path: string;
    refs: Reference[];
    isLocal: boolean;
    hidden?: boolean;
}

interface ReferenceGroupState {
    hidden?: boolean;
}

export class ReferencesGroup extends React.Component<ReferenceGroupProps, ReferenceGroupState> {
    constructor(props: ReferenceGroupProps) {
        super(props);
        this.state = { hidden: props.hidden };
    }

    public render(): JSX.Element | null {
        const uriSplit = this.props.uri.split('/');
        const uriStr = uriSplit.length > 1 ? uriSplit.slice(1).join('/') : this.props.uri;
        const pathSplit = this.props.path.split('/');
        const filePart = pathSplit.pop();

        let refs: JSX.Element | null = null;
        if (!this.state.hidden) {
            refs = <div className='references-group__list'>
                {
                    this.props.refs
                        .sort((a, b) => {
                            if (a.range.start.line < b.range.start.line) {
                                return -1;
                            }
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
                        })
                        .map((ref, i) => {
                            const uri = URI.parse(ref.uri);
                            const href = getRefURL(ref, uri.query);
                            return (
                                <Link
                                    to={href}
                                    key={i}
                                    className='references-group__reference'
                                    onClick={e => {
                                        (this.props.isLocal ? events.GoToLocalRefClicked : events.GoToExternalRefClicked).log();
                                        url.openFromJS(href, e);
                                    }}
                                >
                                <CodeExcerpt
                                    uri={uri.hostname + uri.path} rev={uri.query}
                                    path={uri.fragment}
                                    line={ref.range.start.line}
                                    char={ref.range.start.character}
                                    highlightLength={ref.range.end.character - ref.range.start.character}
                                    previewWindowExtraLines={1}
                                />
                                </Link>
                            );
                        })
                }
            </div>;
        }

        return (
            <div className='references-group'>
                <div className='references-group__title' onClick={() => this.setState({ hidden: !this.state.hidden })}>
                    {this.props.isLocal ? <RepoIcon className='references-group__icon' /> : <GlobeIcon className='references-group__icon' />}
                    <div className='references-group__uri-path-part'>{uriStr}</div>
                    <div>{pathSplit.join('/')}{pathSplit.length > 0 ? '/' : ''}</div>
                    <div className='references-group__file-path-part'>{filePart}</div>
                    {this.state.hidden ? <RightIcon className='references-group__expand-icon' /> : <DownIcon className='references-group__expand-icon' />}
                </div>
                {refs}
            </div>
        );
    }
}

interface Props {
    onDismiss(): void;
}

interface State extends ReferencesState {
    docked: boolean;
    group: 'local' | 'external';
}

export class ReferencesWidget extends React.Component<Props, State> {
    public subscription: any;

    constructor(props: Props) {
        super(props);
        const u = url.parseBlob();
        const onRefs = Boolean(u.path && u.modal && u.modal === 'references');
        this.state = { ...store.getValue(), group: this.getRefsGroupFromUrl(window.location.href), docked: onRefs };
    }

    public componentDidMount(): void {
        this.subscription = store.subscribe(state => {
            this.setState({ ...state, group: this.state.group, docked: this.state.docked });
        });
        const u = url.parseBlob();
        if (this.state.docked) {
            triggerReferences({
                loc: {
                    repoURI: u.uri!,
                    rev: pageVars.Rev,
                    commitID: pageVars.CommitID,
                    path: u.path!,
                    line: u.line!,
                    char: u.char!
                },
                word: '' // TODO: derive the correct word from somewhere
            });
        }
        window.addEventListener('hashchange', this.handleHashChange);
    }

    public getRefsGroupFromUrl(urlStr: string): 'local' | 'external' {
        if (urlStr.indexOf('$references:local') !== -1) {
            return 'local';
        }
        if (urlStr.indexOf('$references:external') !== -1) {
            return 'external';
        }
        return 'local';
    }

    public componentWillUnmount(): void {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
        window.removeEventListener('hashchange', this.handleHashChange);
    }

    public isLoading(group: 'local' | 'external'): boolean {
        if (!this.state.context) {
            return false;
        }

        const state = store.getValue();
        const loadingRefs = state.fetches.get(refsFetchKey(this.state.context.loc, true)) === 'pending';
        const loadingXRefs = state.fetches.get(refsFetchKey(this.state.context.loc, false)) === 'pending';

        switch (group) {
            case 'local':
                return loadingRefs;
            case 'external':
                return loadingXRefs;
        }

        return false;
    }

    public render(): JSX.Element | null {
        if (!this.state.context) {
            return null;
        }
        const loc = locKey(this.state.context.loc);
        const refs = this.state.refsByLoc.get(loc);

        // References by fully qualified URI, like git://github.com/gorilla/mux?rev#mux.go
        const refsByUri = _.groupBy(refs, ref => ref.uri);

        const localPrefix = 'git://' + this.state.context.loc.repoURI;
        const [localRefs, externalRefs] = _(refsByUri).keys().partition(uri => uri.startsWith(localPrefix)).value();

        const isEmptyGroup = () => {
            switch (this.state.group) {
                case 'local':
                    return localRefs.length === 0;
                case 'external':
                    return externalRefs.length === 0;
            }
            return false;
        };

        const l = this.state.context.loc;
        return (
            <div className='references-widget'>
                <div className='references-widget__title-bar'>
                    <div className='references-widget__title-bar-title'>
                        {this.state.context.word}
                    </div>
                    <a className={'references-widget__title-bar-group' + (this.state.group === 'local' ? ' references-widget__title-bar-group--active' : '')}
                        href={url.toBlob({ uri: l.repoURI, rev: l.rev, path: l.path, line: l.line, char: l.char, modalMode: 'local', modal: 'references' })}
                        onClick={() => events.ShowLocalRefsButtonClicked.log()}>
                        This repository
                    </a>
                    <div className='references-widget__badge'>{localRefs.length}</div>
                    <a className={'references-widget__title-bar-group' + (this.state.group === 'external' ? ' references-widget__title-bar-group--active' : '')}
                        href={url.toBlob({ uri: l.repoURI, rev: l.rev, path: l.path, line: l.line, char: l.char, modalMode: 'external', modal: 'references' })}
                        onClick={() => events.ShowExternalRefsButtonClicked.log()}>
                        Other repositories
                    </a>
                    <div className='references-widget__badge'>{externalRefs.length}</div>
                    <CloseIcon className='references-widget__close-icon' onClick={() => this.props.onDismiss()} />
                </div>
                {
                    isEmptyGroup() && <div className='references-widget__placeholder'>
                        {this.isLoading(this.state.group) ? 'Working...' : 'No results'}
                    </div>
                }
                <div className='references-widget__groups'>
                    {
                        this.state.group === 'local' && localRefs.sort().map((uri, i) => {
                            const parsed = URI.parse(uri);
                            return <ReferencesGroup key={i} uri={parsed.hostname + parsed.path} path={parsed.fragment} isLocal={true} refs={refsByUri[uri]} />;
                        })
                    }
                    {
                        this.state.group === 'external' && externalRefs.map((uri, i) => { /* don't sort, to avoid jerky UI as new repo results come in */
                            const parsed = URI.parse(uri);
                            return <ReferencesGroup key={i} uri={parsed.hostname + parsed.path} path={parsed.fragment} isLocal={false} refs={refsByUri[uri]} />;
                        })
                    }
                </div>
            </div>
        );
    }

    private handleHashChange = (e: HashChangeEvent) => {
        const u = url.parseBlob(e!.newURL!);
        const shouldShow = Boolean(u.path && u.modal && u.modal === 'references');
        if (shouldShow) {
            this.setState({ ...this.state, group: this.getRefsGroupFromUrl(e!.newURL!), docked: true });
        }
    }
}

function getRefURL(ref: Reference, rev: string): string {
    if (rev) {
        rev = `@${rev}`;
    }
    const uri = URI.parse(ref.uri);
    return `/${uri.hostname}${uri.path}${rev}/-/blob/${uri.fragment}#L${ref.range.start.line + 1}`;
}
