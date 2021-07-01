const getEnvironmentBoolean = (value = 'false'): boolean => Boolean(JSON.parse(value))

export const environment = {
    customStoriesGlob: process.env.STORIES_GLOB,
    isBundleAnalyzerEnabled: getEnvironmentBoolean(process.env.WEBPACK_BUNDLE_ANALYZER),
    isSpeedAnalyzerEnabled: getEnvironmentBoolean(process.env.WEBPACK_SPEED_ANALYZER),
    shouldMinify: getEnvironmentBoolean(process.env.MINIFY),
}
