<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import type { Progress, Skipped } from '$lib/shared'

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
            <div class="scene">
                <h4>It {stateName}</h4>
                <div>
                    <ResultsIndicator {state} {progress} {suggestedItems} {severity} />
                </div>
            </div>
        {/each}
    </section>
</Story>

<style lang="scss">
    section {
        display: flex;
        flex-flow: column wrap;
        max-height: 65vh;
        max-width: 100vw;
        gap: 0.5rem 0.5rem;
    }
    .scene {
        div {
            border: 1px solid var(--border-color);
            width: fit-content;
            padding: 0.5rem;
            border-radius: 4px;
            margin-bottom: 3rem;
        }
    }
</style>
