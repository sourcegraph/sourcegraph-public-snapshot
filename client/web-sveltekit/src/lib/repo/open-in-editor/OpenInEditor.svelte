<script lang="ts">
    import type {Settings} from '$root/client/shared/src/schema/settings.schema'
    import {buildEditorUrl, buildRepoBaseNameAndPath} from './build-url';
    import {parseBrowserRepoURL} from '$lib/utils/url';
    import type {EditorSettings} from './editor-settings';
    import {getEditor} from './editors';
    import {getEditorSettingsErrorMessage} from './build-url';
    import Tooltip from '$lib/Tooltip.svelte';
    import EditorIcon from '$lib/repo/open-in-editor/EditorIcon.svelte';

    export let settings: Settings;
    export let externalServiceType: string = "";
    export let sourcegraphURL: string;

    const editorSettingsErrorMessage = getEditorSettingsErrorMessage(
        settings?.openInEditor,
        sourcegraphURL
    )
    const editorIds = (settings?.openInEditor as EditorSettings | undefined)?.editorIds ?? []
    const editors = !editorSettingsErrorMessage ? editorIds.map(getEditor) : undefined

    const {repoName, filePath, position, range} = parseBrowserRepoURL(window.location.href)
    const start = position || range?.start;
</script>

{#if editors}
    {#each editors as e, i}
        {#if e}
            <Tooltip tooltip={`Open in ${e.name}`}>
                <a href={buildEditorUrl(
                        buildRepoBaseNameAndPath(repoName, externalServiceType, filePath),
                        start,
                        settings.openInEditor,
                        sourcegraphURL,
                        i
                    ).toString()}
                   target="_blank" rel="noopener noreferrer">
                <EditorIcon editor={e} />
                    <span data-action-label>
                        Editor
                    </span>
                </a>
            </Tooltip>
        {/if}
    {/each}
{:else if editorSettingsErrorMessage}
    <Tooltip tooltip={editorSettingsErrorMessage}>
        <span data-action-label>
            Editor
        </span>
    </Tooltip>
{/if}

<style lang="scss">
    a {
        color: var(--body-color);
        text-decoration: none;
        white-space: nowrap;
    }
</style>
