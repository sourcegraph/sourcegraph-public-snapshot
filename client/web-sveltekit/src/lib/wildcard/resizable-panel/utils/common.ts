import type { PanelConstraints, PanelId, PanelInfo, PanelsLayout } from '../types'

let counter = 0

export function getId(): string {
    counter++
    return counter.toString()
}

export function findPanelDataIndex(panelDataArray: PanelInfo[], panelId: PanelId): number {
    return panelDataArray.findIndex(prevPanelData => prevPanelData.id === panelId)
}

export interface PanelMetadata extends PanelConstraints {
    panelSize: number
    pivotIndices: number[]
}

export function getPanelMetadata(panelDataArray: PanelInfo[], panel: PanelInfo, layout: PanelsLayout): PanelMetadata {
    const panelIndex = findPanelDataIndex(panelDataArray, panel.id)

    const isLastPanel = panelIndex === panelDataArray.length - 1
    const pivotIndices = isLastPanel ? [panelIndex - 1, panelIndex] : [panelIndex, panelIndex + 1]

    const panelSize = layout[panelIndex]

    return {
        panelSize,
        pivotIndices,
        ...panel.constraints,
    }
}

export function sortPanels(panels: PanelInfo[]): PanelInfo[] {
    return panels.sort((panelA, panelB) => {
        const orderA = panelA.order
        const orderB = panelB.order
        if (orderA === undefined && orderB === undefined) {
            return 0
        }

        if (orderA === undefined) {
            return -1
        }

        if (orderB === undefined) {
            return 1
        }

        return orderA - orderB
    })
}
