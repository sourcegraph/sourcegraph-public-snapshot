import React from 'react'

import { afterEach, describe, expect, it } from '@jest/globals'
import { cleanup, fireEvent } from '@testing-library/react'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { type Panel, TabbedPanelContent, useBuiltinTabbedPanelViews } from './TabbedPanelContent'
import { panels, panelProps } from './TabbedPanelContent.fixtures'

describe('TabbedPanel', () => {
    const location = {
        pathname: `/${panelProps.repoName}`,
        search: '?L4:7',
        hash: `#tab=${panels[0].id}`,
    }
    const route = `${location.pathname}${location.search}${location.hash}`

    afterEach(cleanup)

    const TabbedPanelContentWithPanels: React.FunctionComponent<
        { panels: Panel[] } & React.ComponentPropsWithoutRef<typeof TabbedPanelContent>
    > = ({ panels, ...props }) => {
        useBuiltinTabbedPanelViews(panels)
        return <TabbedPanelContent {...props} />
    }

    it('preserves `location.pathname` and `location.hash` on tab change', async () => {
        const renderResult = renderWithBrandedContext(
            <TabbedPanelContentWithPanels panels={panels} {...panelProps} />,
            { route }
        )

        const panelToSelect = panels[2]
        const panelButton = await renderResult.findByRole('tab', { name: panelToSelect.title })
        fireEvent.click(panelButton)

        expect(renderResult.locationRef.current?.pathname).toEqual(location.pathname)
        expect(renderResult.locationRef.current?.search).toEqual(location.search)
        expect(renderResult.locationRef.current?.hash).toEqual(`#tab=${panelToSelect.id}`)
    })
})
