<script lang="ts">
    import { SourcegraphURL } from '$lib/common'
    import type { InfinityQueryStore } from '$lib/graphql'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Scroller from '$lib/Scroller.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Alert } from '$lib/wildcard'
    import Panel from '$lib/wildcard/resizable-panel/Panel.svelte'
    import PanelGroup from '$lib/wildcard/resizable-panel/PanelGroup.svelte'
    import PanelResizeHandle from '$lib/wildcard/resizable-panel/PanelResizeHandle.svelte'

    import FilePreview from './FilePreview.svelte'
    import type { ReferencePanel_LocationConnection, ReferencePanel_Location } from './ReferencePanel.gql'
    import ReferencePanelCodeExcerpt from './ReferencePanelCodeExcerpt.svelte'

    export let references: InfinityQueryStore<ReferencePanel_LocationConnection['nodes']>

    // It appears that the backend returns duplicate locations. We need to filter them out.
    function unique(locations: ReferencePanel_Location[]): ReferencePanel_Location[] {
        const seen = new Set<string>()
        return locations.filter(location => {
            const key = location.canonicalURL
            if (seen.has(key)) {
                return false
            }
            seen.add(key)
            return true
        })
    }

    function getPreviewURL(location: ReferencePanel_Location) {
        const url = SourcegraphURL.from(location.canonicalURL)
        if (location.range) {
            url.setLineRange({
                line: location.range.start.line + 1,
                character: location.range.start.character + 1,
            })
        }
        return url.toString()
    }

    let selectedLocation: ReferencePanel_Location | null = null

    $: previewURL = selectedLocation ? getPreviewURL(selectedLocation) : null
    $: locations = $references.data ? unique($references.data) : []
</script>

<div class="root">
    <PanelGroup id="references">
        <Panel id="references-list">
            <Scroller margin={600} on:more={references.fetchMore}>
                {#if !$references.fetching && !$references.error && locations.length === 0}
                    <div class="info">
                        <Alert variant="info">No references found.</Alert>
                    </div>
                {/if}
                <ul>
                    {#each locations as location (location.canonicalURL)}
                        {@const selected = selectedLocation?.canonicalURL === location.canonicalURL}
                        <!-- todo(fkling): Implement a11y concepts. What to do exactly depends on whether
                             we'll keep the preview panel or not. -->
                        <li
                            class="location"
                            class:selected
                            on:click={() => (selectedLocation = selected ? null : location)}
                        >
                            <span class="code-file">
                                <span class="code">
                                    <ReferencePanelCodeExcerpt {location} />
                                </span>
                                <span class="file">
                                    <Tooltip tooltip={location.resource.path}>
                                        <span>{location.resource.name}</span>
                                    </Tooltip>
                                </span>
                            </span>
                            {#if location.range}
                                <span class="range"
                                    >:{location.range.start.line + 1}:{location.range.start.character + 1}</span
                                >
                            {/if}
                        </li>
                    {/each}
                </ul>
                {#if $references.fetching}
                    <div class="loader"><LoadingSpinner center /></div>
                {:else if $references.error}
                    <div class="loader">
                        <Alert variant="danger">Unable to load references: {$references.error.message}</Alert>
                    </div>
                {/if}
            </Scroller>
        </Panel>
        {#if previewURL}
            <PanelResizeHandle />
            <Panel defaultSize={50} id="reference-panel-preview">
                <FilePreview href={previewURL} on:close={() => (selectedLocation = null)} />
            </Panel>
        {/if}
    </PanelGroup>
</div>

<style lang="scss">
    .root {
        height: 100%;

        :global([data-panel-id='reference-panel-preview']) {
            z-index: 0;
            position: relative;
        }
    }

    ul {
        margin: 0;
        padding: 0;
        display: grid;
        grid-template-columns: 1fr max-content;
    }

    li {
        display: grid;
        grid-column: span 2;
        grid-template-columns: subgrid;
        color: inherit;
        align-items: center;
        padding: 0.25rem;
        cursor: pointer;

        &:hover {
            text-decoration: none;
            background-color: var(--color-bg-2);
        }

        &.selected {
            background-color: var(--color-bg-2);
        }
    }

    ul:not(:empty) + .loader,
    li + li {
        border-top: 1px solid var(--border-color);
    }

    .code-file {
        display: flex;
        align-items: center;
        min-width: 0;
        gap: 0.5rem;
    }

    .code {
        flex: 1;
        text-overflow: ellipsis;
        overflow: hidden;
    }

    .file {
        text-align: right;
        color: var(--text-muted);
    }

    .range {
        color: var(--oc-violet-6);
        text-align: left;
    }

    .loader {
        padding: 1rem;
    }
</style>
