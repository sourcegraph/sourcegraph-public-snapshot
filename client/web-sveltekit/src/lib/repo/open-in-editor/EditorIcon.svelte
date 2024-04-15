<script lang="ts">
    import type { Editor } from '$lib/web'
    export let editorId: Editor['id']

    let Icon: any;
    $: import(`./editors/${editorId}.svelte`).then(i => Icon = i.default);
</script>

 {#if Icon}
    <Icon />
 {/if}
