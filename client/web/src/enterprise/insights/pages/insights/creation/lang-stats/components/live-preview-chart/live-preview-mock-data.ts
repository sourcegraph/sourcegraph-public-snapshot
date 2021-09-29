import { random } from 'lodash'
import { PieChartContent } from 'sourcegraph'

export const DEFAULT_PREVIEW_MOCK: PieChartContent<any> = {
    chart: 'pie' as const,
    pies: [
        {
            dataKey: 'value',
            nameKey: 'name',
            fillKey: 'fill',
            linkURLKey: 'linkURL',
            data: [
                {
                    name: 'Covered',
                    value: 0.3,
                    fill: 'var(--oc-grape-7)',
                    linkURL: '#Covered',
                },
                {
                    name: 'Not covered',
                    value: 0.7,
                    fill: 'var(--oc-orange-7)',
                    linkURL: '#Not_covered',
                },
            ],
        },
    ],
}

export function getRandomLangStatsMock(): PieChartContent<any> {
    const randomFirstPieValue = random(0, 0.6)
    const randomSecondPieValue = 1 - randomFirstPieValue

    return {
        chart: 'pie' as const,
        pies: [
            {
                dataKey: 'value',
                nameKey: 'name',
                fillKey: 'fill',
                linkURLKey: 'linkURL',
                data: [
                    {
                        name: 'JavaScript',
                        value: randomFirstPieValue,
                        fill: 'var(--oc-grape-7)',
                        linkURL: '#Covered',
                    },
                    {
                        name: 'Typescript',
                        value: randomSecondPieValue,
                        fill: 'var(--oc-orange-7)',
                        linkURL: '#Not_covered',
                    },
                ],
            },
        ],
    }
}
