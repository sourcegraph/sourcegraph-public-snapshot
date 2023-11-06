// JSDOM does not have support for DOMRect, needed for tooltips.
// https://github.com/radix-ui/primitives/issues/420#issuecomment-771615182
if ('DOMRect' in window === false) {
    window.DOMRect = {
        fromRect: () => ({ top: 0, left: 0, bottom: 0, right: 0, width: 0, height: 0 }),
    } as any
}
