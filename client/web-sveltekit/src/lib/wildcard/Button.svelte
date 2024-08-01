<script lang="ts">
    // In addition to the props explicitly listed here, this component also
    // accepts any HTMLButton attributes. Note that those will only be used when
    // the default implementation is used.

    import type { HTMLButtonAttributes } from 'svelte/elements'

    import { type BUTTON_DISPLAY, type BUTTON_SIZES, type BUTTON_VARIANTS, getButtonClassName } from './Button'

    type $$Props = {
        variant?: typeof BUTTON_VARIANTS[number]
        size?: typeof BUTTON_SIZES[number]
        display?: typeof BUTTON_DISPLAY[number]
        outline?: boolean
        // This is already allowed by HTMLButtonAttributes ([key: `data-${string}`]: any)
        // but for some reason it's not recognized when using `svelte-check` and an error
        // is thrown instead.
        'data-testid'?: string
        'data-scope-button'?: boolean
    } & HTMLButtonAttributes

    export let variant: $$Props['variant'] = 'primary'
    export let size: $$Props['size'] = undefined
    export let display: $$Props['display'] = undefined
    export let outline: $$Props['outline'] = undefined

    $: buttonClass = getButtonClassName({ variant, outline, display, size })
</script>

<slot name="custom" {buttonClass}>
    <!-- $$restProps holds all the additional props that are passed to the component -->
    <button class={buttonClass} type="button" {...$$restProps} on:click>
        <slot />
    </button>
</slot>
