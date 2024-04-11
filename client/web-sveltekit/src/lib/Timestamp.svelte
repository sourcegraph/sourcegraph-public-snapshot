<script lang="ts" context="module">
    import { addMinutes, format as dateFnsFormat } from 'date-fns'

    export function formatDate(date: Date, options: { utc?: boolean; format?: TimestampFormat }): string {
        if (options.utc) {
            date = addMinutes(date, date.getTimezoneOffset())
        }
        const dateHasTime = date.toString().includes('T')
        const defaultFormat =
            options.format ?? (dateHasTime ? TimestampFormat.FULL_DATE_TIME : TimestampFormat.FULL_DATE)
        return dateFnsFormat(date, defaultFormat) + (options.utc ? ' UTC' : '')
    }

    export enum TimestampFormat {
        FULL_TIME = 'HH:mm:ss',
        FULL_DATE = 'yyyy-MM-dd',
        FULL_DATE_TIME = 'yyyy-MM-dd pp',
    }
</script>

<script lang="ts">
<<<<<<< HEAD
    import { formatDistance, formatDistanceStrict } from 'date-fns'
=======
    import { addMinutes, format, formatDistance, formatDistanceStrict } from 'date-fns'
>>>>>>> cf1e9356ac9 (reduce html elements)

    import { currentDate } from './stores'
    import Tooltip from './Tooltip.svelte'

    /** The date (if string, in ISO 8601 format). */
    export let date: Date | string

    /** Use exact timestamps (i.e. omit "less than", "about", etc.) */
    export let strict: boolean = false

    /** Show absolute timestamp and show relative timestamp in label*/
    export let showAbsolute: boolean = false

    /** Hide suffix (e.g. ago) */
    export let hideSuffix: boolean = false

    /** Show time in UTC */
    export let utc: boolean | undefined = undefined

    $: dateObj = typeof date === 'string' ? new Date(date) : date
    $: formattedDate = formatDate(dateObj, { utc })
    $: relativeDate = (strict ? formatDistanceStrict : formatDistance)(dateObj, $currentDate, {
        addSuffix: !hideSuffix,
    })
</script>

<Tooltip tooltip={showAbsolute ? relativeDate : formattedDate}>
<<<<<<< HEAD
<<<<<<< HEAD
    <span class="timestamp" data-testid="timestamp">{showAbsolute ? formattedDate : relativeDate} </span>
=======
    {#if small}
        <span class="timestamp" data-testid="timestamp"
            ><small>{showAbsolute ? formattedDate : relativeDate}</small>
        </span>
    {:else}
        <span class="timestamp" data-testid="timestamp">{showAbsolute ? formattedDate : relativeDate} </span>
    {/if}
>>>>>>> cf1e9356ac9 (reduce html elements)
=======
    <span class="timestamp" data-testid="timestamp">{showAbsolute ? formattedDate : relativeDate} </span>
>>>>>>> 9a4935ff015 (restructure last commit section from a column to a row)
</Tooltip>
