import { Tree, TreeHeader } from '@sourcegraph/components/lib/Tree';
import * as H from 'history';
import * as React from 'react';
import { Subject, Subscription } from 'rxjs';
import { ReferencesWidget } from 'sourcegraph/references/ReferencesWidget';
import { fetchBlobHighlightContentTable, listAllFiles } from 'sourcegraph/repo/backend';
import { addAnnotations } from 'sourcegraph/tooltips';
import { clearTooltip } from 'sourcegraph/tooltips/store';
import { getCodeCellsForAnnotation, getPathExtension, highlightAndScrollToLine, highlightLine, supportedExtensions } from 'sourcegraph/util';
import * as url from 'sourcegraph/util/url';
import { RepoNav } from './RepoNav';

export interface Props {
    repoPath: string;
    rev?: string;
    commitID: string;
    filePath?: string;
    location: H.Location;
    history: H.History;
}

interface State {
    /**
     * show the references panel
     */
    showRefs: boolean;
    /**
     * show the file tree explorer
     */
    showTree: boolean;
    /**
     * an array of file paths in the repository
     */
    files?: string[];
    /**
     * the HTML string for the Blob component
     */
    highlightedContents?: string;
}

export class Repository extends React.Component<Props, State> {
    public state: State = {
        showTree: true,
        showRefs: false
    };
    private componentUpdates = new Subject<Props>();
    private subscriptions = new Subscription();

    constructor(props: Props) {
        super(props);
        const u = url.parseBlob();
        this.state.showRefs = Boolean(u.path && u.modal && u.modal === 'references');
        this.subscriptions.add(
            this.componentUpdates
                .switchMap(props => listAllFiles({ repoPath: props.repoPath, commitID: props.commitID }))
                .subscribe((files: string[]) => this.setState({ files }))
        );
        this.subscriptions.add(
            this.componentUpdates
                .filter(props => !!props.filePath)
                .switchMap(props => fetchBlobHighlightContentTable({ repoPath: props.repoPath, commitID: props.commitID, filePath: props.filePath! }))
                .subscribe((highlightedContents: string) => this.setState({ highlightedContents }))
        );
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props);
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps);
        const hash = url.parseHash(nextProps.location.hash);
        const showRefs = Boolean(nextProps.filePath && hash.modal && hash.modal === 'references');
        if (showRefs !== this.state.showRefs) {
            this.setState({ showRefs });
        }
        if (this.props.location.hash !== nextProps.location.hash && nextProps.history.action === 'POP') {
            // handle 'back' and 'forward'
            this.scrollToLine(nextProps);
        } else if (this.props.location.pathname !== nextProps.location.pathname) {
            clearTooltip(); // clear tooltip when transitioning between files
            this.scrollToLine(nextProps);
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe();
    }

    public render(): JSX.Element | null {
        return (
            <div className='repository'>
                <RepoNav {...this.props} onClickNavigation={() => this.setState({ showTree: !this.state.showTree })} />
                <div className='repository__content'>
                    {
                        this.state.showTree &&
                            <div id='explorer' className='repository__sidebar'>
                                <TreeHeader className='repository__tree-header' title='Files' onDismiss={() => this.setState({ showTree: false })} />
                                <Tree
                                    className='repository__tree'
                                    scrollRootSelector='#explorer'
                                    selectedPath={this.props.filePath}
                                    onSelectPath={this.selectTreePath}
                                    paths={this.state.files || []}
                                />
                            </div>
                    }
                    <div className='repository__viewer'>
                        {
                            this.state.highlightedContents ?
                                <Blob onClick={this.handleBlobClick}
                                    applyAnnotations={this.applyAnnotations}
                                    scrollToLine={this.scrollToLine}
                                    html={this.state.highlightedContents} /> :
                                /* render placeholder for layout before content is fetched */
                                <div></div>
                        }
                        {
                            this.state.showRefs &&
                                <ReferencesWidget onDismiss={() => {
                                    const currURL = url.parseBlob();
                                    this.props.history.push(url.toBlob({ ...currURL, modal: undefined, modalMode: undefined }));
                                }} />
                        }
                    </div>
                </div>
            </div>
        );
    }

    private selectTreePath = (path: string, isDir: boolean) => {
        if (!isDir) {
            this.props.history.push(url.toBlob({ uri: this.props.repoPath, rev: this.props.rev, path }));
        }
    }

    private handleBlobClick: React.MouseEventHandler<HTMLDivElement> = e => {
        const target = e.target!;
        const row: HTMLTableRowElement = (target as any).closest('tr');
        if (!row) {
            return;
        }
        const line = parseInt(row.firstElementChild!.getAttribute('data-line')!, 10);
        highlightLine(this.props.history, this.props.repoPath, this.props.commitID, this.props.filePath!, line, getCodeCellsForAnnotation(), true);
    }

    private scrollToLine = (props: Props = this.props) => {
        const line = url.parseHash(props.location.hash).line;
        if (line) {
            highlightAndScrollToLine(props.history, props.repoPath,
                props.commitID, props.filePath!, line, getCodeCellsForAnnotation(), false);
        }
    }

    private applyAnnotations = () => {
        const cells = getCodeCellsForAnnotation();
        if (supportedExtensions.has(getPathExtension(this.props.filePath!))) {
            addAnnotations(this.props.history, this.props.filePath!,
                { repoURI: this.props.repoPath!, rev: this.props.rev!, commitID: this.props.commitID }, cells);
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
        return (
            <div className='blob' onClick={this.props.onClick} ref={ref => {
                if (!this.ref && ref) {
                    // first mount, do initial scroll
                    this.props.scrollToLine();
                }
                this.ref = ref;
                this.props.applyAnnotations();
            }} dangerouslySetInnerHTML={{ __html: this.props.html }} />
        );
    }
}
