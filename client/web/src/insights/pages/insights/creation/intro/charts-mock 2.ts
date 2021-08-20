import { LineChartContent, PieChartContent } from 'sourcegraph'

/** Mock data for search-based insight preview for create-intro page. */
export const LINE_CHART_DATA: LineChartContent<any, string> = {
    chart: 'line',
    data: [
        {
            date: 1597698000000,
            'Imports of old GQL.* types': 250,
            'Imports of new graphql-operations types': 61,
        },
        {
            date: 1601326800000,
            'Imports of old GQL.* types': 183,
            'Imports of new graphql-operations types': 207,
        },
        {
            date: 1604955600000,
            'Imports of old GQL.* types': 178,
            'Imports of new graphql-operations types': 242,
        },
        {
            date: 1608584400000,
            'Imports of old GQL.* types': 139,
            'Imports of new graphql-operations types': 329,
        },
        {
            date: 1612213200000,
            'Imports of old GQL.* types': 139,
            'Imports of new graphql-operations types': 341,
        },
        {
            date: 1615842000000,
            'Imports of old GQL.* types': 139,
            'Imports of new graphql-operations types': 360,
        },
        { date: 1619470800000, 'Imports of old GQL.* types': 130, 'Imports of new graphql-operations types': 396 },
    ],
    series: [
        {
            dataKey: 'Imports of old GQL.* types',
            name: 'Imports of old GQL.* types',
            stroke: 'var(--oc-red-7)',
        },
        {
            dataKey: 'Imports of new graphql-operations types',
            name: 'Imports of new graphql-operations types',
            stroke: 'var(--oc-blue-7)',
        },
    ],
    xAxis: { dataKey: 'date', type: 'number', scale: 'time' },
}

/** Mock data for lang-stats insight preview for create-intro page. */
export const PIE_CHART_DATA: PieChartContent<any> = {
    chart: 'pie',
    pies: [
        {
            data: [
                {
                    name: 'Go',
                    totalLines: 377693,
                    fill: '#00ADD8',
                },
                {
                    name: 'HTML',
                    totalLines: 224972,
                    fill: '#e34c26',
                },
                {
                    name: 'TypeScript',
                    totalLines: 165086,
                    fill: '#2b7489',
                },
                {
                    name: 'Markdown',
                    totalLines: 48327,
                    fill: '#083fa1',
                },
                {
                    name: 'YAML',
                    totalLines: 26305,
                    fill: '#cb171e',
                },
                {
                    name: 'Other',
                    totalLines: 59884,
                    fill: 'gray',
                },
            ],
            dataKey: 'totalLines',
            nameKey: 'name',
            fillKey: 'fill',
            linkURLKey: 'linkURL',
        },
    ],
}
