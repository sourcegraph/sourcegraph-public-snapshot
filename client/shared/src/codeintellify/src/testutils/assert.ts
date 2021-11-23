export const isWithinOne = (a: number, b: number): void =>
    chai.assert(Math.abs(a - b) < 1, `expected the difference between ${a} and ${b} to be less than 1`)
