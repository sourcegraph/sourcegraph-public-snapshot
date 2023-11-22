import type { CategoricalChartContent } from '../../../../../../core'

export interface PreviewDatum {
    name: string
    value: number
    fill: string
}

export const DEFAULT_PREVIEW_MOCK: CategoricalChartContent<PreviewDatum> = {
    data: [
        {
            name: 'Covered',
            value: 0.3,
            fill: 'var(--oc-grape-7)',
        },
        {
            name: 'Not covered',
            value: 0.7,
            fill: 'var(--oc-orange-7)',
        },
    ],
    getDatumName: datum => datum.name,
    getDatumColor: datum => datum.fill,
    getDatumValue: datum => datum.value,
}
