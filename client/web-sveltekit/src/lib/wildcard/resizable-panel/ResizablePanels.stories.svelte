<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import Panel from './Panel.svelte'
    import PanelGroup from './PanelGroup.svelte'
    import PanelResizeHandle from './PanelResizeHandle.svelte'

    export const meta = {
        component: PanelGroup,
    }
</script>

<script lang="ts">
    let showLeftPanel = true
    let showRightPanel = true

    function handleClick(side: 'left' | 'right') {
        switch (side) {
            case 'left': {
                showLeftPanel = !showLeftPanel
                break
            }
            case 'right': {
                showRightPanel = !showRightPanel
                break
            }
        }
    }

    // Collapsable panels example
    let leftPanel: Panel

    function handleExpandCollapsePanel() {
        if (leftPanel.isCollapsed()) {
            leftPanel.expand()
        } else {
            leftPanel.collapse()
        }
    }
</script>

<Story name="Horizontal panels">
    <section class="root">
        <PanelGroup id="main-group" direction="horizontal">
            <Panel defaultSize={30} minSize={20} id="first" order={1}>
                <div class="item">left</div>
            </Panel>
            <PanelResizeHandle />
            <Panel minSize={30} id="second" order={2}>
                <div class="item">middle</div>
            </Panel>
            <PanelResizeHandle />
            <Panel defaultSize={30} minSize={20} id="third" order={3}>
                <div class="item">right</div>
            </Panel>
        </PanelGroup>
    </section>
</Story>

<Story name="Vertical panels">
    <section class="root">
        <PanelGroup id="main-vertical" direction="vertical">
            <Panel maxSize={75}>
                <div class="item">Top</div>
            </Panel>
            <PanelResizeHandle />
            <Panel maxSize={75}>
                <div class="item">Bottom</div>
            </Panel>
        </PanelGroup>
    </section>
</Story>

<Story name="Conditional panels">
    <section>
        <header>
            <button on:click={() => handleClick('left')}>Toggle left panel</button>
            <button on:click={() => handleClick('right')}>Toggle right panel</button>
        </header>

        <div class="root">
            <PanelGroup id="main-conditional-group" direction="horizontal">
                {#if showLeftPanel}
                    <Panel minSize={20} id="first" order={1}>
                        <div class="item">left</div>
                    </Panel>
                    <PanelResizeHandle />
                {/if}

                <Panel minSize={30} id="second" order={2}>
                    <div class="item">middle</div>
                </Panel>

                {#if showRightPanel}
                    <PanelResizeHandle />
                    <Panel minSize={20} id="third" order={3}>
                        <div class="item">right</div>
                    </Panel>
                {/if}
            </PanelGroup>
        </div>
    </section>
</Story>

<Story name="Nested panels">
    <section class="root">
        <PanelGroup id="main-horizontal" direction="horizontal">
            <Panel defaultSize={20} minSize={10}>
                <div class="item">left</div>
            </Panel>
            <PanelResizeHandle />
            <Panel minSize={35}>
                <PanelGroup id="main-nested-vertical" direction="vertical">
                    <Panel defaultSize={35} minSize={10}>
                        <div class="item">Top</div>
                    </Panel>
                    <PanelResizeHandle />
                    <Panel minSize={10}>
                        <PanelGroup id="main-nested-horizontal" direction="horizontal">
                            <Panel minSize={10}>
                                <div class="item">left</div>
                            </Panel>
                            <PanelResizeHandle />
                            <Panel minSize={10}>
                                <div class="item">right</div>
                            </Panel>
                        </PanelGroup>
                    </Panel>
                </PanelGroup>
            </Panel>
            <PanelResizeHandle />
            <Panel defaultSize={20} minSize={10}>
                <div class="item">right</div>
            </Panel>
        </PanelGroup>
    </section>
</Story>

<Story name="Collapsable panels">
    <button on:click={handleExpandCollapsePanel}>Collapse/expand left panel</button>

    <section class="root">
        <PanelGroup id="main-group" direction="horizontal">
            <Panel
                id="first"
                order={1}
                minSize={40}
                defaultSize={50}
                collapsible
                collapsedSize={15}
                bind:this={leftPanel}
            >
                <svelte:fragment let:isCollapsed>
                    <div class="item" class:isCollapsed>left</div>
                </svelte:fragment>
            </Panel>
            <PanelResizeHandle />
            <Panel minSize={20} id="second" order={2}>
                <div class="item">middle</div>
            </Panel>
            <PanelResizeHandle />
            <Panel defaultSize={30} minSize={20} id="third" order={3}>
                <div class="item">right</div>
            </Panel>
        </PanelGroup>
    </section>
</Story>

<style>
    .root {
        height: 40rem;
        color: white;
    }

    .item {
        background-color: #192230;
        word-break: break-all;
        border-radius: 0.25rem;
        flex: auto;
        justify-content: center;
        align-items: center;
        padding: 0.5rem;
        font-size: 1rem;
        display: flex;
        overflow: hidden;
        height: 100%;
    }

    .isCollapsed {
        background-color: darkblue;
    }
</style>
