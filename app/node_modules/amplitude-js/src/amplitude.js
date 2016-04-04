var cookieStorage = require('./cookiestorage');
var getUtmData = require('./utm');
var Identify = require('./identify');
var JSON = require('json'); // jshint ignore:line
var localStorage = require('./localstorage');  // jshint ignore:line
var md5 = require('JavaScript-MD5');
var object = require('object');
var Request = require('./xhr');
var type = require('./type');
var UAParser = require('ua-parser-js');
var utils = require('./utils');
var UUID = require('./uuid');
var version = require('./version');
var DEFAULT_OPTIONS = require('./options');

var IDENTIFY_EVENT = '$identify';
var API_VERSION = 2;
var MAX_STRING_LENGTH = 1024;
var LocalStorageKeys = {
  LAST_EVENT_ID: 'amplitude_lastEventId',
  LAST_EVENT_TIME: 'amplitude_lastEventTime',
  LAST_IDENTIFY_ID: 'amplitude_lastIdentifyId',
  LAST_SEQUENCE_NUMBER: 'amplitude_lastSequenceNumber',
  REFERRER: 'amplitude_referrer',
  SESSION_ID: 'amplitude_sessionId',

  // Used in cookie as well
  DEVICE_ID: 'amplitude_deviceId',
  OPT_OUT: 'amplitude_optOut',
  USER_ID: 'amplitude_userId'
};

/*
 * Amplitude API
 */
var Amplitude = function() {
  this._unsentEvents = [];
  this._unsentIdentifys = [];
  this._ua = new UAParser(navigator.userAgent).getResult();
  this.options = object.merge({}, DEFAULT_OPTIONS);
  this.cookieStorage = new cookieStorage().getStorage();
  this._q = []; // queue for proxied functions before script load
};

Amplitude.prototype._eventId = 0;
Amplitude.prototype._identifyId = 0;
Amplitude.prototype._sequenceNumber = 0;
Amplitude.prototype._sending = false;
Amplitude.prototype._lastEventTime = null;
Amplitude.prototype._sessionId = null;
Amplitude.prototype._newSession = false;
Amplitude.prototype._updateScheduled = false;

Amplitude.prototype.Identify = Identify;

/**
 * Initializes Amplitude.
 * apiKey The API Key for your app
 * opt_userId An identifier for this user
 * opt_config Configuration options
 *   - saveEvents (boolean) Whether to save events to local storage. Defaults to true.
 *   - includeUtm (boolean) Whether to send utm parameters with events. Defaults to false.
 *   - includeReferrer (boolean) Whether to send referrer info with events. Defaults to false.
 */
