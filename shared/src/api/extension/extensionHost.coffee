###*
# Starts the extension host, which runs extensions. It is a Web Worker or other similar isolated
# JavaScript execution context. There is exactly 1 extension host, and it has zero or more
# extensions activated (and running).
#
# It expects to receive a message containing {@link InitData} from the client application as the
# first message.
#
# @param endpoints The endpoints to the client.
# @returns An unsubscribable to terminate the extension host.
###

startExtensionHost = (endpoints) ->
  subscription = new (rxjs_1.Subscription)
  # Wait for "initialize" message from client application before proceeding to create the
  # extension host.
  initialized = false
  extensionAPI = new Promise((resolve) ->

    factory = (initData) ->
      `var extensionAPI`
      if initialized
        throw new Error('extension host is already initialized')
      initialized = true
      _a = initializeExtensionHost(endpoints, initData)
      extHostSubscription = _a.subscription
      extensionAPI = _a.extensionAPI
      extensionHostAPI = _a.extensionHostAPI
      subscription.add extHostSubscription
      resolve extensionAPI
      extensionHostAPI

    comlink.expose factory, endpoints.expose
    return
)
  {
    unsubscribe: ->
      subscription.unsubscribe()
    extensionAPI: extensionAPI
  }

###*
# Initializes the extension host using the {@link InitData} from the client application. It is
# called by {@link startExtensionHost} after the {@link InitData} is received.
#
# The extension API is made globally available to all requires/imports of the "sourcegraph" module
# by other scripts running in the same JavaScript context.
#
# @param endpoints The endpoints to the client.
# @param initData The information to initialize this extension host.
# @returns An unsubscribable to terminate the extension host.
###

initializeExtensionHost = (endpoints, initData) ->
  subscription = new (rxjs_1.Subscription)
  _a = createExtensionAPI(initData, endpoints)
  extensionAPI = _a.extensionAPI
  extensionHostAPI = _a.extensionHostAPI
  apiSubscription = _a.subscription
  subscription.add apiSubscription

  global.require = (modulePath) ->
    if modulePath == 'sourcegraph'
      return extensionAPI
    # All other requires/imports in the extension's code should not reach here because their JS
    # bundler should have resolved them locally.
    throw new Error('require: module not found: ' + modulePath)
    return

  subscription.add ->

    global.require = ->
      # Prevent callers from attempting to access the extension API after it was
      # unsubscribed.
      throw new Error('require: Sourcegraph extension API was unsubscribed')
      return

    return
  {
    subscription: subscription
    extensionAPI: extensionAPI
    extensionHostAPI: extensionHostAPI
  }

createExtensionAPI = (initData, endpoints) ->
  _a = undefined
  _this = this
  subscription = new (rxjs_1.Subscription)
  # EXTENSION HOST WORKER
  util_1.registerComlinkTransferHandlers()

  ###* Proxy to main thread ###

  proxy = comlink.proxy(endpoints.proxy)
  # For debugging/tests.

  sync = ->
    tslib_1.__awaiter _this, undefined, undefined, ->
      tslib_1.__generator this, (_a) ->
        switch _a.label
          when 0
            return [
              4
              proxy.ping()
            ]
          when 1
            _a.sent()
            return [ 2 ]
        return

  context = new (context_1.ExtContext)(proxy.context)
  documents = new (documents_1.ExtDocuments)(sync)
  extensions = new (extensions_1.ExtExtensions)
  subscription.add extensions
  roots = new (roots_1.ExtRoots)
  windows = new (windows_1.ExtWindows)(proxy, documents)
  views = new (views_1.ExtViews)(proxy.views)
  configuration = new (configuration_1.ExtConfiguration)(proxy.configuration)
  languageFeatures = new (languageFeatures_1.ExtLanguageFeatures)(proxy.languageFeatures, documents)
  search = new (search_1.ExtSearch)(proxy.search)
  commands = new (commands_1.ExtCommands)(proxy.commands)
  content = new (content_1.ExtContent)(proxy.content)
  # Expose the extension host API to the client (main thread)
  extensionHostAPI = _a = {}
  _a[comlink.proxyValueSymbol] = true

  _a.ping = ->
    'pong'

