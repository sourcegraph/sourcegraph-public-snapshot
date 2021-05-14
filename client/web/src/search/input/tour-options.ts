import Shepherd from 'shepherd.js'

export const defaultTourOptions: Shepherd.Tour.TourOptions = {
    useModalOverlay: false,
    defaultStepOptions: {
        arrow: false,
        classes: 'web-content tour-card shadow-lg',
        attachTo: { on: 'bottom' },
        scrollTo: false,
    },
}
