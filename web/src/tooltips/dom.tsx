import { highlightBlock } from 'highlight.js';
import * as H from 'history';
import * as marked from 'marked';
import { triggerReferences } from 'sourcegraph/references';
import { getSearchPath } from 'sourcegraph/search';
import { clearTooltip, store, TooltipState } from 'sourcegraph/tooltips/store';
import * as styles from 'sourcegraph/tooltips/styles';
import { events } from 'sourcegraph/tracking/events';
import { getCodeCellsForAnnotation, getModeFromExtension, highlightAndScrollToLine } from 'sourcegraph/util';
import * as url from 'sourcegraph/util/url';
import * as URI from 'urijs';

let tooltip: HTMLElement;
let loadingTooltip: HTMLElement;
let tooltipActions: HTMLElement;
let j2dAction: HTMLAnchorElement;
let findRefsAction: HTMLAnchorElement;
let searchAction: HTMLAnchorElement;
let moreContext: HTMLElement;

// tslint:disable-next-line:max-line-length
const closeIconSVG = '<svg width="10px" height="10px"><path fill="#93A9C8"  xmlns="http://www.w3.org/2000/svg" id="path0_fill" d="M 7.8565 7.86521C 7.66117 8.06054 7.3445 8.06054 7.14917 7.86521L 3.99917 4.71521L 0.851833 7.86254C 0.655167 8.05721 0.3385 8.05521 0.1445 7.85854C -0.0481667 7.66388 -0.0481667 7.34988 0.1445 7.15454L 3.29183 4.00721L 0.145167 0.860543C -0.0475001 0.663209 -0.0428332 0.346543 0.155167 0.153876C 0.349167 -0.0347905 0.6585 -0.0347905 0.8525 0.153876L 3.99917 3.30054L 7.1485 0.151209C 7.34117 -0.0467907 7.65783 -0.0507906 7.85583 0.141876C 8.05383 0.334543 8.05783 0.651209 7.86517 0.849209C 7.86183 0.852543 7.85917 0.855209 7.85583 0.858543L 4.7065 4.00788L 7.8565 7.15788C 8.05183 7.35321 8.0525 7.66988 7.8565 7.86521Z" /></svg >';
const searchIconSVG = '<svg width="12px" height="12px"><path fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg" id="path13_fill" d="M 4.75021 4.65905e-06C 7.36999 -0.00361534 9.49667 2.1172 9.50027 4.73698C 9.50167 5.7595 9.17264 6.7551 8.5622 7.5754L 11.1265 9.74432C 11.5382 10.0957 11.5872 10.7144 11.2358 11.1261C 10.8844 11.5379 10.2657 11.5868 9.85399 11.2355C 9.81473 11.202 9.77819 11.1654 9.74467 11.1261L 7.5752 8.56228C 5.46856 10.1236 2.49507 9.68156 0.933725 7.5749C -0.627615 5.46824 -0.185555 2.49476 1.92111 0.933425C 2.73957 0.326825 3.73145 -0.000435341 4.75021 4.65905e-06ZM 4.75021 8.5C 6.82127 8.5 8.50021 6.82106 8.50021 4.75C 8.50021 2.67894 6.82127 1 4.75021 1C 2.67915 1 1.00021 2.67894 1.00021 4.75C 1.00023 6.82106 2.67915 8.49998 4.75021 8.49998L 4.75021 8.5Z"/></svg>';
// tslint:disable-next-line:max-line-length
const referencesIconSVG = '<svg width="12px" height="8px"><path fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg" id="path15_fill" d="M 6.00625 8C 2.33125 8 0.50625 5.075 0.05625 4.225C -0.01875 4.075 -0.01875 3.9 0.05625 3.775C 0.50625 2.925 2.33125 0 6.00625 0C 9.68125 0 11.5063 2.925 11.9563 3.775C 12.0312 3.925 12.0312 4.1 11.9563 4.225C 11.5063 5.075 9.68125 8 6.00625 8ZM 6.00625 1.25C 4.48125 1.25 3.25625 2.475 3.25625 4C 3.25625 5.525 4.48125 6.75 6.00625 6.75C 7.53125 6.75 8.75625 5.525 8.75625 4C 8.75625 2.475 7.53125 1.25 6.00625 1.25ZM 6.00625 5.75C 5.03125 5.75 4.25625 4.975 4.25625 4C 4.25625 3.025 5.03125 2.25 6.00625 2.25C 6.98125 2.25 7.75625 3.025 7.75625 4C 7.75625 4.975 6.98125 5.75 6.00625 5.75Z"/></svg>';
const definitionIconSVG = '<svg width="11px" height="9px"><path fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg" id="path10_fill" d="M 6.325 8.4C 6.125 8.575 5.8 8.55 5.625 8.325C 5.55 8.25 5.5 8.125 5.5 8L 5.5 6C 2.95 6 1.4 6.875 0.825 8.7C 0.775 8.875 0.6 9 0.425 9C 0.2 9 -4.44089e-16 8.8 -4.44089e-16 8.575C -4.44089e-16 8.575 -4.44089e-16 8.575 -4.44089e-16 8.55C 0.125 4.825 1.925 2.675 5.5 2.5L 5.5 0.5C 5.5 0.225 5.725 8.88178e-16 6 8.88178e-16C 6.125 8.88178e-16 6.225 0.05 6.325 0.125L 10.825 3.875C 11.025 4.05 11.075 4.375 10.9 4.575C 10.875 4.6 10.85 4.625 10.825 4.65L 6.325 8.4Z"/></svg>';

