<script lang="ts">
    import type { Readme_Blob } from './Readme.gql'

    export let file: Readme_Blob
</script>

<div>
    {#if file.richHTML}
        {@html file.richHTML}
    {:else if file.content}
        <pre>{file.content}</pre>
    {/if}
</div>

<style lang="scss">
    div {
        :global(img) {
            max-width: 100%;
        }

        pre {
            white-space: pre-wrap;
        }
    }
</style>
