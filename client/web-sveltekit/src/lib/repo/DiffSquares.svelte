<script lang="ts">
    const enum SquareType {
        Added,
        Deleted,
        Neutral,
    }

    const NUM_SQUARES = 5

    export let added: number
    export let deleted: number

    let squares: SquareType[] = []
    $: {
        const total = added + deleted
        const numSquares = Math.min(NUM_SQUARES, total)
        const ratio = numSquares / total
        const result: SquareType[] = []

        for (let i = 0; i < Math.floor(added * ratio); i++) {
            result.push(SquareType.Added)
        }
        for (let i = 0; i < Math.floor(deleted * ratio); i++) {
            result.push(SquareType.Deleted)
        }
        for (let i = Math.max(NUM_SQUARES - result.length, 0); i > 0; i--) {
            result.push(SquareType.Neutral)
        }

        squares = result
    }
</script>

<span class="root">
    {#each squares as type}
        <span class="square" class:added={type === SquareType.Added} class:deleted={type === SquareType.Deleted} />
    {/each}
</span>

<style lang="scss">
    .root {
        display: inline-flex;
        gap: 0.125rem;
        margin-left: 0.25rem;
    }

    .square {
        display: inline-block;
        width: 0.5rem;
        height: 0.5rem;
        background-color: var(--color-bg-3);
    }

    .added {
        background-color: var(--success);
    }

    .deleted {
        background-color: var(--danger);
    }
</style>