/**
 * createTooltips initializes the DOM elements used for the hover
 * tooltip and "Loading..." text indicator, adding the former
 * to the DOM (but hidden). It is idempotent.
 */
export function createTooltips(history: H.History): void {
    if (document.querySelector('.sg-tooltip')) {
        return; // idempotence
    }

    tooltip = document.createElement('DIV');
    Object.assign(tooltip.style, styles.tooltip);
    tooltip.classList.add('sg-tooltip');
    tooltip.style.visibility = 'hidden';

    document.querySelector('.blob')!.appendChild(tooltip);

    loadingTooltip = document.createElement('DIV');
    loadingTooltip.appendChild(document.createTextNode('Loading...'));
    Object.assign(loadingTooltip.style, styles.loadingTooltip);

    tooltipActions = document.createElement('DIV');
    Object.assign(tooltipActions.style, styles.tooltipActions);

    moreContext = document.createElement('DIV');
    Object.assign(moreContext.style, styles.tooltipMoreActions);
    moreContext.appendChild(document.createTextNode('Click for more actions'));

    const definitionIcon = document.createElement('svg');
    definitionIcon.innerHTML = definitionIconSVG;
    Object.assign(definitionIcon.style, styles.definitionIcon);

    j2dAction = document.createElement('A') as HTMLAnchorElement;
    j2dAction.appendChild(definitionIcon);
    j2dAction.appendChild(document.createTextNode('Go to definition'));
    j2dAction.className = `btn btn-sm BtnGroup-item`;
    Object.assign(j2dAction.style, styles.tooltipAction);
    Object.assign(j2dAction.style, styles.tooltipActionNotLast);
    j2dAction.onclick = (e: MouseEvent) => {
        const before = url.parseBlob();
        events.GoToDefClicked.log();
        if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
            return;
        }
        e.preventDefault();
        const uri = URI.parse(j2dAction.href);

        const after = url.parseBlob(j2dAction.href);
        clearTooltip();
        if (after.line && before.uri === after.uri && before.rev === after.rev && before.line !== after.line) {
            // Handles URL update.
            highlightAndScrollToLine(history, after.uri!,
                after.rev!, after.path!, after.line, getCodeCellsForAnnotation(), false);
        } else {
            history.push(uri.path + '#' + uri.fragment);
        }
    };

    const referencesIcon = document.createElement('svg');
    referencesIcon.innerHTML = referencesIconSVG;
    Object.assign(referencesIcon.style, styles.referencesIcon);

    findRefsAction = document.createElement('A') as HTMLAnchorElement;
    findRefsAction.appendChild(referencesIcon);
    findRefsAction.appendChild(document.createTextNode('Find references'));
    Object.assign(findRefsAction.style, styles.tooltipAction);
    Object.assign(findRefsAction.style, styles.tooltipActionNotLast);
    findRefsAction.addEventListener('click', () => {
        const { context } = store.getValue();
        if (!context || !context.coords) {
            return;
        }
        const loc = {
            ...context.repoRevCommit,
            path: context.path,
            line: context.coords.line,
            char: context.coords.char
        };
        events.FindRefsClicked.log();
        triggerReferences({ loc, word: context.coords.word });
        hideTooltip();
    });

    const searchIcon = document.createElement('svg');
    searchIcon.innerHTML = searchIconSVG;
    Object.assign(searchIcon.style, styles.searchIcon);

    searchAction = document.createElement('A') as HTMLAnchorElement;
    searchAction.appendChild(searchIcon);
    searchAction.appendChild(document.createTextNode('Search'));
    Object.assign(searchAction.style, styles.tooltipAction);
    searchAction.onclick = (e: MouseEvent) => {
        events.SearchClicked.log();
        if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
            return;
        }
        e.preventDefault();
        const uri = URI.parse(searchAction.href);
        history.push(uri.path + '?' + uri.query);
    };

    tooltipActions.appendChild(j2dAction);
    tooltipActions.appendChild(findRefsAction);
    tooltipActions.appendChild(searchAction);
}

function constructBaseTooltip(): void {
    tooltip.appendChild(loadingTooltip);
    tooltip.appendChild(moreContext);
    tooltip.appendChild(tooltipActions);
}

/**
 * hideTooltip makes the tooltip on the DOM invisible.
 */
export function hideTooltip(): void {
    if (!tooltip) {
        return;
    }

    while (tooltip.firstChild) {
        tooltip.removeChild(tooltip.firstChild);
    }
    tooltip.style.visibility = 'hidden'; // prevent black dot of empty content
}

/**
 * updateTooltip displays the appropriate tooltip given current state (and may hide
 * the tooltip if no text is available).
 */
