(function(window, document) {
  var amplitude = window.amplitude || {'_q':[]};
  var as = document.createElement('script');
  as.type = 'text/javascript';
  as.async = true;
  as.src = 'https://d24n15hnbwhuhn.cloudfront.net/libs/amplitude-2.12.1-min.gz.js';
  as.onload = function() {window.amplitude.runQueuedFunctions();};
  var s = document.getElementsByTagName('script')[0];
  s.parentNode.insertBefore(as, s);
  function proxy(obj, fn) {
    obj.prototype[fn] = function() {
      this._q.push([fn].concat(Array.prototype.slice.call(arguments, 0))); return this;
    };
  }
  var Identify = function() {this._q = []; return this;};
  var identifyFuncs = ['add', 'append', 'clearAll', 'prepend', 'set', 'setOnce', 'unset'];
  for (var i = 0; i < identifyFuncs.length; i++) {proxy(Identify,identifyFuncs[i]);}
  amplitude.Identify = Identify;
  var Revenue = function() {this._q = []; return this;};
  var revenueFuncs = ['setProductId', 'setQuantity', 'setPrice', 'setRevenueType', 'setEventProperties'];
  for (var j = 0; j < revenueFuncs.length; j++) {proxy(Revenue, revenueFuncs[j]);}
  amplitude.Revenue = Revenue;
  var funcs = ['init', 'logEvent', 'logRevenue', 'setUserId', 'setUserProperties',
               'setOptOut', 'setVersionName', 'setDomain', 'setDeviceId',
               'setGlobalUserProperties', 'identify', 'clearUserProperties',
               'setGroup', 'logRevenueV2', 'regenerateDeviceId'];
  function setUpProxy(instance) {
    function proxyMain(fn) {
      instance[fn] = function() {
        instance._q.push([fn].concat(Array.prototype.slice.call(arguments, 0)));
      };
    }
    for (var k = 0; k < funcs.length; k++) {proxyMain(funcs[k]);}
  }
  setUpProxy(amplitude);
  window.amplitude = amplitude;
})(window, document);
