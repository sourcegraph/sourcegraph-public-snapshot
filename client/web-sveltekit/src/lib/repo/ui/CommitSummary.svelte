<script context="module" lang="ts">
    interface User {
        displayName: string
        avatarURL: string
    }
</script>
<script lang="ts">
    import UserAvatar from "$lib/UserAvatar.svelte"

    export let subject: string;
    export let author: User;
    export let url: string;
    export let id: string;
</script>


<div>
    <span><UserAvatar user={author} /></span>
    <p>{subject}</p>
    <a href={url}>{id}</a>
</div>

<style lang="scss">
    div {
        flex: 1;
        display: flex;
        align-items: center;
    }
    p {
        flex: 1;
        margin: 0;
        margin-left: 0.5rem;
    }
</style>