Amplitude.prototype.init = function(apiKey, opt_userId, opt_config, callback) {
  try {
    this.options.apiKey = apiKey;
    if (opt_config) {
      if (opt_config.saveEvents !== undefined) {
        this.options.saveEvents = !!opt_config.saveEvents;
      }
      if (opt_config.domain !== undefined) {
        this.options.domain = opt_config.domain;
      }
      if (opt_config.includeUtm !== undefined) {
        this.options.includeUtm = !!opt_config.includeUtm;
      }
      if (opt_config.includeReferrer !== undefined) {
        this.options.includeReferrer = !!opt_config.includeReferrer;
      }
      if (opt_config.batchEvents !== undefined) {
        this.options.batchEvents = !!opt_config.batchEvents;
      }
      this.options.platform = opt_config.platform || this.options.platform;
      this.options.language = opt_config.language || this.options.language;
      this.options.sessionTimeout = opt_config.sessionTimeout || this.options.sessionTimeout;
      this.options.uploadBatchSize = opt_config.uploadBatchSize || this.options.uploadBatchSize;
      this.options.eventUploadThreshold = opt_config.eventUploadThreshold || this.options.eventUploadThreshold;
      this.options.savedMaxCount = opt_config.savedMaxCount || this.options.savedMaxCount;
      this.options.eventUploadPeriodMillis = opt_config.eventUploadPeriodMillis || this.options.eventUploadPeriodMillis;
    }

    this.cookieStorage.options({
      expirationDays: this.options.cookieExpiration,
      domain: this.options.domain
    });
    this.options.domain = this.cookieStorage.options().domain;

    _upgradeCookeData(this);
    _loadCookieData(this);

    this.options.deviceId = (opt_config && opt_config.deviceId !== undefined &&
        opt_config.deviceId !== null && opt_config.deviceId) ||
        this.options.deviceId || UUID();
    this.options.userId = (opt_userId !== undefined && opt_userId !== null && opt_userId) || this.options.userId || null;

    var now = new Date().getTime();
    if (!this._sessionId || !this._lastEventTime || now - this._lastEventTime > this.options.sessionTimeout) {
      this._newSession = true;
      this._sessionId = now;
    }
    this._lastEventTime = now;
    _saveCookieData(this);

    //utils.log('initialized with apiKey=' + apiKey);
    //opt_userId !== undefined && opt_userId !== null && utils.log('initialized with userId=' + opt_userId);

    if (this.options.saveEvents) {
      this._unsentEvents = this._loadSavedUnsentEvents(this.options.unsentKey) || this._unsentEvents;
      this._unsentIdentifys = this._loadSavedUnsentEvents(this.options.unsentIdentifyKey) || this._unsentIdentifys;

      // validate event properties for unsent events
      for (var i = 0; i < this._unsentEvents.length; i++) {
        var eventProperties = this._unsentEvents[i].event_properties;
        this._unsentEvents[i].event_properties = utils.validateProperties(eventProperties);
      }

      this._sendEventsIfReady();
    }

    if (this.options.includeUtm) {
      this._initUtmData();
    }

    if (this.options.includeReferrer) {
      this._saveReferrer(this._getReferrer());
    }
  } catch (e) {
    utils.log(e);
  }

  if (callback && type(callback) === 'function') {
    callback();
  }
};

Amplitude.prototype.runQueuedFunctions = function () {
  for (var i = 0; i < this._q.length; i++) {
    var fn = this[this._q[i][0]];
    if (fn && type(fn) === 'function') {
      fn.apply(this, this._q[i].slice(1));
    }
  }
  this._q = []; // clear function queue after running
};

Amplitude.prototype._apiKeySet = function(methodName) {
  if (!this.options.apiKey) {
    utils.log('apiKey cannot be undefined or null, set apiKey with init() before calling ' + methodName);
    return false;
  }
  return true;
};

Amplitude.prototype._loadSavedUnsentEvents = function(unsentKey) {
  var savedUnsentEventsString = this._getFromStorage(localStorage, unsentKey);
  if (savedUnsentEventsString) {
    try {
      return JSON.parse(savedUnsentEventsString);
    } catch (e) {
      // utils.log(e);
    }
  }
  return null;
};

Amplitude.prototype.isNewSession = function() {
  return this._newSession;
};

Amplitude.prototype.getSessionId = function() {
  return this._sessionId;
};

Amplitude.prototype.nextEventId = function() {
  this._eventId++;
  return this._eventId;
};

Amplitude.prototype.nextIdentifyId = function() {
  this._identifyId++;
  return this._identifyId;
};

Amplitude.prototype.nextSequenceNumber = function() {
  this._sequenceNumber++;
  return this._sequenceNumber;
};

// returns the number of unsent events and identifys
Amplitude.prototype._unsentCount = function() {
  return this._unsentEvents.length + this._unsentIdentifys.length;
};

// returns true if sendEvents called immediately
Amplitude.prototype._sendEventsIfReady = function(callback) {
  if (this._unsentCount() === 0) {
    return false;
  }

  if (!this.options.batchEvents) {
    this.sendEvents(callback);
    return true;
  }

  if (this._unsentCount() >= this.options.eventUploadThreshold) {
    this.sendEvents(callback);
    return true;
  }

  if (!this._updateScheduled) {
    this._updateScheduled = true;
    setTimeout(
      function() {
        this._updateScheduled = false;
        this.sendEvents();
      }.bind(this), this.options.eventUploadPeriodMillis
    );
  }

  return false;
};

