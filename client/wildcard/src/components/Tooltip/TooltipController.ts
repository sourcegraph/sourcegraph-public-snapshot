import { UseTooltipReturn } from './useTooltipState'

type TooltipInstance = Pick<UseTooltipReturn, 'forceUpdate'>

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
