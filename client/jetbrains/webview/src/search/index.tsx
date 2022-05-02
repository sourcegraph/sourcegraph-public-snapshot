import { render } from 'react-dom'

import { ContentMatch } from '@sourcegraph/shared/src/search/stream'
import { setLinkComponent, AnchorLink } from '@sourcegraph/wildcard'

import { App } from './App'

setLinkComponent(AnchorLink)

// @TODO: We need a mechanism for fetching the whole file for the follwoing two functions and make
// sure to cache the result (so that any given file is only ever downloaded once). A simple Map
// should do the trick.
//
// When we have this, these callbacks should just link through to our Java bridge.
function onOpen(match: ContentMatch, lineIndex: number): void {
    const line = match.lineMatches[lineIndex]
    console.log('open', match, line)
}
function onPreviewChange(match: ContentMatch, lineIndex: number): void {
    const line = match.lineMatches[lineIndex]
    console.log('preview', match, line)
}
function onPreviewClear(): void {
    // nothing
}

const node = document.querySelector('#main') as HTMLDivElement
render(<App onOpen={onOpen} onPreviewChange={onPreviewChange} onPreviewClear={onPreviewClear} />, node)