// storage argument allows for localStorage and sessionStorage
Amplitude.prototype._getFromStorage = function(storage, key) {
  return storage.getItem(key);
};

// storage argument allows for localStorage and sessionStorage
Amplitude.prototype._setInStorage = function(storage, key, value) {
  storage.setItem(key, value);
};

/*
 * cookieData (deviceId, userId, optOut, sessionId, lastEventTime, eventId, identifyId, sequenceNumber)
 * can be stored in many different places (localStorage, cookie, etc).
 * Need to unify all sources into one place with a one-time upgrade/migration for the defaultInstance.
 */
var _upgradeCookeData = function(scope) {
  // skip if migration already happened
  var cookieData = scope.cookieStorage.get(scope.options.cookieName);
  if (cookieData && cookieData.deviceId && cookieData.sessionId && cookieData.lastEventTime) {
    return;
  }

  var _getAndRemoveFromLocalStorage = function(key) {
    var value = localStorage.getItem(key);
    localStorage.removeItem(key);
    return value;
  };

  // in v2.6.0, deviceId, userId, optOut was migrated to localStorage with keys + first 6 char of apiKey
  var apiKeySuffix = '_' + scope.options.apiKey.slice(0, 6);
  var localStorageDeviceId = _getAndRemoveFromLocalStorage(LocalStorageKeys.DEVICE_ID + apiKeySuffix);
  var localStorageUserId = _getAndRemoveFromLocalStorage(LocalStorageKeys.USER_ID + apiKeySuffix);
  var localStorageOptOut = _getAndRemoveFromLocalStorage(LocalStorageKeys.OPT_OUT + apiKeySuffix);
  if (localStorageOptOut !== null && localStorageOptOut !== undefined) {
    localStorageOptOut = String(localStorageOptOut) === 'true'; // convert to boolean
  }

  // pre-v2.7.0 event and session meta-data was stored in localStorage. move to cookie for sub-domain support
  var localStorageSessionId = parseInt(_getAndRemoveFromLocalStorage(LocalStorageKeys.SESSION_ID));
  var localStorageLastEventTime = parseInt(_getAndRemoveFromLocalStorage(LocalStorageKeys.LAST_EVENT_TIME));
  var localStorageEventId = parseInt(_getAndRemoveFromLocalStorage(LocalStorageKeys.LAST_EVENT_ID));
  var localStorageIdentifyId = parseInt(_getAndRemoveFromLocalStorage(LocalStorageKeys.LAST_IDENTIFY_ID));
  var localStorageSequenceNumber = parseInt(_getAndRemoveFromLocalStorage(LocalStorageKeys.LAST_SEQUENCE_NUMBER));

  var _getFromCookie = function(key) {
    return cookieData && cookieData[key];
  };
  scope.options.deviceId = _getFromCookie('deviceId') || localStorageDeviceId;
  scope.options.userId = _getFromCookie('userId') || localStorageUserId;
  scope._sessionId = _getFromCookie('sessionId') || localStorageSessionId || scope._sessionId;
  scope._lastEventTime = _getFromCookie('lastEventTime') || localStorageLastEventTime || scope._lastEventTime;
  scope._eventId = _getFromCookie('eventId') || localStorageEventId || scope._eventId;
  scope._identifyId = _getFromCookie('identifyId') || localStorageIdentifyId || scope._identifyId;
  scope._sequenceNumber = _getFromCookie('sequenceNumber') || localStorageSequenceNumber || scope._sequenceNumber;

  // optOut is a little trickier since it is a boolean
  scope.options.optOut = localStorageOptOut || false;
  if (cookieData && cookieData.optOut !== undefined && cookieData.optOut !== null) {
    scope.options.optOut = String(cookieData.optOut) === 'true';
  }

  _saveCookieData(scope);
};

