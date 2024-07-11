<script lang="ts">
    import classNames from 'classnames'

    import styles from './Button.module.scss'

    const BUTTON_GROUP_DIRECTION = ['vertical', 'horizontal'] as const

    export let direction: typeof BUTTON_GROUP_DIRECTION[number] = 'horizontal'

    $: buttonClass = classNames(styles.btnGroup, direction === 'vertical' && styles.btnGroupVertical)
</script>

<slot name="custom" role="group" {buttonClass}>
    <div role="group" class={buttonClass}>
        <slot />
    </div>
</slot>
