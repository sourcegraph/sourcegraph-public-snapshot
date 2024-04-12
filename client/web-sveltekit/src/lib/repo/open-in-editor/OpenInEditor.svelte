<script lang="ts">
    import { buildEditorUrl, buildRepoBaseNameAndPath } from './build-url'
    import { parseBrowserRepoURL } from '$lib/utils/url'
    import type { EditorSettings } from './editor-settings'
    import { getEditor } from './editors'
    import { getEditorSettingsErrorMessage } from './build-url'
    import Tooltip from '$lib/Tooltip.svelte'
    import EditorIcon from '$lib/repo/open-in-editor/EditorIcon.svelte'
    import { settings } from '$lib/stores'
    import { mdiCodeBraces } from '@mdi/js'
    import Icon from '$lib/Icon.svelte';

    export let externalServiceType: string = ''

    let openInEditor = $settings?.openInEditor

    const editorSettingsErrorMessage = getEditorSettingsErrorMessage(openInEditor)
    const editorIds = (openInEditor as EditorSettings | undefined)?.editorIds ?? []
    const editors = !editorSettingsErrorMessage ? editorIds.map(getEditor) : undefined

    const { repoName, filePath, position, range } = parseBrowserRepoURL(window.location.href)
    const start = position || range?.start
</script>

{#if editors}
    {#each editors as e, i}
        {#if e}
            <Tooltip tooltip={`Open in ${e.name}`}>
                <a
                    href={buildEditorUrl(
                        buildRepoBaseNameAndPath(repoName, externalServiceType, filePath),
                        start,
                        openInEditor,
                        i
                    ).toString()}
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    <EditorIcon editor={e} />
                    <span data-action-label> Editor </span>
                </a>
            </Tooltip>
        {/if}
    {/each}
{:else if editorSettingsErrorMessage}
    <Tooltip tooltip={editorSettingsErrorMessage}>
        <a href="https://sourcegraph.com/docs/integration/open_in_editor" target="_blank">
            <Icon aria-label="Editor" svgPath={mdiCodeBraces} inline />
            <span data-action-label> Editor </span>
        </a>
    </Tooltip>
{/if}

<style lang="scss">
    a {
        color: var(--body-color);
        text-decoration: none;
        white-space: nowrap;
    }
</style>