var _loadCookieData = function(scope) {
  var cookieData = scope.cookieStorage.get(scope.options.cookieName);
  if (cookieData) {
    if (cookieData.deviceId) {
      scope.options.deviceId = cookieData.deviceId;
    }
    if (cookieData.userId) {
      scope.options.userId = cookieData.userId;
    }
    if (cookieData.optOut !== null && cookieData.optOut !== undefined) {
      scope.options.optOut = cookieData.optOut;
    }
    if (cookieData.sessionId) {
      scope._sessionId = parseInt(cookieData.sessionId);
    }
    if (cookieData.lastEventTime) {
      scope._lastEventTime = parseInt(cookieData.lastEventTime);
    }
    if (cookieData.eventId) {
      scope._eventId = parseInt(cookieData.eventId);
    }
    if (cookieData.identifyId) {
      scope._identifyId = parseInt(cookieData.identifyId);
    }
    if (cookieData.sequenceNumber) {
      scope._sequenceNumber = parseInt(cookieData.sequenceNumber);
    }
  }
};

var _saveCookieData = function(scope) {
  scope.cookieStorage.set(scope.options.cookieName, {
    deviceId: scope.options.deviceId,
    userId: scope.options.userId,
    optOut: scope.options.optOut,
    sessionId: scope._sessionId,
    lastEventTime: scope._lastEventTime,
    eventId: scope._eventId,
    identifyId: scope._identifyId,
    sequenceNumber: scope._sequenceNumber
  });
};

/**
 * Parse the utm properties out of cookies and query for adding to user properties.
 */
Amplitude.prototype._initUtmData = function(queryParams, cookieParams) {
  queryParams = queryParams || location.search;
  cookieParams = cookieParams || this.cookieStorage.get('__utmz');
  this._utmProperties = getUtmData(cookieParams, queryParams);
};

Amplitude.prototype._getReferrer = function() {
  return document.referrer;
};

Amplitude.prototype._getReferringDomain = function(referrer) {
  if (referrer === null || referrer === undefined || referrer === '') {
    return null;
  }
  var parts = referrer.split('/');
  if (parts.length >= 3) {
    return parts[2];
  }
  return null;
};

// since user properties are propagated on the server, only send once per session, don't need to send with every event
Amplitude.prototype._saveReferrer = function(referrer) {
  if (referrer === null || referrer === undefined || referrer === '') {
    return;
  }

  // always setOnce initial referrer
  var referring_domain = this._getReferringDomain(referrer);
  var identify = new Identify().setOnce('initial_referrer', referrer);
  identify.setOnce('initial_referring_domain', referring_domain);

  // only save referrer if not already in session storage or if storage disabled
  var hasSessionStorage = false;
  try {
    if (window.sessionStorage) {
      hasSessionStorage = true;
    }
  } catch (e) {
    // utils.log(e); // sessionStorage disabled
  }

  if ((hasSessionStorage && !(this._getFromStorage(sessionStorage, LocalStorageKeys.REFERRER))) || !hasSessionStorage) {
    identify.set('referrer', referrer).set('referring_domain', referring_domain);

    if (hasSessionStorage) {
      this._setInStorage(sessionStorage, LocalStorageKeys.REFERRER, referrer);
    }
  }

  this.identify(identify);
};

Amplitude.prototype.saveEvents = function() {
  if (!this._apiKeySet('saveEvents()')) {
    return;
  }

  try {
    this._setInStorage(localStorage, this.options.unsentKey, JSON.stringify(this._unsentEvents));
    this._setInStorage(localStorage, this.options.unsentIdentifyKey, JSON.stringify(this._unsentIdentifys));
  } catch (e) {
    // utils.log(e);
  }
};

Amplitude.prototype.setDomain = function(domain) {
  if (!this._apiKeySet('setDomain()')) {
    return;
  }

  try {
    this.cookieStorage.options({
      domain: domain
    });
    this.options.domain = this.cookieStorage.options().domain;
    _loadCookieData(this);
    _saveCookieData(this);
    // utils.log('set domain=' + domain);
  } catch (e) {
    utils.log(e);
  }
};

