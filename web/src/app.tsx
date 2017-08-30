import { Tree, TreeHeader } from '@sourcegraph/components/lib/Tree';
import { content, flex, vertical } from 'csstips';
import * as moment from 'moment';
import * as React from 'react';
import { render } from 'react-dom';
import * as backend from 'sourcegraph/backend';
import * as xhr from 'sourcegraph/backend/xhr';
import { triggerBlame } from 'sourcegraph/blame';
import { dismissReferencesWidget, injectReferencesWidget } from 'sourcegraph/references/inject';
import { injectAdvancedSearchDrawer, injectAdvancedSearchToggle, injectSearchBox, injectSearchInputHandler, injectSearchResults } from 'sourcegraph/search/inject';
import { injectShareWidget } from 'sourcegraph/share';
import { addAnnotations } from 'sourcegraph/tooltips';
import { handleQueryEvents } from 'sourcegraph/tracking/analyticsUtils';
import { events, viewEvents } from 'sourcegraph/tracking/events';
import { getPathExtension, supportedExtensions } from 'sourcegraph/util';
import * as activeRepos from 'sourcegraph/util/activeRepos';
import { pageVars } from 'sourcegraph/util/pageVars';
import { sourcegraphContext } from 'sourcegraph/util/sourcegraphContext';
import { CodeCell } from 'sourcegraph/util/types';
import * as url from 'sourcegraph/util/url';
import { style } from 'typestyle';

const navHeight = '48px';

window.onhashchange = hash => {
    const oldURL = url.parseBlob(hash.oldURL!);
    const newURL = url.parseBlob(hash.newURL!);
    if (!newURL.path || !newURL.line) {
        return;
    }
    if (oldURL.line === newURL.line) {
        // prevent e.g. re-scrolling to same line on toggling refs group
        //
        // also prevent highlightLine from retriggering onhashchange
        // recursively.
        return;
    }
    const cells = getCodeCellsForAnnotation();
    highlightAndScrollToLine(newURL.uri!, pageVars.CommitID, newURL.path, newURL.line, cells, false);
};

window.addEventListener('DOMContentLoaded', () => {
    registerListeners();
    xhr.useAccessToken(sourcegraphContext.accessToken);

    // Be a bit proactive and try to fetch/store active repos now. This helps
    // on the first search query, and when the data in local storage is stale.
    activeRepos.get().catch(err => console.error(err));

    if (window.location.pathname === '/') {
        viewEvents.Home.log();
        injectSearchBox();
    } else {
        injectSearchInputHandler();
        injectAdvancedSearchToggle();
        injectAdvancedSearchDrawer();
    }
    if (window.location.pathname === '/search') {
        viewEvents.SearchResults.log();
        injectSearchResults();
    }

    const cloning = document.querySelector('#cloning');
    if (cloning) {
        // TODO: Actually poll the backend instead of just reloading the page
        // every 5s.
        setTimeout(() => {
            window.location.reload(false);
        }, 5000);
    }

    injectTreeViewer();
    injectReferencesWidget();
    injectShareWidget();
    const u = url.parseBlob();
    if (u.uri && u.path) {
        // blob view, add tooltips
        const rev = pageVars.Rev;
        const commitID = pageVars.CommitID;
        const cells = getCodeCellsForAnnotation();
        if (supportedExtensions.has(getPathExtension(u.path!))) {
            addAnnotations(u.path!, { repoURI: u.uri!, rev, commitID }, cells);
        }
        if (u.line) {
            highlightAndScrollToLine(u.uri!, commitID, u.path!, u.line, cells, false);
        }

        // Log blob view
        viewEvents.Blob.log({ repo: u.uri!, commitID, path: u.path!, language: getPathExtension(u.path!) });

        // Add click handlers to all lines of code, which highlight and add
        // blame information to the line.
        for (const [index, tr] of document.querySelectorAll('.blobview tr').entries()) {
            tr.addEventListener('click', () => {
                if (u.uri && u.path) {
                    highlightLine(u.uri, commitID, u.path, index + 1, cells, true);
                }
            });
        }
    } else if (u.uri) {
        // tree view
        viewEvents.Tree.log();
    }

    // Log events, if necessary, based on URL querystring, and strip tracking-related parameters
    // from the URL and browser history
    // Note that this is a destructive operation (it changes the page URL and replaces browser state)
    handleQueryEvents(window.location.href);
});

