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
        <span
            class="square"
            class:bg-success={type === SquareType.Added}
            class:bg-danger={type === SquareType.Deleted}
        />
    {/each}
</span>

<style lang="scss">
    .root {
        display: inline-flex;
    }

    .square {
        display: inline-block;
        width: 0.5rem;
        height: 0.5rem;
        background-color: var(--text-muted);
        margin-left: 0.125rem;
    }
</style>
