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
    public state: S = {
        ...store.getValue(),
        showTree: true,
        showRefs: false
    };
    private subscription: Rx.Subscription;

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
        if (this.props.location.hash !== nextProps.location.hash && nextProps.history.action === 'POP') {
            // handle 'back' and 'forward'
            this.scrollToLine(nextProps);
        } else if (this.props.location.pathname !== nextProps.location.pathname) {
            this.scrollToLine(nextProps);
        }
    }

    public componentWillUnmount(): void {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    }

    public render(): JSX.Element | null {
        const key = contextKey(this.props.uri, this.props.commitID, this.props.path);
        const files = this.state.files.get(key) || [];
        return <div className='repo'>
            <RepoSubnav {...this.props} onClickNavigation={() => this.setState({ showTree: !this.state.showTree })} />
            <div className='container'>
                {this.state.showTree && <div id='tree'>
                    <TreeHeader title='Files' onDismiss={() => this.setState({ showTree: false })} />
                    <div className='contents'><Tree scrollRootSelector='#tree' selectedPath={this.props.path} onSelectPath={this.selectTreePath} paths={files} /></div>
                </div>}
                <div className='blob'>
                    {!this.state.highlightedContents.has(key) && <div className='content' /> /* render placeholder for layout before content is fetched */}
                    {this.state.highlightedContents.has(key) &&
                        <Blob onClick={this.handleBlobClick}
                            applyAnnotations={this.applyAnnotations}
                            scrollToLine={this.scrollToLine}
                            html={this.state.highlightedContents.get(key)} />}
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

    private selectTreePath = (path: string, isDir: boolean) => {
        if (!isDir) {
            this.props.history.push(url.toBlob({ uri: this.props.uri, rev: this.props.rev, path }));
        }
    }

    private handleBlobClick: React.MouseEventHandler<HTMLDivElement> = e => {
        const target = e.target!;
        const row: HTMLTableRowElement = (target as any).closest('tr');
        if (!row) {
            return;
        }
        const line = parseInt(row.firstElementChild!.getAttribute('data-line')!, 10);
        highlightLine(this.props.history, this.props.uri, this.props.commitID, this.props.path!, line, getCodeCellsForAnnotation(), true);
    }

    private scrollToLine = (props: Props = this.props) => {
        const line = url.parseHash(props.location.hash).line;
        if (line) {
            highlightAndScrollToLine(props.history, props.uri,
                props.commitID, props.path!, line, getCodeCellsForAnnotation(), false);
        }
    }

    private applyAnnotations = () => {
        const cells = getCodeCellsForAnnotation();
        if (supportedExtensions.has(getPathExtension(this.props.path!))) {
            addAnnotations(this.props.history, this.props.path!,
                { repoURI: this.props.uri!, rev: this.props.rev!, commitID: this.props.commitID }, cells);
        }
    }

    private fetch(props: Props): void {
        const key = contextKey(props.uri, props.commitID, props.path);
        if (!this.state.files.has(key)) {
            listAllFiles(props.uri, props.commitID)
                .then(files => {
                    const state = store.getValue();
                    setState({ ...state, files: state.files.set(key, files.map(file => file.name)) });
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

interface BlobProps {
    html: string;
    onClick: React.MouseEventHandler<HTMLDivElement>;
    applyAnnotations: () => void;
    scrollToLine: () => void;
}

class Blob extends React.Component<BlobProps, {}> {
    private ref: any;

    public shouldComponentUpdate(nextProps: BlobProps): boolean {
        return this.props.html !== nextProps.html;
    }

    public render(): JSX.Element | null {
        return <div className='content' onClick={this.props.onClick} ref={ref => {
            if (!this.ref && ref) {
                // first mount, do initial scroll
                this.props.scrollToLine();
            }
            this.ref = ref;
            this.props.applyAnnotations();
        }} dangerouslySetInnerHTML={{ __html: this.props.html }} />;
    }
}