_a.configuration = configuration
  _a.documents = documents
  _a.extensions = extensions
  _a.roots = roots
  _a.windows = windows
  _a
  # Expose the extension API to extensions
  # "redefines" everything instead of exposing internal Ext* classes directly so as to:
  # - Avoid exposing private methods to extensions
  # - Avoid exposing proxy.* to extensions, which gives access to the main thread
  extensionAPI = 
    URI: URL
    Position: extension_api_classes_1.Position
    Range: extension_api_classes_1.Range
    Selection: extension_api_classes_1.Selection
    Location: extension_api_classes_1.Location
    MarkupKind: extension_api_classes_1.MarkupKind
    NotificationType: notifications_1.NotificationType
    app:
      activeWindowChanges: windows.activeWindowChanges
      activeWindow: ->
        windows.activeWindow
      windows: ->
        windows.getAll()
      createPanelView: (id) ->
        views.createPanelView id
      createDecorationType: decorations_1.createDecorationType
    workspace:
      textDocuments: ->
        documents.getAll()
      onDidOpenTextDocument: documents.openedTextDocuments
      openedTextDocuments: documents.openedTextDocuments
      roots: ->
        roots.getAll()
      onDidChangeRoots: roots.changes
      rootChanges: roots.changes
    configuration: Object.assign(configuration.changes.asObservable(), get: ->
      configuration.get()
    )
    languages:
      registerHoverProvider: (selector, provider) ->
        languageFeatures.registerHoverProvider selector, provider
      registerDefinitionProvider: (selector, provider) ->
        languageFeatures.registerDefinitionProvider selector, provider
      registerTypeDefinitionProvider: ->
        console.warn 'sourcegraph.languages.registerTypeDefinitionProvider was removed. Use sourcegraph.languages.registerLocationProvider instead.'
        { unsubscribe: ->
          undefined
 }
      registerImplementationProvider: ->
        console.warn 'sourcegraph.languages.registerImplementationProvider was removed. Use sourcegraph.languages.registerLocationProvider instead.'
        { unsubscribe: ->
          undefined
 }
      registerReferenceProvider: (selector, provider) ->
        languageFeatures.registerReferenceProvider selector, provider
      registerLocationProvider: (id, selector, provider) ->
        languageFeatures.registerLocationProvider id, selector, provider
      registerCompletionItemProvider: (selector, provider) ->
        languageFeatures.registerCompletionItemProvider selector, provider
    search: registerQueryTransformer: (provider) ->
      search.registerQueryTransformer provider
    commands:
      registerCommand: (command, callback) ->
        commands.registerCommand
          command: command
          callback: callback
      executeCommand: (command) ->
        args = []
        _i = 1
        while _i < arguments.length
          args[_i - 1] = arguments[_i]
          _i++
        commands.executeCommand command, args
    content: registerLinkPreviewProvider: (urlMatchPattern, provider) ->
      content.registerLinkPreviewProvider urlMatchPattern, provider
    internal:
      sync: sync
      updateContext: (updates) ->
        context.updateContext updates
      sourcegraphURL: new URL(initData.sourcegraphURL)
      clientApplication: initData.clientApplication
  {
    extensionHostAPI: extensionHostAPI
    extensionAPI: extensionAPI
    subscription: subscription
  }

'use strict'
exports.__esModule = true
tslib_1 = require('tslib')
comlink = tslib_1.__importStar(require('@sourcegraph/comlink'))
extension_api_classes_1 = require('@sourcegraph/extension-api-classes')
rxjs_1 = require('rxjs')
notifications_1 = require('../client/services/notifications')
commands_1 = require('./api/commands')
configuration_1 = require('./api/configuration')
content_1 = require('./api/content')
context_1 = require('./api/context')
decorations_1 = require('./api/decorations')
documents_1 = require('./api/documents')
extensions_1 = require('./api/extensions')
languageFeatures_1 = require('./api/languageFeatures')
roots_1 = require('./api/roots')
search_1 = require('./api/search')
views_1 = require('./api/views')
windows_1 = require('./api/windows')
util_1 = require('../util')
exports.startExtensionHost = startExtensionHost
