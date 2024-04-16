<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import type { Progress, Skipped } from '$lib/shared'

    import ResultsIndicator from './ResultsIndicator.svelte'

    export const meta = {
        component: ResultsIndicator,
    }
</script>

<script lang="ts">
    const SEARCH_JOB_THRESHOLD = 10000
    const states: {
        stateName: string
        state: 'error' | 'loading' | 'complete'
        progress: Progress
        severity: 'info' | 'warn' | 'error'
    }[] = [
        {
            stateName: 'done',
            state: 'complete',
            severity: 'info',
            progress: {
                done: true,
                matchCount: 10200,
                durationMs: 2700,
                skipped: [
                    {
                        reason: 'shard-match-limit',
                        title: 'Display limit hit',
                        message: 'Nope',
                        severity: 'info',
                        suggested: {
                            title: 'increase count',
                            queryExpression: 'count:11000',
                        },
                    },
                ],
            },
        },
        {
            stateName: 'taking too long',
            state: 'loading',
            severity: 'info',
            progress: {
                done: false,
                matchCount: 4549,
                durationMs: 12000,
                skipped: [
                    {
                        reason: 'repository-archive',
                        title: '96 archived',
                        message: 'add forked:yes',
                        severity: 'info',
                        suggested: {
                            title: 'adjust query with',
                            queryExpression: 'forked:yes',
                        },
                    },
                ],
            },
        },
        {
            stateName: 'error',
            state: 'complete',
            severity: 'error',
            progress: {
                done: true,
                matchCount: 2364,
                durationMs: 19000,
                skipped: [
                    {
                        reason: 'shard-timedout',
                        title: 'Internal server error',
                        message: 'There was an error',
                        severity: 'error',
                    },
                ],
            },
        },
        {
            stateName: 'mostSevere: forked',
            state: 'complete',
            severity: 'info',
            progress: {
                done: true,
                matchCount: 10200,
                durationMs: 2700,
                skipped: [
                    {
                        reason: 'repository-fork',
                        title: '96 forked',
                        message: 'add forked:yes',
                        severity: 'info',
                        suggested: {
                            title: 'adjust query with',
                            queryExpression: 'forked:yes',
                        },
                    },
                ],
            },
        },
        {
            stateName: 'mostSevere: archived',
            state: 'complete',
            severity: 'info',
            progress: {
                done: true,
                matchCount: 4549,
                durationMs: 738,
                skipped: [
                    {
                        reason: 'repository-archive',
                        title: '96 archived',
                        message: 'add forked:yes',
                        severity: 'info',
                        suggested: {
                            title: 'adjust query with',
                            queryExpression: 'forked:yes',
                        },
                    },
                ],
            },
        },
        {
            stateName: 'loading',
            state: 'loading',
            severity: 'info',
            progress: {
                done: false,
                matchCount: 230,
                durationMs: 6400,
                skipped: new Array<Skipped>(),
            },
        },
    ]
</script>

<Story name="Default">
    <section>
        {#each states as { stateName, state, progress, suggestedItems, severity }}
            <h4>Results Indicator in {stateName} state</h4>
            <div>
                <ResultsIndicator {state} {progress} {suggestedItems} {severity} />
            </div>
        {/each}
    </section>
</Story>

<style lang="scss">
    div {
        border: 1px solid var(--border-color);
        width: fit-content;
        padding: 0.5rem;
        border-radius: 4px;
        margin-bottom: 3rem;
    }
</style>
