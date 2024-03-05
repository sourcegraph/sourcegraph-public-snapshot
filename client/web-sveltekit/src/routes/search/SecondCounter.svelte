<script lang="ts">
    import { onMount } from 'svelte'

    let displaySeconds = 0
    let displayMilliseconds = 0

    export let secondsElapsed = 0
    let millisecondsElapsed = 0

    let lastUpdate = Date.now()

    onMount(() => {
        const startTime = Date.now()

        setInterval(() => {
            const now = Date.now()
            const delta = now - startTime

            secondsElapsed = Math.floor(delta / 1000)
            millisecondsElapsed = delta % 100

            if (now - lastUpdate >= 5) {
                displaySeconds = secondsElapsed
                displayMilliseconds = millisecondsElapsed
                lastUpdate = now
            }
        }, 1300)
    })
</script>

<span>
    {secondsElapsed}.{millisecondsElapsed}s
</span>

<style>
</style>