Amplitude.prototype.setUserId = function(userId) {
  if (!this._apiKeySet('setUserId()')) {
    return;
  }

  try {
    this.options.userId = (userId !== undefined && userId !== null && ('' + userId)) || null;
    _saveCookieData(this);
    // utils.log('set userId=' + userId);
  } catch (e) {
    utils.log(e);
  }
};

Amplitude.prototype.setOptOut = function(enable) {
  if (!this._apiKeySet('setOptOut()')) {
    return;
  }

  try {
    this.options.optOut = enable;
    _saveCookieData(this);
    // utils.log('set optOut=' + enable);
  } catch (e) {
    utils.log(e);
  }
};

Amplitude.prototype.setDeviceId = function(deviceId) {
  if (!this._apiKeySet('setDeviceId()')) {
    return;
  }

  try {
    if (deviceId) {
      this.options.deviceId = ('' + deviceId);
      _saveCookieData(this);
    }
  } catch (e) {
    utils.log(e);
  }
};

Amplitude.prototype.setUserProperties = function(userProperties) {
  if (!this._apiKeySet('setUserProperties()')) {
    return;
  }
  // convert userProperties into an identify call
  var identify = new Identify();
  for (var property in userProperties) {
    if (userProperties.hasOwnProperty(property)) {
      identify.set(property, userProperties[property]);
    }
  }
  this.identify(identify);
};

// Clearing user properties is irreversible!
Amplitude.prototype.clearUserProperties = function(){
  if (!this._apiKeySet('clearUserProperties()')) {
    return;
  }

  var identify = new Identify();
  identify.clearAll();
  this.identify(identify);
};

Amplitude.prototype.identify = function(identify, callback) {
  if (!this._apiKeySet('identify()')) {
    if (callback && type(callback) === 'function') {
      callback(0, 'No request sent');
    }
    return;
  }

  if (type(identify) === 'object' && '_q' in identify) {
    var instance = new Identify();
    // Apply the queued commands
    for (var i = 0; i < identify._q.length; i++) {
        var fn = instance[identify._q[i][0]];
        if (fn && type(fn) === 'function') {
          fn.apply(instance, identify._q[i].slice(1));
        }
    }
    identify = instance;
  }

  if (identify instanceof Identify && Object.keys(identify.userPropertiesOperations).length > 0) {
    this._logEvent(IDENTIFY_EVENT, null, null, identify.userPropertiesOperations, callback);
  } else if (callback && type(callback) === 'function') {
    callback(0, 'No request sent');
  }
};

Amplitude.prototype.setVersionName = function(versionName) {
  try {
    this.options.versionName = versionName;
    // utils.log('set versionName=' + versionName);
  } catch (e) {
    utils.log(e);
  }
};

// truncate string values in event and user properties so that request size does not get too large
Amplitude.prototype._truncate = function(value) {
  if (type(value) === 'array') {
    for (var i = 0; i < value.length; i++) {
      value[i] = this._truncate(value[i]);
    }
  } else if (type(value) === 'object') {
    for (var key in value) {
      if (value.hasOwnProperty(key)) {
        value[key] = this._truncate(value[key]);
      }
    }
  } else {
    value = _truncateValue(value);
  }

  return value;
};

var _truncateValue = function(value) {
  if (type(value) === 'string') {
    return value.length > MAX_STRING_LENGTH ? value.substring(0, MAX_STRING_LENGTH) : value;
  }
  return value;
};

/**
 * Private logEvent method. Keeps apiProperties from being publicly exposed.
 */
