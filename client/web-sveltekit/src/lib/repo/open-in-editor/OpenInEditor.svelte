<script lang="ts">
    import { getEditor, parseBrowserRepoURL, buildRepoBaseNameAndPath, buildEditorUrl } from '$lib/web'
    import { getEditorSettingsErrorMessage } from './build-url'
    import Tooltip from '$lib/Tooltip.svelte'
    import EditorIcon from '$lib/repo/open-in-editor/EditorIcon.svelte'
    import { settings } from '$lib/stores'
    import { page } from '$app/stores'
    import type { ExternalRepository } from '$lib/graphql-types'
    import DefaultEditorIcon from '$lib/repo/open-in-editor/DefaultEditorIcon.svelte'

    export let externalServiceType: ExternalRepository['serviceType'] = ''

    $: openInEditor = $settings?.openInEditor

    $: editorSettingsErrorMessage = getEditorSettingsErrorMessage(openInEditor)
    $: editorIds = openInEditor?.editorIds ?? []
    $: editors = !editorSettingsErrorMessage ? editorIds.map(getEditor) : undefined

    $: sourcegraphBaseURL = new URL($page.url).origin

    $: ({ repoName, filePath, position, range } = parseBrowserRepoURL($page.url.toString()))
    $: start = position ?? range?.start
</script>

{#if editors}
    {#each editors as editor, editorIndex}
        {#if editor}
            <Tooltip tooltip={`Open in ${editor.name}`}>
                <a
                    href={buildEditorUrl(
                        buildRepoBaseNameAndPath(repoName, externalServiceType, filePath),
                        start,
                        openInEditor,
                        sourcegraphBaseURL,
                        editorIndex
                    ).toString()}
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    <EditorIcon editorId={editor.id} />
                    <span data-action-label> Editor </span>
                </a>
            </Tooltip>
        {/if}
    {/each}
{:else if editorSettingsErrorMessage}
    <Tooltip tooltip={editorSettingsErrorMessage}>
        <a href="/help/integration/open_in_editor" target="_blank">
            <DefaultEditorIcon />
            <span data-action-label> Editor </span>
        </a>
    </Tooltip>
{/if}

<style lang="scss">
    a {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        color: var(--text-body);
        text-decoration: none;
        white-space: nowrap;

        &:hover {
            color: var(--text-title);
        }
    }
</style>
