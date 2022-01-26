/*
    https://github.com/react-bootstrap/react-bootstrap/issues/4997
    Popper causes "Warning: `NaN` is an invalid value for the `left` css style property."
    This mock prevents that.
*/
jest.mock('popper.js', () => {
    const StockPopperJs = jest.requireActual('popper.js')

    return function PopperJs() {
        const placements = StockPopperJs.placements

        return {
            destroy: () => {},
            scheduleUpdate: () => {},
            update: () => {},
            placements,
        }
    }
})
