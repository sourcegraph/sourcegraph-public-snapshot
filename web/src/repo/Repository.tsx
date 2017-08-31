import { Tree, TreeHeader } from '@sourcegraph/components/lib/Tree';
import * as H from 'history';
import * as React from 'react';
import * as Rx from 'rxjs';
import { fetchBlobHighlightContentTable, listAllFiles } from 'sourcegraph/backend';
import { RepoSubnav } from 'sourcegraph/nav';
import { ReferencesWidget } from 'sourcegraph/references/ReferencesWidget';
import { contextKey, setState, State, store } from 'sourcegraph/repo/store';
import { addAnnotations } from 'sourcegraph/tooltips';
import { getCodeCellsForAnnotation, getPathExtension, highlightAndScrollToLine, highlightLine, supportedExtensions } from 'sourcegraph/util';
import * as url from 'sourcegraph/util/url';

export interface Props {
    uri: string;
    rev?: string;
    commitID: string;
    path?: string;
    location: H.Location;
    history: H.History;
}

export interface S extends State {
    showRefs: boolean;
    showTree: boolean;
}

export class Repository extends React.Component<Props, S> {
    public hashWatcher: any;
    public subscription: Rx.Subscription;
    public state: S = {
        ...store.getValue(),
        showTree: true,
        showRefs: false
    };

    constructor(props: Props) {
        super(props);

        const u = url.parseBlob();
        this.state.showRefs = Boolean(u.path && u.modal && u.modal === 'references');
    }

    public componentDidMount(): void {
        this.subscription = store.subscribe(state => this.setState(state));
        this.fetch(this.props);
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.fetch(nextProps);
        const hash = url.parseHash(nextProps.location.hash);
        const showRefs = Boolean(nextProps.path && hash.modal && hash.modal === 'references');
        if (showRefs !== this.state.showRefs) {
            this.setState({ showRefs });
        }
    }

    public componentWillUnmount(): void {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    }

    public render(): JSX.Element | null {
        const key = contextKey(this.props.uri, this.props.commitID, this.props.path);
        return <div className='repo'>
            <RepoSubnav {...this.props} onClickNavigation={() => this.setState({ showTree: !this.state.showTree })} />
            <div className='container'>
                {this.state.showTree && <div id='tree'>
                    <TreeHeader title='Files' onDismiss={() => this.setState({ showTree: false })} />
                    {this.state.files.has(key) && <div className='contents'><Tree initSelectedPath={this.props.path} onSelectPath={(p, isDir) => {
                        if (!isDir) {
                            this.props.history.push(url.toBlob({ uri: this.props.uri, rev: this.props.rev, path: p }));
                        }/* else if (!url.isBlob(blobURL)) {
                            // Directory, and on a tree or repo page. Update the URL so
                            // the user can share a link to a specific dir.
                            const newURL = url.toTree({ uri: this.props.uri, rev: this.props.rev, path: p });
                            if (newURL === (window.location.pathname + window.location.hash)) {
                                return; // don't push state twice if user clicks twice
                            }
                            window.history.pushState(null, '', newURL);
                        }*/
                    }} paths={this.state.files.get(key).map(res => res.name)} /></div>}
                </div>}
                <div className='blob'>
                    {!this.state.highlightedContents.has(key) && <div className='content' /> /* render placeholder for layout before content is fetched */}
                    {this.state.highlightedContents.has(key) &&
                        <div className='content' ref={ref => {
                            if (this.props.path && ref) {
                                const cells = getCodeCellsForAnnotation();
                                if (supportedExtensions.has(getPathExtension(this.props.path!))) {
                                    addAnnotations(this.props.path!,
                                        { repoURI: this.props.uri!, rev: this.props.rev!, commitID: this.props.commitID }, cells);
                                }

                                // TODO(john): it's dangerous to do this here.
                                const line = url.parseHash(this.props.location.hash).line;
                                if (line) {
                                    highlightAndScrollToLine(this.props.uri,
                                        this.props.commitID, this.props.path!, line, getCodeCellsForAnnotation(), false);
                                }

                                for (const [index, tr] of document.querySelectorAll('.blob tr').entries()) {
                                    if (tr.classList.contains('sg-line-listener')) {
                                        return;
                                    }
                                    tr.classList.add('sg-line-listener');
                                    tr.addEventListener('click', () => {
                                        highlightLine(this.props.uri, this.props.commitID, this.props.path!, index + 1, getCodeCellsForAnnotation(), true);
                                    });
                                }
                            }
                        }} dangerouslySetInnerHTML={{ __html: this.state.highlightedContents.get(key) }} />}
                    {this.state.showRefs && <div className='ref-panel'>
                        <ReferencesWidget onDismiss={() => {
                            const currURL = url.parseBlob();
                            this.props.history.push(url.toBlob({ ...currURL, modal: undefined, modalMode: undefined }));
                        }} />
                    </div>}
                </div>
            </div>
        </div>;
    }

    private fetch(props: Props): void {
        const key = contextKey(props.uri, props.commitID, props.path);
        if (!this.state.files.has(key)) {
            listAllFiles(props.uri, props.commitID)
                .then(files => {
                    const state = store.getValue();
                    setState({ ...state, files: state.files.set(key, files) });
                });
        }

        if (props.path && !this.state.highlightedContents.has(key)) {
            fetchBlobHighlightContentTable(props.uri, props.commitID, props.path)
                .then(highlightedContents => {
                    const state = store.getValue();
                    setState({ ...state, highlightedContents: state.highlightedContents.set(key, highlightedContents) });
                });
        }
    }
}
