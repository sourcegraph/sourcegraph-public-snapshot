<script lang="ts" context="module">
    export enum FuzzyFinderTabType {
        All = 'all',
        Repos = 'repos',
        Symbols = 'symbols',
        Files = 'files',
    }

    export type FuzzyFinderTabId = FuzzyFinderTabType | `${FuzzyFinderTabType}`
</script>

<script lang="ts">
    import { tick } from 'svelte'

    import { isMacPlatform } from '@sourcegraph/common'

    import { dirname } from '$lib/common'
    import { nextSibling, onClickOutside, previousSibling } from '$lib/dom'
    import { getGraphQLClient } from '$lib/graphql'
    import Icon from '$lib/Icon.svelte'
    import FileIcon from '$lib/repo/FileIcon.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import EmphasizedLabel from '$lib/search/EmphasizedLabel.svelte'
    import SymbolKindIcon from '$lib/search/SymbolKindIcon.svelte'
    import { displayRepoName } from '$lib/shared'
    import TabsHeader, { type Tab } from '$lib/TabsHeader.svelte'
    import { Alert, Input } from '$lib/wildcard'
    import Button from '$lib/wildcard/Button.svelte'

    import { allHotkey, filesHotkey, reposHotkey, symbolsHotkey } from './keys'
    import { type CompletionSource, createFuzzyFinderSource } from './sources'
    import { isViewportMobile } from '$lib/stores'

    export let open = false
    export let scope = ''

    export function selectTab(tabID: FuzzyFinderTabId) {
        if (selectedTab.id !== tabID) {
            selectedOption = 0
            selectedTab = tabs.find(t => t.id === tabID) ?? tabs[0]
        }
    }

    const client = getGraphQLClient()
    const tabs: (Tab & { source: CompletionSource })[] = [
        {
            id: 'all',
            title: 'All',
            shortcut: allHotkey,
            source: createFuzzyFinderSource({
                client,
                queryBuilder: value =>
                    `patterntype:keyword (type:repo OR type:path OR type:symbol) count:50 ${scope} ${value}`,
            }),
        },
        {
            id: 'repos',
            title: 'Repos',
            shortcut: reposHotkey,
            source: createFuzzyFinderSource({
                client,
                queryBuilder: value => `patterntype:keyword type:repo count:50 ${value}`,
            }),
        },
        {
            id: 'symbols',
            title: 'Symbols',
            shortcut: symbolsHotkey,
            source: createFuzzyFinderSource({
                client,
                queryBuilder: value => `patterntype:keyword type:symbol count:50 ${scope} ${value}`,
            }),
        },
        {
            id: 'files',
            title: 'Files',
            shortcut: filesHotkey,
            source: createFuzzyFinderSource({
                client,
                queryBuilder: value => `patterntype:keyword type:path count:50 ${scope} ${value}`,
            }),
        },
    ]

    let dialog: HTMLDialogElement | undefined
    let listbox: HTMLElement | undefined
    let input: HTMLInputElement | undefined
    let query = ''
    let selectedTab = tabs[0]
    let selectedOption: number = 0

    $: useScope = scope && selectedTab.id !== 'repos'
    $: source = selectedTab.source
    $: if (open) {
        source.next(query)
    }
    $: if (open) {
        dialog?.showModal()
        input?.select()
    } else {
        dialog?.close()
    }
    $: placeholder = (function () {
        switch (selectedTab.id) {
            case 'repos':
                return 'Find repositories...'
            case 'symbols':
                return 'Find symbols...'
            case 'files':
                return 'Find files...'
            default:
                return 'Find anything...'
        }
    })()

    function selectNext() {
        let next: HTMLElement | null = null
        const current = listbox?.querySelector('[aria-selected="true"]')
        if (current) {
            next = nextSibling(current, '[role="option"]', true) as HTMLElement | null
        } else {
            next = listbox?.querySelector('[role="option"]') as HTMLElement | null
        }

        if (next) {
            selectOption(next)
        }
    }

    function selectPrevious() {
        let prev: HTMLElement | null = null
        const current = listbox?.querySelector('[aria-selected="true"]')
        if (current) {
            prev = previousSibling(current, '[role="option"]', true) as HTMLElement | null
        } else {
            prev = listbox?.querySelector('[role="option"]:last-child') as HTMLElement | null
        }

        if (prev) {
            selectOption(prev)
        }
    }

    function selectOption(node: HTMLElement): void {
        if (node.dataset.index) {
            selectedOption = +node.dataset.index
            tick().then(() => node.scrollIntoView({ block: 'nearest' }))
        }
    }

    function handleKeyboardEvent(event: KeyboardEvent): void {
        switch (event.key) {
            // Select the next/first option
            case 'ArrowDown': {
                event.preventDefault()
                selectNext()
                break
            }
            // Select previous/last option
            case 'ArrowUp': {
                event.preventDefault()
                selectPrevious()
                break
            }
            // Select first option
            case 'Home': {
                event.preventDefault()
                const option = listbox?.querySelector('[role="option"]')
                if (option) {
                    selectedOption = 0
                    tick().then(() => option.scrollIntoView({ block: 'nearest' }))
                }
                break
            }
            // Select last option
            case 'End': {
                const options = listbox?.querySelectorAll('[role="option"]')
                if (options && options.length > 0) {
                    selectedOption = options.length - 1
                    tick().then(() => options[selectedOption].scrollIntoView({ block: 'nearest' }))
                }
                break
            }
            // Activate selected option
            case 'Enter': {
                event.preventDefault()
                const current = listbox?.querySelector('[aria-selected="true"]')
                if (current) {
                    current.querySelector('a')?.click()
                    dialog?.close()
                }
                break
            }
        }
    }

    function handleMacOSKeyboardEvent(event: KeyboardEvent): void {
        if (!event.ctrlKey) {
            return
        }

        switch (event.key) {
            case 'n': {
                event.preventDefault()
                selectNext()
                break
            }
            case 'p': {
                event.preventDefault()
                selectPrevious()
                break
            }
        }
    }

    function handleClick(event: MouseEvent) {
        const target = event.target as HTMLElement
        const option = target.closest('[role="option"]') as HTMLElement | null
        if (option?.dataset.index) {
            selectedOption = +option.dataset.index
            dialog?.close()
        }
    }
