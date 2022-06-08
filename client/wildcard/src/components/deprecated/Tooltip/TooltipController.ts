import { UseTooltipReturn } from './useTooltipState'

type TooltipInstance = Pick<UseTooltipReturn, 'forceUpdate'>

/**
 * @deprecated We will be moving away from the `data-tooltip` pattern and now provide a `Tooltip` component
 * to use directly. This controller will be removed after all uses of `data-tooltip` are migrated to use the
 * new `Tooltip` component.
 */
// eslint-disable-next-line @typescript-eslint/no-extraneous-class
export class TooltipController {
    private static instance: TooltipInstance | undefined

    public static setInstance(instance?: TooltipInstance): void {
        this.instance = instance
    }

    public static forceUpdate(): void {
        this.instance?.forceUpdate()
    }
}
