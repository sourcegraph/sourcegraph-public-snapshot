<script lang="ts">
    import { page } from '$app/stores'
    import { getPlatform } from '$lib/common'
    import type { ExternalRepository, SettingsEdit } from '$lib/graphql-types'
    import Popover from '$lib/Popover.svelte'
    import DefaultEditorIcon from '$lib/repo/open-in-editor/DefaultEditorIcon.svelte'
    import EditorIcon from '$lib/repo/open-in-editor/EditorIcon.svelte'
    import { settings } from '$lib/stores'
    import Tooltip from '$lib/Tooltip.svelte'
    import {
        getEditor,
        parseBrowserRepoURL,
        buildRepoBaseNameAndPath,
        buildEditorUrl,
        isProjectPathValid,
        supportedEditors,
    } from '$lib/web'
    import { Button } from '$lib/wildcard'

    import { getEditorSettingsErrorMessage } from './build-url'

    export let externalServiceType: ExternalRepository['serviceType'] = ''
    export let updateUserSetting: (edit: SettingsEdit) => Promise<void>

    $: openInEditor = $settings?.openInEditor

    $: editorSettingsErrorMessage = getEditorSettingsErrorMessage(openInEditor)
    $: editorIds = openInEditor?.editorIds ?? []
    $: editors = !editorSettingsErrorMessage ? editorIds.map(getEditor) : undefined

    $: sourcegraphBaseURL = new URL($page.url).origin

    $: ({ repoName, filePath, position, range } = parseBrowserRepoURL($page.url.toString()))
    $: start = position ?? range?.start

    $: defaultProjectPath = ''
    $: selectedEditorId = undefined

    $: areSettingsValid = !!selectedEditorId && isProjectPathValid(defaultProjectPath)

    let isSaving = false
    $: handleEditorUpdate = async (): Promise<void> => {
        if (!selectedEditorId || !defaultProjectPath) {
            return
        }
        isSaving = true
        try {
            await updateUserSetting({
                value: defaultProjectPath,
                keyPath: [{ property: 'openInEditor' }, { property: 'projectPaths.default' }],
            })
            await updateUserSetting({
                value: [selectedEditorId],
                keyPath: [{ property: 'openInEditor' }, { property: 'editorIds' }],
            })

            openInEditor = {
                editorIds: [selectedEditorId],
                'projectPaths.default': defaultProjectPath,
            }
        } finally {
            isSaving = false
        }
    }

    function getSystemAwareProjectPathExample(suffix?: string) {
        switch (getPlatform()) {
            case 'windows':
                return 'C:\\Users\\username\\Projects' + (suffix ? `\\${suffix}` : '')
            case 'linux':
                return '/home/username/Projects' + (suffix ? `/${suffix}` : '')
            case 'mac':
            default:
                return '/Users/username/Projects' + (suffix ? `/${suffix}` : '')
        }
    }
</script>

{#if editors}
    {#each editors as editor, editorIndex}
        {#if editor}
            <Tooltip tooltip={`Open in ${editor.name}`}>
                <a
                    class="action"
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
    <Popover let:registerTrigger let:toggle placement="left-start">
        <Tooltip tooltip="Set your preferred editor">
            <button class="action" use:registerTrigger on:click={() => toggle()}>
                <DefaultEditorIcon />
                <span data-action-label>Editor</span>
            </button>
        </Tooltip>
        <div slot="content" class="open-in-editor-popover">
            <form on:submit={handleEditorUpdate} novalidate>
                <h3>Set your preferred editor</h3>
                <p>
                    Open this and other files directly in your editor. Set your path and editor to get started. Update
                    any time in your user settings.
                </p>
                <label>
                    Default projects path
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
                </label>

                <p class="small form-info">
                    The directory that contains your repository checkouts. For example, if this repository is checked
                    out to <code>{`${getSystemAwareProjectPathExample('cody')}`}</code>, then set your default projects
                    path to <code>{getSystemAwareProjectPathExample()}</code>.
                </p>
                <label>
                    Editor
                    <select class="form-input" id="OpenInEditorForm-editor" bind:value={selectedEditorId}>
                        <option value="" />
                        {#each supportedEditors
                            .sort((a, b) => a.name.localeCompare(b.name))
                            .filter(editor => editor.id !== 'custom') as editor}
                            <option value={editor.id}>{editor.name}</option>
                        {/each}
                    </select>
                </label>

                <p class="small form-info">
                    Use a different editor?{' '}
                    <a href="/help/integration/open_in_editor" target="_blank" rel="noreferrer noopener"
                        >Set up a different editor</a
                    >
                </p>
                <Button variant="primary" type="submit" disabled={!areSettingsValid || isSaving}>Save</Button>
            </form>
        </div>
    </Popover>
{/if}

<style lang="scss">
    .action {
        all: unset;

        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        color: var(--text-body);
        text-decoration: none;
        white-space: nowrap;
        cursor: pointer;

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
