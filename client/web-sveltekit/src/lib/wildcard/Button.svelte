<script lang="ts">
    // In addition to the props explicitly listed here, this component also
    // accepts any HTMLButton attributes. Note that those will only be used when
    // the default implementation is used.
    import type { HTMLButtonAttributes } from 'svelte/elements'

    import { type BUTTON_DISPLAY, type BUTTON_SIZES, type BUTTON_VARIANTS, getButtonClassName } from './Button'

    interface $$Props extends HTMLButtonAttributes {
        variant?: typeof BUTTON_VARIANTS[number]
        size?: typeof BUTTON_SIZES[number]
        display?: typeof BUTTON_DISPLAY[number]
        outline?: boolean
    }

    export let variant: $$Props['variant'] = 'primary'
    export let size: $$Props['size'] = undefined
    export let display: $$Props['display'] = undefined
    export let outline: $$Props['outline'] = undefined

    $: brandedButtonClassname = getButtonClassName({ variant, outline, display, size })
</script>

<slot name="custom" className={brandedButtonClassname}>
    <!-- $$restProps holds all the additional props that are passed to the component -->
    <button class={brandedButtonClassname} {...$$restProps} on:click|preventDefault>
        <slot />
    </button>
</slot>
