<script lang="ts">
    import { EditorView } from '@codemirror/view'

    export let view: EditorView
    export let parent: HTMLDivElement

    export function show() {
        visible = true
    }
    export function hide() {
        visible = false
    }

    let visible = false
</script>

{#if visible}
    <div>Hello world</div>
    <div>Hello world</div>
    <div>Hello world</div>
    <div>Hello world</div>
    <div>Hello world</div>
    <div>Hello world</div>
    <div>Hello world</div>
{/if}

<style lang="scss">
    div {
        margin-top: 0.5rem;
    }
</style>
