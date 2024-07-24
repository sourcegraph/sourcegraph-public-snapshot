<script lang="ts">
    import classNames from 'classnames'

    import { getButtonDisplay, type BUTTON_DISPLAY } from './Button'

    import styles from './Button.module.scss'

    const BUTTON_GROUP_DIRECTION = ['vertical', 'horizontal'] as const

    export let direction: typeof BUTTON_GROUP_DIRECTION[number] = 'horizontal'
    export let display: typeof BUTTON_DISPLAY[number] | undefined = undefined

    $: buttonClass = classNames(
        styles.btnGroup,
        direction === 'vertical' && styles.btnGroupVertical,
        display && getButtonDisplay({ display })
    )
</script>

<slot name="custom" role="group" {buttonClass}>
    <div role="group" class={buttonClass}>
        <slot />
    </div>
</slot>