</script>

<dialog bind:this={dialog} on:close>
    <!-- We cannot use the `use:onClickOutside` directive on the dialog element itself because the element will take
         up the entire viewport and the event will never be triggered. -->
    <div class="content" use:onClickOutside on:click-outside={() => dialog?.close()}>
        <header>
            <TabsHeader
                id="fuzzy-finder"
                {tabs}
                selected={tabs.indexOf(selectedTab)}
                on:select={event => {
                    selectedTab = tabs[event.detail]
                    selectedOption = 0
                    input?.focus()
                }}
            />
            <span class="close">
                {#if $isViewportMobile}
                    <Button variant="secondary" on:click={() => dialog?.close()} size="lg" display="block">
                        Close
                    </Button>
                {:else}
                    <Button variant="icon" on:click={() => dialog?.close()} size="sm">
                        <Icon icon={ILucideX} aria-label="Close" />
                    </Button>
                {/if}
            </span>
        </header>
        <main>
            <div class="input">
                <Input
                    type="text"
                    bind:input
                    {placeholder}
                    autofocus
                    value={query}
                    onInput={event => {
                        selectedOption = 0
                        if (listbox) {
                            listbox.scrollTop = 0
                        }
                        query = event.currentTarget.value
                    }}
                    loading={$source.pending}
                    on:keydown={handleKeyboardEvent}
                    on:keydown={isMacPlatform() ? handleMacOSKeyboardEvent : undefined}
                />
                {#if useScope}
                    <div class="scope">Searching in <code>{scope}</code></div>
                {/if}
            </div>
            <ul role="listbox" bind:this={listbox} aria-label="Search results">
                {#if $source.pending}
                    <li class="message">Waiting for response...</li>
                {:else if $source.error}
                    <li class="error"><Alert variant="danger">{$source.error.message}</Alert></li>
                {:else if $source.value?.results}
                    {#each $source.value.results as item, index (item)}
                        {@const repo = item.repository.name}
                        {@const displayRepo = displayRepoName(repo)}
                        <li role="option" aria-selected={selectedOption === index} data-index={index}>
                            {#if item.type === 'repo'}
                                {@const matchOffset = repo.length - displayRepo.length}
                                <a href="/{item.repository.name}" on:click={handleClick}>
                                    <span class="icon"><CodeHostIcon repository={item.repository.name} /></span>
                                    <span class="label"
                                        ><EmphasizedLabel label={displayRepo} offset={matchOffset} /></span
                                    >
                                    <span class="info">{repo}</span>
                                </a>
                            {:else if item.type == 'symbol'}
                                <a href={item.symbol.location.url} on:click={handleClick}>
                                    <span class="icon"><SymbolKindIcon symbolKind={item.symbol.kind} /></span>
                                    <span class="label"><EmphasizedLabel label={item.symbol.name} /></span>
                                    <span class="info mono"
                                        >{#if !useScope}{displayRepo} &middot; {/if}{item.file.path}</span
                                    >
                                </a>
                            {:else if item.type == 'file'}
                                {@const fileName = item.file.name}
                                {@const folderName = dirname(item.file.path)}
                                <a href={item.file.url} on:click={handleClick}>
                                    <span class="icon"><FileIcon file={item.file} inline /></span>
                                    <span class="label"
                                        ><EmphasizedLabel label={fileName} offset={folderName.length + 1} /></span
                                    >
                                    <span class="info mono">
                                        {#if !useScope}{displayRepo} &middot; {/if}
                                        <EmphasizedLabel label={folderName} />
                                    </span>
                                </a>
                            {/if}
                        </li>
                    {:else}
                        <li class="message">No matches</li>
                    {/each}
                {/if}
            </ul>
        </main>
    </div>
</dialog>

<style lang="scss">
    dialog {
        width: 80vw;
        height: 80vh;
        border-radius: 0.75rem;
        border: 1px solid var(--border-color);
        padding: 0;
        overflow: hidden;
        background-color: var(--color-bg-1);

        box-shadow: var(--fuzzy-finder-shadow);

        &::backdrop {
            background: var(--fuzzy-finder-backdrop);
        }

        @media (--mobile) {
            border-radius: 0;
            border: none;
            position: fixed;
            width: 100vw;
            height: 100vh;
            max-height: 100vh;
            max-width: 100vw;
        }
    }

    .content {
        display: flex;
        flex-direction: column;
        height: 100%;
    }

    .input {
        padding: 1rem;
        border-bottom: 1px solid var(--border-color);

        .scope {
            margin-top: 1rem;
            color: var(--text-muted);
        }
    }

    main {
        flex: 1;
        overflow: hidden;
        display: flex;
        flex-direction: column;

        ul {
            margin: 0;
            padding: 0;
            overflow-y: auto;
            list-style: none;
        }

        [role='option'] {
            a {
                display: grid;
                grid-template-columns: [icon] auto [label] 1fr;
                grid-template-rows: auto;
                grid-template-areas: 'icon label' '. info';
                column-gap: 0.5rem;

                cursor: pointer;
                padding: 0.25rem 0.75rem;

                text-decoration: none;
                color: var(--body-color);
                font-size: var(--font-size-small);
            }

            small {
                color: var(--text-muted);
            }

            &[aria-selected='true'] a,
            a:hover {
                background-color: var(--color-bg-2);
            }

            .icon {
                grid-area: icon;

                // Centers the icon vertically
                display: flex;
                align-items: center;
            }

            .label {
                grid-area: label;
            }

            .info {
                grid-area: info;
                color: var(--text-muted);
                font-size: var(--font-size-small);

                &.mono {
                    font-family: var(--code-font-family);
                }
            }
        }

        .message,
        .error {
            padding: 1rem;
        }

        .message {
            text-align: center;
            color: var(--text-muted);
        }
    }

    header {
        position: relative;
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 1rem;

        &::before {
            content: '';
            position: absolute;
            border-bottom: 1px solid var(--border-color);
            width: 100%;
            bottom: 0;
            left: 0;
        }

        .close {
            grid-area: close;
            position: fixed;
            right: 2rem;
            background-color: var(--color-bg-1);
            border-radius: 50%;

            &:hover {
                background-color: var(--color-bg-2);
            }
        }

        @media (--mobile) {
            display: grid;
            grid-template-columns: 1fr;
            grid-template-areas: 'close' 'tabs';
            padding: 0;

            .close {
                position: static;
            }
        }
    }
</style>
