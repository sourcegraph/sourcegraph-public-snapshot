export const noRolloutWindowMockResult = {
    data: {
        site: {
            configuration: {
                effectiveContents: '{}',
            },
        },
    },
}

export const rolloutWindowConfigMockResult = {
    data: {
        site: {
            configuration: {
                effectiveContents:
                    '{"batchChanges.rolloutWindows":[{"rate":"unlimited"},{"rate":"3/hour","days":["monday","wednesday","thursday"]},{"rate":"0/hour","days":["friday"],"start":"08:00","end":"20:00"}]}',
            },
        },
    },
}
