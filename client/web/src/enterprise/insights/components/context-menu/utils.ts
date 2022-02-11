import { getCollisions, Position } from '@reach/popover'

const DEFAULT_PADDING = 6

/**
 * Custom popover position calculator. Returns position objects (top,left,right,bottom) styles
 * with values such that the target and the popover element have the same right borders.
 *
 * <pre>
 * ------------ | Target | --------
 * ----|*****************| --------
 * ----|*****************| --------
 * ----|*** Popover *****| --------
 * ----|*****************| --------
 * ----|*****************| --------
 * </pre>
 *
 * @param targetRectangle - bounding client rect of the target element
 * @param popoverRectangle - bounding client rect of the pop-over element. All calculation props
 * that are returned from this function will be applied to this element.
 */
export const positionBottomRight: Position = (targetRectangle, popoverRectangle) => {
    if (!targetRectangle || !popoverRectangle) {
        return {}
    }

    const { directionUp } = getCollisions(targetRectangle, popoverRectangle)

    return {
        left: `${targetRectangle.right - popoverRectangle.width + window.scrollX}px`,
        top: directionUp
            ? `${targetRectangle.top - popoverRectangle.height + window.scrollY - DEFAULT_PADDING}px`
            : `${targetRectangle.top + targetRectangle.height + window.scrollY + DEFAULT_PADDING}px`,
    }
}

/**
 * Custom position calculator with flip logic.
 *
 * <pre>
 * In case if it's enough space at right
 * --| Target ||*****************|--
 * ------------|*****************|--
 * ------------|**** Popover ****|--
 * ------------|*****************|--
 * ------------|*****************|--
 *
 * In other case if it's enough space at left side
 * --|*****************|| Target |
 * --|*****************|
 * --|**** Popover ****|
 * --|*****************|
 * --|*****************|
 *
 * And as a fallback plan place it below the target
 * ------------ | Target | --------
 * ----|*****************| --------
 * ----|*****************| --------
 * ----|*** Popover *****| --------
 * ----|*****************| --------
 * ----|*****************| --------
 * </pre>
 *
 * @param targetRectangle - bounding client rect of the target element
 * @param popoverRectangle - bounding client rect of the pop-over element. All calculation props
 * that are returned from this function will be applied to this element
 */
export const flipRightPosition: Position = (targetRectangle, popoverRectangle) => {
    if (!targetRectangle || !popoverRectangle) {
        return {}
    }

    const isEnoughSpaceLeft = targetRectangle.left - popoverRectangle.width > 0
    const isEnoughSpaceRight = window.innerWidth > targetRectangle.right + popoverRectangle.width

    const { directionUp } = getCollisions(targetRectangle, popoverRectangle)

    if (isEnoughSpaceRight) {
        return {
            left: `${targetRectangle.right + window.scrollX + DEFAULT_PADDING}px`,
            top: `${targetRectangle.top + window.scrollY - 4}px`,
        }
    }

    if (isEnoughSpaceLeft) {
        return {
            left: `${targetRectangle.left - popoverRectangle.width + window.scrollX - DEFAULT_PADDING}px`,
            top: `${targetRectangle.top + window.scrollY - 4}px`,
        }
    }

    return {
        left: `${targetRectangle.right - popoverRectangle.width + window.scrollX}px`,
        top: directionUp
            ? `${targetRectangle.top - popoverRectangle.height + window.scrollY - DEFAULT_PADDING}px`
            : `${targetRectangle.top + targetRectangle.height + window.scrollY + DEFAULT_PADDING}px`,
    }
}
