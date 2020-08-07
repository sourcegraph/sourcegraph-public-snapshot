// Fork of https://github.com/facebook/jest/blob/c95abca8082ad1e472828a6b2e5097745371707f/packages/jest-environment-jsdom/src/index.ts
// But exposes jsdom as a global variable
// and uses latest JSDOM

'use strict'

/**
 * Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

const { installCommonGlobals } = require('jest-util')
const { ModuleMocker } = require('jest-mock')
const { LegacyFakeTimers, ModernFakeTimers } = require('@jest/fake-timers')
const { JSDOM, VirtualConsole } = require('jsdom')

class JSDOMEnvironment {
  constructor(config, options = {}) {
    this.dom = new JSDOM('<!DOCTYPE html>', {
      pretendToBeVisual: true,
      runScripts: 'dangerously',
      url: config.testURL,
      virtualConsole: new VirtualConsole().sendTo(options.console || console),
      ...config.testEnvironmentOptions,
    })
    const global = (this.global = this.dom.window.document.defaultView)

    // Expose JSDOM as global to allow reconfiguring the URL
    global.jsdom = this.dom

    // JSDOM does not have SVGAElement implemented. Use a quick and dirty polyfill.
    // This does not implement href and target, which is impossible without mofifying JSDOM.
    global.SVGAElement = class SVGAElement extends global.SVGGraphicsElement {}

    // jsdom doesn't support document.queryCommandSupported(), needed for monaco-editor.
    // https://github.com/testing-library/react-testing-library/issues/546
    // eslint-disable-next-line @typescript-eslint/unbound-method
    this.dom.window.document.queryCommandSupported = () => false

    if (!global) {
      throw new Error('JSDOM did not return a Window object')
    }

    // In the `jsdom@16`, ArrayBuffer was not added to Window, ref: https://github.com/jsdom/jsdom/commit/3a4fd6258e6b13e9cf8341ddba60a06b9b5c7b5b
    // Install ArrayBuffer to Window to fix it. Make sure the test is passed, ref: https://github.com/facebook/jest/pull/7626
    global.ArrayBuffer = ArrayBuffer

    // Node's error-message stack size is limited at 10, but it's pretty useful
    // to see more than that when a test fails.
    this.global.Error.stackTraceLimit = 100
    installCommonGlobals(global, config.globals)

    // Report uncaught errors.
    this.errorEventListener = event => {
      if (userErrorListenerCount === 0 && event.error) {
        process.emit('uncaughtException', event.error)
      }
    }
    global.addEventListener('error', this.errorEventListener)

    // However, don't report them as uncaught if the user listens to 'error' event.
    // In that case, we assume the might have custom error handling logic.
    // eslint-disable-next-line @typescript-eslint/unbound-method
    const originalAddListener = global.addEventListener
    // eslint-disable-next-line @typescript-eslint/unbound-method
    const originalRemoveListener = global.removeEventListener
    let userErrorListenerCount = 0
    global.addEventListener = function (...args) {
      if (args[0] === 'error') {
        userErrorListenerCount++
      }
      return originalAddListener.apply(this, args)
    }
    global.removeEventListener = function (...args) {
      if (args[0] === 'error') {
        userErrorListenerCount--
      }
      return originalRemoveListener.apply(this, args)
    }

    this.moduleMocker = new ModuleMocker(global)

    const timerConfig = {
      idToRef: id => id,
      refToId: reference => reference,
    }

    this.fakeTimers = new LegacyFakeTimers({
      config,
      global,
      moduleMocker: this.moduleMocker,
      timerConfig,
    })

    this.fakeTimersModern = new ModernFakeTimers({ config, global })
  }

  setup() {
    return Promise.resolve()
  }

  teardown() {
    if (this.fakeTimers) {
      this.fakeTimers.dispose()
    }
    if (this.fakeTimersModern) {
      this.fakeTimersModern.dispose()
    }
    if (this.global) {
      if (this.errorEventListener) {
        this.global.removeEventListener('error', this.errorEventListener)
      }
      // Dispose "document" to prevent "load" event from triggering.
      Object.defineProperty(this.global, 'document', { value: null })
      this.global.close()
    }
    this.errorEventListener = null
    this.global = null
    this.dom = null
    this.fakeTimers = null
    this.fakeTimersModern = null
    return Promise.resolve()
  }

  runScript(script) {
    if (this.dom) {
      return script.runInContext(this.dom.getInternalVMContext())
    }
    return null
  }

  getVmContext() {
    if (this.dom) {
      return this.dom.getInternalVMContext()
    }
    return null
  }
}

module.exports = JSDOMEnvironment
