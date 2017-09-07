import { fetchBlameFile } from 'sourcegraph/backend'
import 'sourcegraph/blame/dom'
import { addHunks, BlameContext, setBlame, store } from 'sourcegraph/blame/store'
import { openFromJS } from 'sourcegraph/util/url'

/**
 * Measures the width of the given text string in pixels, using the given font.
 * @param text the literal text to measure.
 * @param font the font string, e.g. '12px Menlo'
 */
function measureTextWidth(text: string, font: string): number {
    const tmp = document.createElement('canvas').getContext('2d')
    tmp!.font = font
    return tmp!.measureText(text).width
}

/**
 * opens the commit if the userTriggered event exists and was inside the blame text shown
 * previously on the same line.
 * @param ctx the blame context
 * @param userTriggered the click event
 */
function maybeOpenCommit(ctx: BlameContext, userTriggered?: React.MouseEvent<HTMLDivElement>): void {
    if (!userTriggered) {
        return
    }
    const prevCtx = store.getValue().context
    const currentlyBlamed = document.querySelector('.blob td.code>.blame')
    if (!prevCtx || prevCtx.line !== ctx.line || !currentlyBlamed) {
        return // Not clicking on a line with blame info already showing.
    }
    const rev = currentlyBlamed.getAttribute('data-blame-rev')
    if (!rev) {
        return // e.g. if blame info failed to load or is currently loading
    }

    /**
     * Blame information is rendered in a ::before pseudo-element to avoid it being copied
     * when trying to copy code. This spared us from having to do absolute positioning of
     * the blame text onto the line itself as a non-child element of the blob view.
     *
     * However, the pseudo-element makes detecting clicks on the blame information (to view
     * the commit) hard because psuedo-elements have no DOM representation. We use a hack
     * here: We know the user clicked on the line with blame information, so we measure the
     * width of the blame text and if the mouse click was in its range then they clicked on
     * the blame text.
     *
     * TODO(future): Let's make blame text absolutely positioned on top of the line (not a
     * child of blob view), and turn all of this into a React component.
     */
    const x = userTriggered.clientX
    const blameTextStart = currentlyBlamed.getBoundingClientRect().right
    const blameTextEnd = blameTextStart + measureTextWidth(currentlyBlamed.getAttribute('data-blame')!, '12px Menlo')
    if (x < blameTextStart || x > blameTextEnd) {
        return // Not clicking on blame text
    }

    // TODO(future): For Umami Phabricator repos, the URL should be to Phabricator per #6487
    openFromJS(`https://${ctx.repoURI}/commit/${rev}`, userTriggered)
}

export function triggerBlame(ctx: BlameContext, userTriggered?: React.MouseEvent<HTMLDivElement>): void {
    maybeOpenCommit(ctx, userTriggered) // important: must come before setBlame() below
    setBlame({ ...store.getValue(), context: ctx, displayLoading: false })

    // Fetch the data.
    fetchBlameFile(ctx.repoURI, ctx.commitID, ctx.path, ctx.line, ctx.line).then(hunks => {
        if (hunks) {
            addHunks(ctx, hunks)
        }
    }).catch(e => {
        // TODO(slimsag): display error in UX
        console.error('failed to fetch blame info', e)
    })

    // After 250ms, if there is no data, the component will display a loading
    // indicator.
    setTimeout(() => {
        setBlame({ ...store.getValue(), displayLoading: true })
    }, 250)
}