Amplitude.prototype._logEvent = function(eventType, eventProperties, apiProperties, userProperties, callback) {
  if (type(callback) !== 'function') {
    callback = null;
  }

  _loadCookieData(this);
  if (!eventType || this.options.optOut) {
    if (callback) {
      callback(0, 'No request sent');
    }
    return;
  }
  try {
    var eventId;
    if (eventType === IDENTIFY_EVENT) {
      eventId = this.nextIdentifyId();
    } else {
      eventId = this.nextEventId();
    }
    var sequenceNumber = this.nextSequenceNumber();
    var eventTime = new Date().getTime();
    var ua = this._ua;
    if (!this._sessionId || !this._lastEventTime || eventTime - this._lastEventTime > this.options.sessionTimeout) {
      this._sessionId = eventTime;
    }
    this._lastEventTime = eventTime;
    _saveCookieData(this);

    userProperties = userProperties || {};
    // Only add utm properties to user properties for events
    if (eventType !== IDENTIFY_EVENT) {
      object.merge(userProperties, this._utmProperties);
    }

    apiProperties = apiProperties || {};
    eventProperties = eventProperties || {};
    var event = {
      device_id: this.options.deviceId,
      user_id: this.options.userId || this.options.deviceId,
      timestamp: eventTime,
      event_id: eventId,
      session_id: this._sessionId || -1,
      event_type: eventType,
      version_name: this.options.versionName || null,
      platform: this.options.platform,
      os_name: ua.browser.name || null,
      os_version: ua.browser.major || null,
      device_model: ua.os.name || null,
      language: this.options.language,
      api_properties: apiProperties,
      event_properties: this._truncate(utils.validateProperties(eventProperties)),
      user_properties: this._truncate(userProperties),
      uuid: UUID(),
      library: {
        name: 'amplitude-js',
        version: version
      },
      sequence_number: sequenceNumber // for ordering events and identifys
      // country: null
    };

    if (eventType === IDENTIFY_EVENT) {
      this._unsentIdentifys.push(event);
      this._limitEventsQueued(this._unsentIdentifys);
    } else {
      this._unsentEvents.push(event);
      this._limitEventsQueued(this._unsentEvents);
    }

    if (this.options.saveEvents) {
      this.saveEvents();
    }

    if (!this._sendEventsIfReady(callback) && callback) {
      callback(0, 'No request sent');
    }

    return eventId;
  } catch (e) {
    utils.log(e);
  }
};

// Remove old events from the beginning of the array if too many
// have accumulated. Don't want to kill memory. Default is 1000 events.
Amplitude.prototype._limitEventsQueued = function(queue) {
  if (queue.length > this.options.savedMaxCount) {
    queue.splice(0, queue.length - this.options.savedMaxCount);
  }
};

Amplitude.prototype.logEvent = function(eventType, eventProperties, callback) {
  if (!this._apiKeySet('logEvent()')) {
    if (callback && type(callback) === 'function') {
      callback(0, 'No request sent');
    }
    return -1;
  }
  return this._logEvent(eventType, eventProperties, null, null, callback);
};

// Test that n is a number or a numeric value.
var _isNumber = function(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
};

Amplitude.prototype.logRevenue = function(price, quantity, product) {
  // Test that the parameters are of the right type.
  if (!this._apiKeySet('logRevenue()') || !_isNumber(price) || quantity !== undefined && !_isNumber(quantity)) {
    // utils.log('Price and quantity arguments to logRevenue must be numbers');
    return -1;
  }

  return this._logEvent('revenue_amount', {}, {
    productId: product,
    special: 'revenue_amount',
    quantity: quantity || 1,
    price: price
  });
};

/**
 * Remove events in storage with event ids up to and including maxEventId. Does
 * a true filter in case events get out of order or old events are removed.
 */
Amplitude.prototype.removeEvents = function (maxEventId, maxIdentifyId) {
  if (maxEventId >= 0) {
    var filteredEvents = [];
    for (var i = 0; i < this._unsentEvents.length; i++) {
      if (this._unsentEvents[i].event_id > maxEventId) {
        filteredEvents.push(this._unsentEvents[i]);
      }
    }
    this._unsentEvents = filteredEvents;
  }

  if (maxIdentifyId >= 0) {
    var filteredIdentifys = [];
    for (var j = 0; j < this._unsentIdentifys.length; j++) {
      if (this._unsentIdentifys[j].event_id > maxIdentifyId) {
        filteredIdentifys.push(this._unsentIdentifys[j]);
      }
    }
    this._unsentIdentifys = filteredIdentifys;
  }
};

