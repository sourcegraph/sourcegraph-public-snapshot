<script context="module" lang="ts">
function getTelemetrySourceClient(): string {
	if (window.context?.sourcegraphDotComMode) {
		return "dotcom.web";
	}
	return "server.web";
}
</script>

<script lang="ts">
    import { createElement } from 'react'

    import { createRoot, type Root } from 'react-dom/client'
    import { onDestroy } from 'svelte'

    import { CodyWebPanel, CodyWebPanelProvider } from '@sourcegraph/cody-web'

    import type { CodySidebar_ResolvedRevision } from './CodySidebar.gql'

    import '@sourcegraph/cody-web/dist/style.css'

    import type { LineOrPositionOrRange } from '@sourcegraph/common'

    export let repository: CodySidebar_ResolvedRevision | undefined = undefined
    export let filePath: string | undefined = undefined
    export let lineOrPosition: LineOrPositionOrRange | undefined = undefined

    let container: HTMLDivElement
    let root: Root | null

    $: if (container) {
        render(repository, filePath, lineOrPosition)
    }

    onDestroy(() => {
        root?.unmount()
        root = null
    })

    function render(
        repository?: CodySidebar_ResolvedRevision,
        filePath?: string,
        lineOrPosition?: LineOrPositionOrRange
    ) {
        if (!root) {
            root = createRoot(container)
        }

        const chat = createElement(CodyWebPanel)
        const hasFileRangeSelection = lineOrPosition?.line

        const provider = createElement(
            CodyWebPanelProvider,
            {
                accessToken: '',
                initialContext: {
                    repositories: repository ? [repository] : [],
                    fileURL: filePath ? (!filePath.startsWith('/') ? `/${filePath}` : filePath) : undefined,
                    // Line range - 1 because of Cody Web initial context file range bug
                    fileRange: hasFileRangeSelection
                        ? {
                              startLine: lineOrPosition.line - 1,
                              endLine: lineOrPosition.endLine ? lineOrPosition.endLine - 1 : lineOrPosition.line - 1,
                          }
                        : undefined,
                },
                serverEndpoint: window.location.origin,
                customHeaders: window.context.xhrHeaders,
                telemetryClientName: getTelemetrySourceClient(),
            },
            [chat]
        )
        root.render(provider)
    }
</script>

<div class="chat" bind:this={container} />

<style lang="scss">
    .chat {
        --vscode-sideBar-background: var(--body-bg);
        --vscode-editor-background: var(--body-bg);
        --vscode-editor-foreground: var(--body-color);
        --vscode-input-background: var(--input-bg);
        --vscode-input-foreground: var(--body-color);
        --vscode-textLink-foreground: var(--primary);
        --vscode-input-border: var(--border-color-2);
        --vscode-inputOption-activeBackground: var(--search-input-token-filter);
        --vscode-inputOption-activeForeground: var(--body-color);
        --vscode-loading-dot-color: var(--body-color);
        --vscode-textPreformat-foreground: var(--body-color);
        --vscode-textPreformat-background: var(--secondary);
        --vscode-sideBarSectionHeader-border: var(--border-color);
        --vscode-editor-font-family: var(--code-font-family);
        --vscode-editor-font-size: var(--code-font-size);
        --mention-color-opacity: 100%;

        // LLM picker tokens
        --vscode-quickInput-background: var(--body-bg);
        --vscode-dropdown-border: var(--border-color);
        --vscode-dropdown-foreground: var(--body-color);
        --vscode-foreground: var(--body-color);
        --vscode-list-activeSelectionBackground: #e8f7ff;
        --vscode-list-activeSelectionForeground: var(--primary);
        --vscode-input-placeholderForeground: var(--body-color);
        --vscode-button-foreground: var(--body-color);
        --vscode-keybindingLabel-background: transparent;
        --vscode-keybindingLabel-foreground: var(--body-color);

        line-height: 1.55;
        flex: 1;
        min-height: 0;

        :global(button) {
            opacity: 1;
        }

        :global(h3),
        :global(h4) {
            font-size: inherit;
            margin: 0;
        }

        :global(ul) {
            margin: 0;
        }

        :global(code) {
            padding: 1px 3px;
            border-radius: 0.25rem;
            color: var(--vscode-textPreformat-foreground);
            background-color: var(--vscode-textPreformat-background);
        }

        :global(pre) {
            // Controls cody snippets (i.e. 'pre code' blocks)
            --code-foreground: var(--body-color);
            --code-background: transparent;

            border-top-right-radius: 2px;
            border-top-left-radius: 2px;

            :global(code) {
                // Overwrite the code styles set above
                padding: initial;
                background-color: inherit;
            }
        }

        // Sourcegraph styles already add [hidden] display none
        // and this breaks chat animation since there is no starting point
        // with display:none element. Override this logic back to visibility: hidden;
        // so chat animation would work again.
        :global([hidden]) {
            visibility: hidden;
            display: block !important;
        }

        // Target all possible animated elements (radix accordions)
        // and disable animation since there are flashes with exit
        // animations.
        :global(.tw-transition-all) {
            animation: none !important;
        }

        :global([cmdk-root] input:focus-visible) {
            box-shadow: unset !important;
        }

        // As of Cody Web 0.4.0 the buttons implemented in Cody does not satisfy
        // the design requirements. Hence here we are overriding the button
        // styles here to fix them.
        :global(button:has(h4)),
        :global([cmdk-root] a) {
            background-color: transparent;
            color: var(--body-color);
            padding: 2px 4px;

            &:hover {
                color: #181b26;
                background-color: #eff2f5;
            }

            &:active {
                color: #1c7ed6;
                background-color: #e8f7ff;
            }

            &:disabled {
                color: #798baf;
                background-color: #798baf;
            }

            &:focus {
                color: #181b26;
                background: transparent;
                box-shadow: 0 0 0 2px #a3d0ff;
            }
        }
        :global(.theme-dark) & {
            --vscode-list-activeSelectionBackground: #031824;

            // As of Cody Web 0.4.0 the buttons implemented in Cody does not satisfy
            // the design requirements. Hence here we are overriding the button
            // styles here to fix them.
            :global(button:has(h4)),
            :global([cmdk-root] a) {
                &:hover {
                    color: #e6ebf2;
                    background-color: #14171f;
                }

                &:active {
                    color: #1c7ed6;
                    background-color: #031824;
                }

                &:disabled {
                    color: #5e6e8c;
                    background-color: #0f111a;
                }

                &:focus {
                    color: #e6ebf2;
                    background: transparent;
                    box-shadow: 0 0 0 2px #0b4c90;
                }
            }
        }
    }

    :global([data-floating-ui-portal]) {
        --vscode-quickInput-background: var(--body-bg);
        --vscode-widget-border: var(--border-color);
        --vscode-list-activeSelectionBackground: var(--primary);
        --vscode-foreground: var(--body-color);
        --vscode-widget-shadow: rgba(36, 41, 54, 0.2);
        // Turn off background color for picker popover element
        // Which causes glitch effect in Cody Web
        --vscode-sideBar-background: transparent;
    }

    :global([cmdk-root]) {
        --vscode-list-activeSelectionBackground: #e8f7ff;
        --vscode-list-activeSelectionForeground: var(--primary);

        :global(.theme-dark) & {
            --vscode-list-activeSelectionBackground: #031824;
        }
    }

    :global([data-cody-web-chat]) {
        height: 100%;
        overflow: auto;
        background-color: var(--vscode-editor-background);
        font-size: var(--vscode-font-size);
        font-family: var(--vscode-font-family);
        color: var(--vscode-editor-foreground);
        padding-bottom: 2rem;
    }
</style>
