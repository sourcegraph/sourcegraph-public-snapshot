// Same polyfills as the webapp
import '../../../../shared/src/polyfills'

// Polyfill global browser API for Chrome
// The API is much nicer to use because it supports Promises
// The polyfill is also very simple.
import browser from 'webextension-polyfill'
Object.assign(self, { browser })