function updateTooltip(state: TooltipState): void {
    hideTooltip(); // hide before updating tooltip text

    const { target, data, docked, context } = state;

    if (!target) {
        // no target to show hover for; tooltip is hidden
        return;
    }
    if (!data) {
        // no data; bail
        return;
    }
    if (!context || (context.selectedText && context.selectedText.trim()) === '') {
        // no context or selected text is only whitespace; bail
        return;
    }

    constructBaseTooltip();
    loadingTooltip.style.display = data.loading ? 'block' : 'none';
    moreContext.style.display = docked || data.loading ? 'none' : 'flex';
    tooltipActions.style.display = docked ? 'flex' : 'none';

    if (context && context.selectedText) {
        j2dAction.style.display = 'none';
        findRefsAction.style.display = 'none';
    } else {
        j2dAction.style.display = 'block';
        findRefsAction.style.display = 'block';
    }

    j2dAction.href = data.j2dUrl ? data.j2dUrl : '';

    if (data && context && context.coords && context.path && context.repoRevCommit) {
        let revString = '';
        if (context.repoRevCommit.rev) {
            revString = `@${context.repoRevCommit.rev}`;
        }
        findRefsAction.href = `/${context.repoRevCommit.repoURI}${revString}/-/blob/${context.path}#L${context.coords.line}:${context.coords.char}$references`;
    } else {
        findRefsAction.href = '';
    }

    const searchText = context!.selectedText ? context!.selectedText! : target!.textContent!;
    if (searchText) {
        searchAction.href = getSearchPath({ files: '', repos: context.repoRevCommit.repoURI, q: searchText, matchCase: false, matchWord: false, matchRegex: false });
    } else {
        searchAction.href = '';
    }

    if (!data.loading) {
        loadingTooltip.style.visibility = 'hidden';

        if (!data.title) {
            // no tooltip text / search context; tooltip is hidden
            return;
        }

        const container = document.createElement('DIV');
        Object.assign(container.style, styles.divider);

        const tooltipText = document.createElement('DIV');
        tooltipText.className = `${getModeFromExtension(context.path)}`;
        Object.assign(tooltipText.style, styles.tooltipTitle);
        tooltipText.appendChild(document.createTextNode(data.title));

        container.appendChild(tooltipText);
        tooltip.insertBefore(container, moreContext);

        const closeContainer = document.createElement('a');
        Object.assign(closeContainer.style, styles.closeIcon);
        // TODO https://github.com/palantir/tslint/issues/2430
        // tslint:disable-next-line:no-unnecessary-callback-wrapper
        closeContainer.onclick = () => clearTooltip();

        if (docked) {
            const closeButton = document.createElement('svg');
            closeButton.innerHTML = closeIconSVG;
            closeContainer.appendChild(closeButton);
            container.appendChild(closeContainer);
        }

        highlightBlock(tooltipText);

        if (data.doc) {
            const tooltipDoc = document.createElement('DIV');
            Object.assign(tooltipDoc.style, styles.tooltipDoc);
            tooltipDoc.innerHTML = marked(data.doc, { gfm: true, breaks: true, sanitize: true });
            tooltip.insertBefore(tooltipDoc, moreContext);

            // Handle scrolling ourselves so that scrolling to the bottom of
            // the tooltip documentation does not cause the page to start
            // scrolling (which is a very jarring experience).
            tooltip.addEventListener('wheel', (e: WheelEvent) => {
                e.preventDefault();
                tooltipDoc.scrollTop += e.deltaY;
            });
        }
    } else {
        loadingTooltip.style.visibility = 'visible';
    }

    const scrollingElement = document.querySelector('.blob')!;
    const scrollingElementBound = scrollingElement.getBoundingClientRect();
    const blobTable = document.querySelector('.blob > table')!; // table that we're positioning tooltips relative to.
    const tableBound = blobTable.getBoundingClientRect(); // tables bounds
    const targetBound = target.getBoundingClientRect(); // our target elements bounds

    // Anchor it horizontally, prior to rendering to account for wrapping
    // changes to vertical height if the tooltip is at the edge of the viewport.
    const relLeft = targetBound.left - tableBound.left;
    tooltip.style.left = relLeft + 'px';

    // Anchor the tooltip vertically.
    const tooltipBound = tooltip.getBoundingClientRect();
    const relTop = (targetBound.top + scrollingElement.scrollTop) - scrollingElementBound.top;
    const margin = 5;
    let tooltipTop = relTop - (tooltipBound.height + margin);
    if ((tooltipTop - scrollingElement.scrollTop) < 0) {
        // Tooltip wouldn't be visible from the top, so display it at the
        // bottom.
        const relBottom = (targetBound.bottom + scrollingElement.scrollTop) - scrollingElementBound.top;
        tooltipTop = relBottom + margin;
    }
    tooltip.style.top = tooltipTop + 'px';

    // Make it all visible to the user.
    tooltip.style.visibility = 'visible';
}

window.addEventListener('keyup', (e: KeyboardEvent) => {
    if (e.keyCode === 27) {
        clearTooltip();
    }
});

store.subscribe(updateTooltip);
