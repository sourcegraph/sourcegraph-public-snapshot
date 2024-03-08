<script lang="ts">
    import styles from './Markdown.module.scss'
    export let dangerousInnerHTML: string = ''
</script>

{#if dangerousInnerHTML}
    <div class={styles.markdown}>
        {@html dangerousInnerHTML}
    </div>
{/if}

<style lang="scss">
</style>
