<script lang="ts">
    import { CodyWebChat, CodyWebChatProvider } from 'cody-web-experimental'
    import { createElement } from 'react'
    import { createRoot, type Root } from 'react-dom/client'
    import { onDestroy } from 'svelte'
    import type { CodySidebar_ResolvedRevision } from './CodySidebar.gql'

    import 'cody-web-experimental/dist/style.css'
    import { createLocalWritable } from '$lib/stores'

    export let repository: CodySidebar_ResolvedRevision
    export let filePath: string

    const chatIDs = createLocalWritable<Record<string, string>>('cody.context-to-chat-ids', {})
    let container: HTMLDivElement
    let root: Root | null

    $: if (container) {
        render(repository, filePath)
    }

    onDestroy(() => {
        root?.unmount()
        root = null
    })

    function render(repository: CodySidebar_ResolvedRevision, filePath: string) {
        if (!root) {
            root = createRoot(container)
        }
        const chat = createElement(CodyWebChat)
        const provider = createElement(
            CodyWebChatProvider,
            {
                accessToken: '',
                chatID: $chatIDs[`${repository.id}-${filePath}`] ?? null,
                initialContext: {
                    repositories: [repository],
                    fileURL: filePath ? (!filePath.startsWith('/') ? `/${filePath}` : filePath) : undefined,
                },
                serverEndpoint: window.location.origin,
                onNewChatCreated: (chatID: string) => {
                    chatIDs.update(ids => {
                        ids[`${repository.id}-${filePath}`] = chatID
                        return ids
                    })
                },
            },
            [chat]
        )
        root.render(provider)
    }
</script>

<div class="chat" bind:this={container} />

<style lang="scss">
    .chat {
        --vscode-editor-background: var(--body-bg);
        --vscode-editor-foreground: var(--body-color);
        --vscode-input-background: var(--input-bg);
        --vscode-input-foreground: var(--body-color);
        --vscode-textLink-foreground: var(--primary);
        --vscode-input-border: var(--border-color-2);
        --vscode-inputOption-activeBackground: var(--search-input-token-filter);
        --vscode-inputOption-activeForeground: var(--body-color);
        --vscode-loading-dot-color: var(--body-color);
        --mention-color-opacity: 100%;

        height: 100%;

        :global(h3) {
            font-size: inherit;
            margin: 0;
        }

        :global(ul) {
            margin: 0;
        }

        :global(a) {
            color: var(--link-color) !important;
        }
    }

    :global([data-floating-ui-portal]) {
        --vscode-quickInput-background: var(--secondary-2);
        --vscode-widget-border: var(--border-color);
        --vscode-list-activeSelectionBackground: var(--primary);
        --vscode-foreground: var(--body-color);
        --vscode-widget-shadow: rgba(36, 41, 54, 0.2);
    }
</style>
