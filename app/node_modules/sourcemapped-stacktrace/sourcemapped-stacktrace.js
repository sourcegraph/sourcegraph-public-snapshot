/*
 * sourcemapped-stacktrace.js
 * created by James Salter <iteration@gmail.com> (2014)
 *
 * https://github.com/novocaine/sourcemapped-stacktrace
 *
 * Licensed under the New BSD license. See LICENSE or:
 * http://opensource.org/licenses/BSD-3-Clause
 */

/*global define */

// note we only include source-map-consumer, not the whole source-map library,
// which includes gear for generating source maps that we don't need
define(['./node_modules/source-map/lib/source-map/source-map-consumer'],
function(source_map_consumer) {
  /**
   * Re-map entries in a stacktrace using sourcemaps if available.
   *
   * @param {Array} stack - Array of strings from the browser's stack
   *                        representation. Currently only Chrome
   *                        format is supported.
   * @param {function} done - Callback invoked with the transformed stacktrace
   *                          (an Array of Strings) passed as the first
   *                          argument
   */
  var mapStackTrace = function(stack, done) {
    var lines;
    var mapForUri = {};
    var rows = {};
    var fields;
    var uri;
    var expected_fields;
    var regex;

    var fetcher = new Fetcher(function() {
      var result = processSourceMaps(lines, rows, fetcher.mapForUri);
      done(result);
    });

    if (isChrome()) {
      regex = /^ +at.+\((.*):([0-9]+):([0-9]+)/;
      expected_fields = 4;
      // (skip first line containing exception message)
      skip_lines = 1;
    } else if (isFirefox()) {
      regex = /@(.*):([0-9]+):([0-9]+)/;
      expected_fields = 4;
      skip_lines = 0;
    } else {
      throw new Error("unknown browser :(");
    }

    lines = stack.split("\n").slice(skip_lines);

    for (var i=0; i < lines.length; i++) {
      fields = lines[i].match(regex);
      if (fields && fields.length === expected_fields) {
        rows[i] = fields;
        uri = fields[1];
        fetcher.fetchScript(uri);
      }
    }
  };

  var isChrome = function() {
    return navigator.userAgent.toLowerCase().indexOf('chrome') > -1;
  };
  
  var isFirefox = function() {
    return navigator.userAgent.toLowerCase().indexOf('firefox') > -1;
  };
  var Fetcher = function(done) {
    this.sem = 0;
    this.mapForUri = {};
    this.done = done;
  };

  Fetcher.prototype.fetchScript = function(uri) {
    if (!(uri in this.mapForUri)) {
      this.sem++;
      this.mapForUri[uri] = null;
    } else {
      return;
    }

    var xhr = createXMLHTTPObject();
    var that = this;
    xhr.onreadystatechange = function(e) {
      that.onScriptLoad.call(that, e, uri);
    };
    xhr.open("GET", uri, true);
    xhr.send();
  };

  var absUrlRegex = new RegExp('^(?:[a-z]+:)?//', 'i');

  Fetcher.prototype.onScriptLoad = function(e, uri) {
    if (e.target.readyState !== 4) {
      return;
    }

    if (e.target.status === 200 ||
      (uri.slice(0, 7) === "file://" && e.target.status === 0))
    {
      // find .map in file.
      //
      // attempt to find it at the very end of the file, but tolerate trailing
      // whitespace inserted by some packers.
      var match = e.target.responseText.match("//# sourceMappingURL=(.*)[\\s]*$", "m");
      if (match && match.length === 2) {
        // get the map
        var mapUri = match[1];

        var embeddedSourceMap = mapUri.match("data:application/json;base64,(.*)");

        if (embeddedSourceMap && embeddedSourceMap[1]) {
          this.mapForUri[uri] = atob(embeddedSourceMap[1]);
          this.done(this.mapForUri);
        } else {
          if (!absUrlRegex.test(mapUri)) {
            // relative url; according to sourcemaps spec is 'source origin'
            var origin;
            var lastSlash = uri.lastIndexOf('/');
            if (lastSlash !== -1) {
              origin = uri.slice(0, lastSlash + 1);
              mapUri = origin + mapUri;
              // note if lastSlash === -1, actual script uri has no slash
              // somehow, so no way to use it as a prefix... we give up and try
              // as absolute
            }
          }

          var xhrMap = createXMLHTTPObject();
          var that = this;
          xhrMap.onreadystatechange = function() {
            if (xhrMap.readyState === 4) {
              that.sem--;
              if (xhrMap.status === 200 ||
                (mapUri.slice(0, 7) === "file://" && xhrMap.status === 0)) {
                that.mapForUri[uri] = xhrMap.responseText;
              }
              if (that.sem === 0) {
                that.done(that.mapForUri);
              }
            }
          };

          xhrMap.open("GET", mapUri, true);
          xhrMap.send();
        }
      } else {
        // no map
        this.sem--;
      }
    } else {
      // HTTP error fetching uri of the script
      this.sem--;
    }

    if (this.sem === 0) {
      this.done(this.mapForUri);
    }
  };

  var processSourceMaps = function(lines, rows, mapForUri) {
    var result = [];
    var map;
    for (var i=0; i < lines.length; i++) {
      var row = rows[i];
      if (row) {
        var uri = row[1];
        var line = parseInt(row[2], 10);
        var column = parseInt(row[3], 10);
        map = mapForUri[uri];

        if (map) {
          // we think we have a map for that uri. call source-map library
          var smc = new source_map_consumer.SourceMapConsumer(map);
          var origPos = smc.originalPositionFor(
            { line: line, column: column });
          result.push(formatOriginalPosition(origPos.source,
            origPos.line, origPos.column, origPos.name));
        } else {
          // we can't find a map for that url, but we parsed the row.
          // reformat unchanged line for consistency with the sourcemapped
          // lines.
          result.push(formatOriginalPosition(uri, line, column, "(unknown)"));
        }
      } else {
        // we weren't able to parse the row, push back what we were given
        result.push(lines[i]);
      }
    }

    return result;
  };

  var formatOriginalPosition = function(source, line, column, name) {
    // mimic chrome's format
    return "    at " + (name ? name : "(unknown)") +
      " (" + source + ":" + line + ":" + column + ")";
  };

  // xmlhttprequest boilerplate
  var XMLHttpFactories = [
	function () {return new XMLHttpRequest();},
	function () {return new ActiveXObject("Msxml2.XMLHTTP");},
	function () {return new ActiveXObject("Msxml3.XMLHTTP");},
	function () {return new ActiveXObject("Microsoft.XMLHTTP");}
  ];

  function createXMLHTTPObject() {
      var xmlhttp = false;
      for (var i=0;i<XMLHttpFactories.length;i++) {
          try {
              xmlhttp = XMLHttpFactories[i]();
          }
          catch (e) {
              continue;
          }
          break;
      }
      return xmlhttp;
  }

  return {
    mapStackTrace: mapStackTrace
  }
});
