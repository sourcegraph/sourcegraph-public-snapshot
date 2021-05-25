import Shepherd from 'shepherd.js'

export const defaultTourOptions: Shepherd.Tour.TourOptions = {
    useModalOverlay: false,
    defaultStepOptions: {
        arrow: false,
        attachTo: { on: 'bottom' },
        scrollTo: false,
    },
}
