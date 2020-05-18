// @ts-expect-error
import configure from 'core-js/configurator'

configure({
    // Polyfill URL because Chrome and Firefox are not spec-compliant
    // Hostnames of URIs with custom schemes (e.g. git) are not parsed out
    usePolyfill: ['URL'],
})
