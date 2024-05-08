<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import type { Progress, Skipped } from '$lib/shared'
    import Button from '$lib/wildcard/Button.svelte'

    import ResultsIndicator from './ResultsIndicator.svelte'

    export const meta = {
        component: ResultsIndicator,
    }
</script>

<script lang="ts">
    const states: {
        stateName: string
        state: 'error' | 'loading' | 'complete'
        progress: Progress
        suggestedItems: Required<Skipped>[]
        severity: 'info' | 'warn' | 'error'
    }[] = [
        {
            stateName: 'should display "loading" state',
            state: 'loading',
            severity: 'info',
            progress: {
                done: false,
                matchCount: 230,
                durationMs: 6400,
                skipped: new Array<Skipped>(),
            },
            suggestedItems: [],
        },
        {
            stateName: 'should display done state',
            state: 'complete',
            severity: 'info',
            progress: {
                done: true,
                matchCount: 10200,
                durationMs: 2700,
                skipped: [],
            },
            suggestedItems: [],
        },
        {
            stateName: 'should not display forked messaging',
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
            suggestedItems: [],
        },
        {
            stateName: 'should not display archived messaging',
            state: 'complete',
            severity: 'info',
            progress: {
                done: true,
                matchCount: 4549,
                durationMs: 738,
                skipped: [
                    {
                        reason: 'excluded-archive',
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
            suggestedItems: [],
        },
        {
            stateName: 'should display "Error" state',
            state: 'error',
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
            suggestedItems: [],
        },
        {
            stateName: 'should display "Error with suggested" state',
            state: 'error',
            severity: 'error',
            progress: {
                done: true,
                matchCount: 2364,
                durationMs: 19000,
                skipped: [
                    {
                        reason: 'shard-timedout',
                        title: 'NOT FOUND',
                        message: 'There was an error',
                        severity: 'error',
                        suggested: {
                            title: 'There was an error',
                            queryExpression: '404',
                        },
                    },
                ],
            },
            suggestedItems: [],
        },
        {
            stateName: 'should display "done with suggested" state',
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
                    {
                        reason: 'excluded-archive',
                        title: '96 archived',
                        message: 'add forked:yes',
                        severity: 'info',
                        suggested: {
                            title: 'adjust query with',
                            queryExpression: 'forked:yes',
                        },
                    },
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
            suggestedItems: [],
        },
        {
            stateName: 'should display "taking to long" state',
            state: 'loading',
            severity: 'info',
            progress: {
                done: false,
                matchCount: 4549,
                durationMs: 12000,
                skipped: [
                    {
                        reason: 'excluded-archive',
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
            suggestedItems: [],
        },
    ]
</script>

<Story name="Default">
    <h1>ResultsIndicator.svelte</h1>
    <section>
        {#each states as { stateName, state, progress, suggestedItems, severity }}
            <h4>It {stateName}</h4>
            <Button variant={state === 'error' ? 'danger' : 'secondary'} size="sm" outline>
                <svelte:fragment slot="custom" let:buttonClass>
                    <button class="{buttonClass} progress-button">
                        <ResultsIndicator {state} {suggestedItems} {progress} {severity} />
                    </button>
                </svelte:fragment>
            </Button>
            <br />
            <br />
        {/each}
    </section>
</Story>

<style lang="scss">
</style>
