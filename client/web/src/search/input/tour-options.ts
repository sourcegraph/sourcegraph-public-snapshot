import Shepherd from 'shepherd.js'

export const defaultTourOptions: Shepherd.Tour.TourOptions = {
    useModalOverlay: false,
    defaultStepOptions: {
        arrow: true,
        classes: 'web-content tour-card card py-4 px-3 shadow-lg',
        attachTo: { on: 'bottom' },
        scrollTo: false,
    },
}
