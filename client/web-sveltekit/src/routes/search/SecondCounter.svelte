<script lang="ts">
    import { onMount, onDestroy } from 'svelte'

    let secondsElapsed: number = 0
    let triggerTime: Date | null = null
    let secondsInterval: NodeJS.Timeout | null = null

    onMount(() => {
        triggerTime = new Date() // Set the trigger time to the current time
        secondsInterval = setInterval(() => {
            // Update the seconds elapsed every second
            if (triggerTime) {
                secondsElapsed = Math.floor((Date.now() - triggerTime.getTime()) / 1000)
            }
        }, 1000)
    })

    onDestroy(() => {
        if (secondsInterval) {
            clearInterval(secondsInterval)
        }
    })
</script>

<span>{secondsElapsed}s</span>

<style>
</style>
