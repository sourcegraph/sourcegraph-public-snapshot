<script lang="ts">
    import { mdiClose } from '@mdi/js'
    import { tick } from 'svelte'

    import { nextSibling, onClickOutside, previousSibling } from '$lib/dom'
    import { getGraphQLClient } from '$lib/graphql'
    import { formatShortcut } from '$lib/Hotkey'
    import Icon from '$lib/Icon.svelte'
    import FileIcon from '$lib/repo/FileIcon.svelte'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import EmphasizedLabel from '$lib/search/EmphasizedLabel.svelte'
    import SymbolKind from '$lib/search/SymbolKind.svelte'
    import TabsHeader, { type Tab } from '$lib/TabsHeader.svelte'
    import { Input } from '$lib/wildcard'
    import Button from '$lib/wildcard/Button.svelte'

    import { filesHotkey, reposHotkey, symbolsHotkey } from './keys'
    import {
        createRepositorySource,
        type CompletionSource,
        createFileSource,
        type FuzzyFinderResult,
        createSymbolSource,
    } from './sources'

    export let open = false
    export let scope = ''

    export function selectTab(tabID: 'repos' | 'symbols' | 'files') {
        if (selectedTab.id !== tabID) {
            selectedOption = 0
            selectedTab = tabs.find(t => t.id === tabID) ?? tabs[0]
        }
    }

    let dialog: HTMLDialogElement | undefined
    let listbox: HTMLElement | undefined
    let input: HTMLInputElement | undefined
    let query = ''

    const client = getGraphQLClient()
    const tabs: (Tab & { source: CompletionSource<FuzzyFinderResult> })[] = [
        { id: 'repos', title: 'Repos', source: createRepositorySource(client) },
        { id: 'symbols', title: 'Symbols', source: createSymbolSource(client, () => scope) },
        { id: 'files', title: 'Files', source: createFileSource(client, () => scope) },
    ]

    async function handleKeyboardEvent(event: KeyboardEvent) {
        switch (event.key) {
            // Select the next/first option
            case 'ArrowDown': {
                event.preventDefault()
                let next: HTMLElement | null = null
                const current = listbox?.querySelector('[aria-selected="true"]')
                if (current) {
                    next = nextSibling(current, '[role="option"]', true) as HTMLElement | null
                } else {
                    next = listbox?.querySelector('[role="option"]') as HTMLElement | null
                }

                if (next?.dataset.index) {
                    selectedOption = +next.dataset.index
                    await tick()
                    next.scrollIntoView({ block: 'nearest' })
                }
                break
            }
            // Select previous/last option
            case 'ArrowUp': {
                event.preventDefault()
                let prev: HTMLElement | null = null
                const current = listbox?.querySelector('[aria-selected="true"]')
                if (current) {
                    prev = previousSibling(current, '[role="option"]', true) as HTMLElement | null
                } else {
                    prev = listbox?.querySelector('[role="option"]:last-child') as HTMLElement | null
                }

                if (prev?.dataset.index) {
                    selectedOption = +prev.dataset.index
                    await tick()
                    prev.scrollIntoView({ block: 'nearest' })
                }
                break
            }
            // Select first option
            case 'Home': {
                const option = listbox?.querySelector('[role="option"]')
                if (option) {
                    selectedOption = 0
                    await tick()
                    option.scrollIntoView({ block: 'nearest' })
                }
                break
            }
            // Select last option
            case 'End': {
                const options = listbox?.querySelectorAll('[role="option"]')
                if (options && options.length > 0) {
                    selectedOption = options.length - 1
                    await tick()
                    options[selectedOption].scrollIntoView({ block: 'nearest' })
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

    function handleClick(event: MouseEvent) {
        const target = event.target as HTMLElement
        const option = target.closest('[role="option"]') as HTMLElement | null
        if (option?.dataset.index) {
            selectedOption = +option.dataset.index
            dialog?.close()
        }
    }

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
</script>

<dialog bind:this={dialog} on:close>
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
            >
                <span slot="after-title" let:tab>
                    {#if tab.id === 'repos'}
                        <kbd>{formatShortcut(reposHotkey)}</kbd>
                    {:else if tab.id === 'symbols'}
                        <kbd>{formatShortcut(symbolsHotkey)}</kbd>
                    {:else if tab.id === 'files'}
                        <kbd>{formatShortcut(filesHotkey)}</kbd>
                    {/if}
                </span>
            </TabsHeader>
            <Button variant="icon" on:click={() => dialog?.close()} size="sm">
                <Icon svgPath={mdiClose} aria-label="Close" />
            </Button>
        </header>
        <main>
            <div class="input">
                <Input
                    type="text"
                    bind:input
                    placeholder="Enter a fuzzy query"
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
                />
                {#if useScope}
                    <div class="scope">Searching in <code>{scope}</code></div>
                {/if}
            </div>
            <ul role="listbox" bind:this={listbox} aria-label="Search results">
                {#if $source.value}
                    {#each $source.value as item, index (item.item)}
                        <li role="option" aria-selected={selectedOption === index} data-index={index}>
                            {#if item.item.type === 'repo'}
                                <a href="/{item.item.repository.name}" on:click={handleClick}>
                                    <CodeHostIcon repository={item.item.repository.name} />
                                    <span
                                        ><EmphasizedLabel
                                            label={item.item.repository.name}
                                            matches={item.positions}
                                        /></span
                                    >
                                </a>
                            {:else if item.item.type == 'symbol'}
                                <a href={item.item.symbol.location.url} on:click={handleClick}>
                                    <SymbolKind symbolKind={item.item.symbol.kind} />
                                    <span
                                        ><EmphasizedLabel
                                            label={item.item.symbol.name}
                                            matches={item.positions}
                                        /></span
                                    >
                                    <small>-</small>
                                    <FileIcon file={item.item.file} inline />
                                    <small
                                        >{#if !useScope}{item.item.repository.name}/{/if}{item.item.file.path}</small
                                    >
                                </a>
                            {:else if item.item.type == 'file'}
                                <a href={item.item.file.url} on:click={handleClick}>
                                    <FileIcon file={item.item.file} inline />
                                    <span
                                        >{#if !useScope}{item.item.repository.name}/{/if}<EmphasizedLabel
                                            label={item.item.file.path}
                                            matches={item.positions}
                                        /></span
                                    >
                                </a>
                            {/if}
                        </li>
                    {:else}
                        <li class="empty">No matches</li>
                    {/each}
                {/if}
            </ul>
        </main>
    </div>
</dialog>

<style lang="scss">
    dialog {
        background-color: var(--color-bg-1);
        width: 80vw;
        height: 80vh;
        border: 1px solid var(--border-color);
        padding: 0;
        overflow: hidden;

        &::backdrop {
            background-color: rgba(0, 0, 0, 0.3);
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
        }

        [role='option'] {
            a {
                display: flex;
                align-items: center;
                padding: 0.25rem 1rem;
                cursor: pointer;
                color: var(--body-color);
                gap: 0.25rem;

                text-decoration: none;
            }

            small {
                color: var(--text-muted);
            }

            &[aria-selected='true'] a,
            a:hover {
                background-color: var(--color-bg-2);
            }
        }

        .empty {
            padding: 1rem;
            text-align: center;
            color: var(--text-muted);
        }
    }

    header {
        padding: 0 1rem;
    }

    header {
        border-bottom: 1px solid var(--border-color);
        display: flex;
        align-items: center;
        justify-content: space-between;
    }
</style>
