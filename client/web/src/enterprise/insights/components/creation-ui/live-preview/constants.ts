import type { Series } from '@sourcegraph/wildcard'

import type { CategoricalChartContent } from '../../../core'

const getYValue = (datum: MockSeriesDatum): number => datum.value
const getXValue = (datum: MockSeriesDatum): Date => new Date(datum.x)

interface MockSeriesDatum {
    value: number
    x: number
}

export const SERIES_MOCK_CHART: Series<MockSeriesDatum>[] = [
    {
        id: 'series_001',
        data: [
            { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, value: 20 },
            { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, value: 40 },
            { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 110 },
            { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 105 },
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 160 },
            { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 184 },
            { x: 1588965700286, value: 200 },
        ],
        name: 'Go 1.11',
        color: 'var(--oc-indigo-7)',
        getYValue,
        getXValue,
    },
    {
        id: 'series_002',
        data: [
            { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, value: 200 },
            { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, value: 177 },
            { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 150 },
            { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 165 },
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 100 },
            { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 85 },
            { x: 1588965700286, value: 50 },
        ],
        name: 'Go 1.12',
        color: 'var(--oc-orange-7)',
        getYValue,
        getXValue,
    },
]

interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    linkURL: string
    group?: string
}

export const COMPUTE_MOCK_CHART: CategoricalChartContent<LanguageUsageDatum> = {
    getDatumValue: datum => datum.value,
    getDatumName: datum => datum.name,
    getDatumColor: datum => datum.fill,
    getCategory: datum => datum.group,
    data: [
        {
            group: 'Group 1',
            name: 'Name 1',
            value: 241,
            fill: 'var(--oc-yellow-9)',
            linkURL: '',
        },
        {
            group: 'Group 1',
            name: 'Name 2',
            value: 148,
            fill: 'var(--oc-grape-9)',
            linkURL: '',
        },
        {
            group: 'Group 1',
            name: 'Name 3',
            value: 87,
            fill: 'var(--oc-cyan-9)',
            linkURL: '',
        },
        {
            group: 'Group 2',
            name: 'Name 1',
            value: 168,
            fill: 'var(--oc-yellow-9)',
            linkURL: '',
        },
        {
            group: 'Group 2',
            name: 'Name 2',
            value: 130,
            fill: 'var(--oc-grape-9)',
            linkURL: '',
        },
        {
            group: 'Group 2',
            name: 'Name 3',
            value: 118,
            fill: 'var(--oc-cyan-9)',
            linkURL: '',
        },
        {
            group: 'Group 3',
            name: 'Name 1',
            value: 125,
            fill: 'var(--oc-yellow-9)',
            linkURL: '',
        },
        {
            group: 'Group 3',
            name: 'Name 2',
            value: 100,
            fill: 'var(--oc-grape-9)',
            linkURL: '',
        },
        {
            group: 'Group 3',
            name: 'Name 3',
            value: 157,
            fill: 'var(--oc-cyan-9)',
            linkURL: '',
        },
        {
            group: 'Group 4',
            name: 'Name 1',
            value: 60,
            fill: 'var(--oc-yellow-9)',
            linkURL: '',
        },
        {
            group: 'Group 4',
            name: 'Name 2',
            value: 114,
            fill: 'var(--oc-grape-9)',
            linkURL: '',
        },
        {
            group: 'Group 4',
            name: 'Name 3',
            value: 191,
            fill: 'var(--oc-cyan-9)',
            linkURL: '',
        },
    ],
}
