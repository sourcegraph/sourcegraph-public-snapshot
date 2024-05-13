<script lang="ts">
    import {
        getEditor,
        parseBrowserRepoURL,
        buildRepoBaseNameAndPath,
        buildEditorUrl,
        isProjectPathValid
    } from '$lib/web'
    import {getEditorSettingsErrorMessage} from './build-url'
    import Tooltip from '$lib/Tooltip.svelte'
    import EditorIcon from '$lib/repo/open-in-editor/EditorIcon.svelte'
    import {settings} from '$lib/stores'
    import {page} from '$app/stores'
    import type {ExternalRepository} from '$lib/graphql-types'
    import DefaultEditorIcon from '$lib/repo/open-in-editor/DefaultEditorIcon.svelte'
    import Popover from '$lib/Popover.svelte';
    import {Button} from '$lib/wildcard';
    import {supportedEditors} from '$lib/web';
    import type {
        PageData
    } from '$root/client/web-sveltekit/.svelte-kit/types/src/routes/[...repo=reporev]/(validrev)/(code)/-/blob/[...path]/$types';

    export let externalServiceType: ExternalRepository['serviceType'] = ''
    export let data: Extract<PageData, { type: 'FileView' }>

    $: openInEditor = $settings?.openInEditor

    $: editorSettingsErrorMessage = getEditorSettingsErrorMessage(openInEditor)
    $: editorIds = openInEditor?.editorIds ?? []
    $: editors = !editorSettingsErrorMessage ? editorIds.map(getEditor) : undefined

    $: sourcegraphBaseURL = new URL($page.url).origin

    $: ({repoName, filePath, position, range} = parseBrowserRepoURL($page.url.toString()))
    $: start = position ?? range?.start

    $: latestSettings = data.subjects.at(-1)
    $: handleEditorUpdate = latestSettings?.latestSettings
        ? async (): Promise<void> => {
            if (latestSettings?.latestSettings) {
                await data.updateEditor(latestSettings.id, latestSettings.latestSettings.id, defaultProjectPath, selectedEditorId)
                window.location.reload()
            }
        }
        : undefined

    let defaultProjectPath = ''; // Assume initial state or fetch from a store
    let selectedEditorId = ''; // Assume initial state or fetch from a store
    $: areSettingsValid = !!selectedEditorId && isProjectPathValid(defaultProjectPath);

</script>

{#if editors}
    {#each editors as editor, editorIndex}
        {#if editor}
            <Tooltip tooltip={`Open in ${editor.name}`}>
                <a
                    class="action-href"
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
                    <EditorIcon editorId={editor.id}/>
                    <span data-action-label> Editor </span>
                </a>
            </Tooltip>
        {/if}
    {/each}
{:else if editorSettingsErrorMessage}
    <Popover let:registerTrigger let:toggle placement="left-start">
        <Tooltip tooltip="Set your preferred editor">
            <span use:registerTrigger on:click={() => toggle()}>
                <DefaultEditorIcon/>
                <span data-action-label> Editor </span>
            </span>
        </Tooltip>
        <div slot="content" class="open-in-editor-popover">
            <form on:submit={handleEditorUpdate} novalidate>
                <h3>Set your preferred editor</h3>
                <p>
                    Open this and other files directly in your editor. Set your path and editor to get started. Update
                    any time in your user settings.
                </p>
                <p class="form-label">Default projects path</p>
                <input
                    id="OpenInEditorForm-projectPath"
                    type="text"
                    name="projectPath"
                    placeholder="/Users/username/projects"
                    required
                    autocorrect="off"
                    autocapitalize="off"
                    spellcheck={false}
                    bind:value={defaultProjectPath}
                    class="form-input"
                />
                <p class="small form-info">
                    The directory that contains your repository checkouts. For example, if this repository is
                    checked out to <code>/Users/username/projects/cody</code>, then set your default projects path
                    to <code>/Users/username/projects</code>.
                </p>
                <p class="form-label editor-label">Editor</p>
                <select class="form-input" id="OpenInEditorForm-editor" bind:value={selectedEditorId}>
                    <option value=""></option>
                    {#each supportedEditors.sort((a, b) => a.name.localeCompare(b.name)).filter(editor => editor.id !== 'custom') as editor}
                        <option value={editor.id}>{editor.name}</option>
                    {/each}
                </select>
                <p class="small form-info">Use a different editor?{' '}
                    <a href="/help/integration/open_in_editor" target="_blank" rel="noreferrer noopener">Set up a
                        different editor</a>
                </p>
                <Button variant="primary" type="submit" disabled={!areSettingsValid}>
                    Save
                </Button>
            </form>
        </div>
    </Popover>
{/if}

<style lang="scss">
    .action-href {
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

    .open-in-editor-popover {
        isolation: isolate;
        width: 25rem;
        padding: 1.25rem 1rem;
        background-color: var(--color-bg-1);
    }

    .form-label {
        font-weight: 500;
    }

    .form-input {
        width: 100%;
        padding: 0.5rem;
        border-radius: 0.25rem;
        border: 1px solid var(--border-color);
    }

    .form-info {
        margin-top: 0.5rem;
    }

</style>
