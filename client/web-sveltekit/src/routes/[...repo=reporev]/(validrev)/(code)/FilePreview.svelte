<script lang="ts">
    import { createEventDispatcher } from 'svelte'

    import { preloadData } from '$app/navigation'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { createPromiseStore } from '$lib/utils'
    import { Alert, Button } from '$lib/wildcard'

    import type { PageData } from './-/blob/[...path]/$types'
    import FileView from './-/blob/[...path]/FileView.svelte'

    /**
     * The URL of the file to preview.
     */
    export let href: string

    const dispatch = createEventDispatcher<{ close: void }>()
    const filePageData = createPromiseStore<PageData>()
    $: filePageData.set(
        preloadData(href).then(result => {
            if (result.type === 'loaded' && result.status === 200) {
                return result.data as PageData
            }
            throw new Error(`Unable to load file preview.`)
        })
    )
</script>

<div class:center={$filePageData.pending || $filePageData.error}>
    {#if $filePageData.pending}
        <LoadingSpinner />
    {:else if $filePageData.error}
        <Alert variant="danger">
            {$filePageData.error.message}
            <br />
            <a {href}>Open file directly</a>
        </Alert>
    {:else if $filePageData.value?.type === 'FileView'}
        <FileView data={$filePageData.value} embedded disableCodeIntel>
            <svelte:fragment slot="actions">
                <Button variant="icon" aria-label="Close preview" on:click={() => dispatch('close')}>
                    <Icon icon={ILucideX} aria-hidden inline />
                </Button>
            </svelte:fragment>
        </FileView>
    {/if}
</div>

<style lang="scss">
    div {
        display: flex;
        flex-direction: column;
        height: 100%;

        &.center {
            padding: 1rem;
            justify-content: center;
        }
    }
</style>
