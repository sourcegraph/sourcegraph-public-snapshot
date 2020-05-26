import * as React from 'react'
import { cleanup, render } from '@testing-library/react'
import { BadgeAttachment } from './BadgeAttachment'
import { BadgeAttachmentRenderOptions } from 'sourcegraph'

const base64icon =
    'data:image/svg+xml;base64,IDxzdmcgeG1sbnM9J2h0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnJyBzdHlsZT0id2lkdGg6MjRweDtoZWlnaHQ6MjRweCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSIjZmZmZmZmIj4gPHBhdGggZD0iIE0xMSwgOUgxM1Y3SDExTTEyLCAyMEM3LjU5LCAyMCA0LCAxNi40MSA0LCAxMkM0LCA3LjU5IDcuNTksIDQgMTIsIDRDMTYuNDEsIDQgMjAsIDcuNTkgMjAsIDEyQzIwLCAxNi40MSAxNi40MSwgMjAgMTIsIDIwTTEyLCAyQTEwLCAxMCAwIDAsIDAgMiwgMTJBMTAsIDEwIDAgMCwgMCAxMiwgMjJBMTAsIDEwIDAgMCwgMCAyMiwgMTJBMTAsIDEwIDAgMCwgMCAxMiwgMk0xMSwgMTdIMTNWMTFIMTFWMTdaIiAvPiA8L3N2Zz4g'

export const oldFormatBadge: Omit<BadgeAttachmentRenderOptions, 'kind'> = {
    icon: base64icon,
    light: { icon: base64icon },
    hoverMessage:
        'Search-based results - click to see how these results are calculated and how to get precise intelligence with LSIF.',
    linkURL: 'https://docs.sourcegraph.com/user/code_intelligence/basic_code_intelligence',
}

export const newFormatBadge: BadgeAttachmentRenderOptions = {
    ...oldFormatBadge,
    kind: 'info',
}

describe('BadgeAttachment', () => {
    afterAll(cleanup)

    it('renders an img element with a base64 icon', () => {
        // note '"kind" is missing'
        const { container } = render(
            <BadgeAttachment attachment={oldFormatBadge as BadgeAttachmentRenderOptions} isLightTheme={true} />
        )
        const item = container.querySelector('.badge-decoration-attachment__contents')
        expect(item).toBeTruthy()
        // we used to render an image with base64 content as a source
        expect(item?.nodeName.toLowerCase()).toBe('img')
    })

    it('renders an svg element with a predefined icon', () => {
        const { container } = render(<BadgeAttachment attachment={newFormatBadge} isLightTheme={true} />)
        const item = container.querySelector('.badge-decoration-attachment__icon-svg')
        expect(item).toBeTruthy()
        // now we are using a proper svg for icons
        expect(item?.nodeName.toLowerCase()).toBe('svg')
    })
})
