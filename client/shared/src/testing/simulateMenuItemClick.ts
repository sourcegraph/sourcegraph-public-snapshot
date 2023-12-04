import { fireEvent } from '@testing-library/react'

/**
 * userEvent.click does not work for Reach menu items. Use this function from Reach's official test code instead.
 * Modified from https://sourcegraph.com/github.com/reach/reach-ui@26c826684729e51e45eef29aa4316df19c0e2c03/-/blob/test/utils.tsx?L105
 */
export function simulateMenuItemClick(element: HTMLElement): void {
    fireEvent.mouseEnter(element)
    fireEvent.keyDown(element, { key: ' ' })
    fireEvent.keyUp(element, { key: ' ' })
}
