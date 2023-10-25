<script lang="ts">
    import QueryExampleChip from '$lib/search/QueryExampleChip.svelte'
    import type { getQueryExamples } from '$lib/search/queryExamples'
    import TabPanel from '$lib/TabPanel.svelte'
    import Tabs from '$lib/Tabs.svelte'

    export let examples: ReturnType<typeof getQueryExamples>
</script>

{#if examples.length > 1}
    <Tabs>
        {#each examples as panel (panel.title)}
            <TabPanel title={panel.title}>
                <div>
                    {#each panel.columns as groups}
                        <ul>
                            {#each groups as group (group.title)}
                                <li>
                                    <h2>{group.title}</h2>
                                    <ul>
                                        {#each group.queryExamples as example}
                                            <li><QueryExampleChip queryExample={example} /></li>
                                        {/each}
                                    </ul>
                                </li>
                            {/each}
                        </ul>
                    {/each}
                </div>
            </TabPanel>
        {/each}
    </Tabs>
{:else}
    {#each examples[0].columns as column}
        <ul>
            {#each column as group (group.title)}
                <li>
                    <h2>{group.title}</h2>
                    <ul>
                        {#each group.queryExamples as example}
                            <li><QueryExampleChip queryExample={example} /></li>
                        {/each}
                    </ul>
                </li>
            {/each}
        </ul>
    {/each}
{/if}

<style lang="scss">
    @import '$lib/breakpoints';
    div {
        display: flex;
        gap: 4rem;

        @media (--xs-breakpoint-down) {
            flex-direction: column;
            gap: 0;
        }
    }

    h2 {
        margin-top: 1rem;
        margin-bottom: 0.75rem;
        color: var(--text-muted);
        font-size: var(--font-size-base);
        font-weight: 400;
    }

    ul {
        margin: 0;
        padding: 0;
        list-style: none;

        li {
            margin: 0.5rem 0;
        }
    }
</style>
