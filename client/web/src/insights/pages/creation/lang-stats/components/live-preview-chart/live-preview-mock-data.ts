export const DEFAULT_PREVIEW_MOCK = {
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