function injectTreeViewer(): void {
    const mount = document.querySelector('#tree-viewer');
    if (!mount) {
        return;
    }

    const repoURL = url.parse();
    const blobURL = url.parseBlob();
    const treeURL = url.parseTree();
    const uri = blobURL.uri || treeURL.uri || repoURL.uri;
    const rev = blobURL.rev || treeURL.rev || repoURL.rev;
    const path = blobURL.path || treeURL.path || '/';

    // Force show the tree viewer on any non-blob page OR on any blob page that
    // is just telling the user "sorry, we don't support binary files."
    const forceShow = !url.isBlob(blobURL) || Boolean(document.querySelector('.blobview-binary'));

    showExplorerTreeIfNecessary(forceShow);
    document.querySelector('#file-explorer')!.addEventListener('click', () => {
        handleToggleExplorerTree();
    });
    backend.localStoreListAllFiles(uri!, pageVars.CommitID).then(resp => {
        // For Firefox, this can look like it's as tall as the overflown content (when the actual parent element is much smaller).
        // We explicitly set the height so scrolling calculations are consistent across browsers.
        const el = <div className={style(vertical, { height: `calc(100vh - ${navHeight} - ${navHeight} - 1px)` })} >
            <TreeHeader className={style(content)} title='Files' onDismiss={handleToggleExplorerTree} />
            <Tree initSelectedPath={path} onSelectPath={(p, isDir) => {
                if (!isDir) {
                    window.location.href = url.toBlob({ uri, rev, path: p });
                    return;
                } else if (!url.isBlob(blobURL)) {
                    // Directory, and on a tree or repo page. Update the URL so
                    // the user can share a link to a specific dir.
                    const newURL = url.toTree({ uri, rev, path: p });
                    if (newURL === (window.location.pathname + window.location.hash)) {
                        return; // don't push state twice if user clicks twice
                    }
                    window.history.pushState(null, '', newURL);
                }
            }} className={style(flex)} paths={resp.map(res => res.name)} />
        </div>;
        render(el, mount);
    }).catch(e => {
        // TODO(slimsag): display error in UX
        console.error('failed to list all files', e);
    });
}

function handleToggleExplorerTree(): void {
    // TODO(slimsag): add eventLogger calls
    // eventLogger.logFileTreeToggleClicked({toggled: toggled});
    const isShown = window.localStorage.getItem('show-explorer') === 'true';
    window.localStorage.setItem('show-explorer', isShown ? 'false' : 'true');
    const treeViewer = document.querySelector('#tree-viewer')! as HTMLElement;
    treeViewer.style.display = isShown ? 'none' : 'flex';
}

function showExplorerTreeIfNecessary(force: boolean): void {
    // TODO(slimsag): add eventLogger calls
    // eventLogger.logFileTreeToggleClicked({toggled: toggled});
    const shouldShow = force || window.localStorage.getItem('show-explorer') === 'true';
    const treeViewer = document.querySelector('#tree-viewer')! as HTMLElement;
    treeViewer.style.display = shouldShow ? 'flex' : 'none';
}

function registerListeners(): void {
    const openOnGitHub = document.querySelector('.github')!;
    if (openOnGitHub) {
        openOnGitHub.addEventListener('click', () => events.OpenInCodeHostClicked.log());
    }
    const openOnDesktop = document.querySelector('.open-on-desktop')!;
    if (openOnDesktop) {
        openOnDesktop.addEventListener('click', () => events.OpenInNativeAppClicked.log());
    }
}

function highlightLine(repoURI: string, commitID: string, path: string, line: number, cells: CodeCell[], userTriggered: boolean): void {
    triggerBlame({
        time: moment(),
        repoURI,
        commitID,
        path,
        line
    });

    const currentlyHighlighted = document.querySelectorAll('.sg-highlighted') as NodeListOf<HTMLElement>;
    for (const cellElem of currentlyHighlighted) {
        cellElem.classList.remove('sg-highlighted');
        cellElem.style.backgroundColor = 'inherit';
    }

    const cell = cells[line - 1];
    cell.cell.style.backgroundColor = '#1c2736';
    cell.cell.classList.add('sg-highlighted');

    // Update the URL.
    const u = url.parseBlob();
    u.line = line;

    // Dismiss the references widget, if highlighting this line was user
    // triggered (not done automatically onload).
    const referencesOpen = u.modal === 'references';
    if (referencesOpen && userTriggered) {
        u.modal = undefined;
        u.modalMode = undefined;
    }

    // Update selection parameter in open-on-desktop button href.
    const openOnDesktop = document.querySelector('.open-on-desktop') as HTMLAnchorElement | undefined;
    if (openOnDesktop) {
        openOnDesktop.href = openOnDesktop.href.replace(/(&selection=[\d:-]+)?$/, `&selection=${line}:1`);
    }

    // Check URL change first, since this function can be called in response to
    // onhashchange.
    if (url.toBlob(u) === (window.location.pathname + window.location.hash)) {
        return;
    }

    window.history.pushState(null, '', url.toBlobHash(u));
    if (referencesOpen && userTriggered) {
        dismissReferencesWidget();
    }
}

function highlightAndScrollToLine(repoURI: string, commitID: string, path: string, line: number, cells: CodeCell[], userTriggered: boolean): void {
    highlightLine(repoURI, commitID, path, line, cells, userTriggered);

    // Scroll to the line.
    const scrollingElement = document.querySelector('#blob-table')!;
    const viewportBound = scrollingElement.getBoundingClientRect();
    const blobTable = document.querySelector('#blob-table>table')!; // table that we're positioning tooltips relative to.
    const tableBound = blobTable.getBoundingClientRect(); // tables bounds
    const cell = cells[line - 1];
    const targetBound = cell.cell.getBoundingClientRect(); // our target elements bounds

    scrollingElement.scrollTop = targetBound.top - tableBound.top - (viewportBound.height / 2) + (targetBound.height / 2);
}

function getCodeCellsForAnnotation(): CodeCell[] {
    const table = document.querySelector('#blob-table>table') as HTMLTableElement;
    const cells = Array.from(table.rows).map(row => {
        const line = parseInt(row.cells[0].getAttribute('data-line')!, 10);
        const codeCell: HTMLTableDataCellElement = row.cells[1]; // the actual cell that has code inside; each row contains multiple columns
        return {
            cell: codeCell as HTMLElement,
            eventHandler: codeCell, // allways the TD element
            line
        };
    });

    return cells;
}
