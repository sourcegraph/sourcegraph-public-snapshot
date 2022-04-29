import { render } from 'react-dom'

import { setLinkComponent, AnchorLink } from '@sourcegraph/wildcard'

import { App } from './App'

setLinkComponent(AnchorLink)

function onOpen(id: string): void {
    console.log('open', id)
}

function onPreviewChange(id: string): void {
    console.log('preview', id)
}

const node = document.querySelector('#main') as HTMLDivElement
render(<App onOpen={onOpen} onPreviewChange={onPreviewChange} />, node)
