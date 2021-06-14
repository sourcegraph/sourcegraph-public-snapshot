import Shepherd from 'shepherd.js'

export const defaultPopperModifiers = [
    {
        name: 'focusAfterRender',
        enabled: false,
    },
    {
        name: 'preventOverflow',
        options: {
            tether: false,
        },
    },
    {
        name: 'hide',
    },
]

export const defaultTourOptions: Shepherd.Tour.TourOptions = {
    useModalOverlay: false,
    defaultStepOptions: {
        arrow: false,
        attachTo: { on: 'bottom' },
        scrollTo: false,
    },
}
