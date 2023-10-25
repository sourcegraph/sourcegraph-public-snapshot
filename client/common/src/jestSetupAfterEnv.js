// jest-fetch-mock assumes the `jest` global is available, but we use explicit imports from
// `vitest`. This is a workaround to make jest-fetch-mock work. See
// https://github.com/jefflau/jest-fetch-mock/issues/104.
global.jest = require('vitest').jest
