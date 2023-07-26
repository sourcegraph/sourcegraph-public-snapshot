import type { Meta, StoryObj } from '@storybook/svelte'

import HistoryPanel from '$lib/repo/HistoryPanel.svelte'
import { createHistoryResults } from '$testdata'

import ForceUpdate from './ForceUpdate.svelte'

const meta: Meta<typeof ForceUpdate> = {
    title: 'stories/HistoryPanel',
    parameters: {
        fullscreen: true,
    },
}

export default meta

export const Default: StoryObj<{ commitCount: number }> = {
    render: args => {
        const [initial, next] = createHistoryResults(2, args.commitCount)

        return {
            Component: ForceUpdate,
            props: {
                component: HistoryPanel,
                history: Promise.resolve(initial),
                fetchMoreHandler: async () => next,
            },
        }
    },
    args: {
        commitCount: 5,
    },
}
