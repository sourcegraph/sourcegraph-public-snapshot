import { fuzzyNumbersEqual } from '../numbers'
import type { PanelInfo, PanelsLayout } from '../types'

import { assert } from './assert'

export function callPanelCallbacks(
    panelsArray: PanelInfo[],
    layout: PanelsLayout,
    panelIdToLastNotifiedSizeMap: Record<string, number>
) {
    for (const size of layout) {
        const index = layout.indexOf(size)
        const panelData = panelsArray[index]
        assert(panelData, `Panel data not found for index ${index}`)

        const { callbacks, constraints, id: panelId } = panelData
        const { collapsedSize = 0, collapsible } = constraints

        const lastNotifiedSize = panelIdToLastNotifiedSizeMap[panelId]

        if (lastNotifiedSize == null || size !== lastNotifiedSize) {
            panelIdToLastNotifiedSizeMap[panelId] = size

            const { onResize, onExpand, onCollapse } = callbacks

            if (onResize) {
                onResize(size, lastNotifiedSize)
            }

            if (collapsible && (onCollapse || onExpand)) {
                if (
                    onExpand &&
                    (lastNotifiedSize == null || fuzzyNumbersEqual(lastNotifiedSize, collapsedSize)) &&
                    !fuzzyNumbersEqual(size, collapsedSize)
                ) {
                    onExpand()
                }

                if (
                    onCollapse &&
                    (lastNotifiedSize == null || !fuzzyNumbersEqual(lastNotifiedSize, collapsedSize)) &&
                    fuzzyNumbersEqual(size, collapsedSize)
                ) {
                    onCollapse()
                }
            }
        }
    }
}