Amplitude.prototype.sendEvents = function(callback) {
  if (!this._apiKeySet('sendEvents()')) {
    if (callback && type(callback) === 'function') {
      callback(0, 'No request sent');
    }
    return;
  }

  if (!this._sending && !this.options.optOut && this._unsentCount() > 0) {
    this._sending = true;
    var url = ('https:' === window.location.protocol ? 'https' : 'http') + '://' +
        this.options.apiEndpoint + '/';

    // fetch events to send
    var numEvents = Math.min(this._unsentCount(), this.options.uploadBatchSize);
    var mergedEvents = this._mergeEventsAndIdentifys(numEvents);
    var maxEventId = mergedEvents.maxEventId;
    var maxIdentifyId = mergedEvents.maxIdentifyId;
    var events = JSON.stringify(mergedEvents.eventsToSend);

    var uploadTime = new Date().getTime();
    var data = {
      client: this.options.apiKey,
      e: events,
      v: API_VERSION,
      upload_time: uploadTime,
      checksum: md5(API_VERSION + this.options.apiKey + events + uploadTime)
    };

    var scope = this;
    new Request(url, data).send(function(status, response) {
      scope._sending = false;
      try {
        if (status === 200 && response === 'success') {
          // utils.log('sucessful upload');
          scope.removeEvents(maxEventId, maxIdentifyId);

          // Update the event cache after the removal of sent events.
          if (scope.options.saveEvents) {
            scope.saveEvents();
          }

          // Send more events if any queued during previous send.
          if (!scope._sendEventsIfReady(callback) && callback) {
            callback(status, response);
          }

        } else if (status === 413) {
          // utils.log('request too large');
          // Can't even get this one massive event through. Drop it.
          if (scope.options.uploadBatchSize === 1) {
            // if massive event is identify, still need to drop it
            scope.removeEvents(maxEventId, maxIdentifyId);
          }

          // The server complained about the length of the request.
          // Backoff and try again.
          scope.options.uploadBatchSize = Math.ceil(numEvents / 2);
          scope.sendEvents(callback);

        } else if (callback) { // If server turns something like a 400
          callback(status, response);
        }
      } catch (e) {
        // utils.log('failed upload');
      }
    });
  } else if (callback) {
    callback(0, 'No request sent');
  }
};

Amplitude.prototype._mergeEventsAndIdentifys = function(numEvents) {
  // coalesce events from both queues
  var eventsToSend = [];
  var eventIndex = 0;
  var maxEventId = -1;
  var identifyIndex = 0;
  var maxIdentifyId = -1;

  while (eventsToSend.length < numEvents) {
    var event;

    // case 1: no identifys - grab from events
    if (identifyIndex >= this._unsentIdentifys.length) {
      event = this._unsentEvents[eventIndex++];
      maxEventId = event.event_id;

    // case 2: no events - grab from identifys
    } else if (eventIndex >= this._unsentEvents.length) {
      event = this._unsentIdentifys[identifyIndex++];
      maxIdentifyId = event.event_id;

    // case 3: need to compare sequence numbers
    } else {
      // events logged before v2.5.0 won't have a sequence number, put those first
      if (!('sequence_number' in this._unsentEvents[eventIndex]) ||
          this._unsentEvents[eventIndex].sequence_number <
          this._unsentIdentifys[identifyIndex].sequence_number) {
        event = this._unsentEvents[eventIndex++];
        maxEventId = event.event_id;
      } else {
        event = this._unsentIdentifys[identifyIndex++];
        maxIdentifyId = event.event_id;
      }
    }

    eventsToSend.push(event);
  }

  return {
    eventsToSend: eventsToSend,
    maxEventId: maxEventId,
    maxIdentifyId: maxIdentifyId
  };
};

/**
 *  @deprecated
 */
Amplitude.prototype.setGlobalUserProperties = Amplitude.prototype.setUserProperties;

Amplitude.prototype.__VERSION__ = version;

module.exports = Amplitude;
