var window = require("global/window")
var test = require("tape")

var xhr = require("../index.js")

test("constructs and calls callback without throwing", function (assert) {
    xhr({}, function (err, resp, body) {
        assert.ok(true, "got here")
        assert.end()
    })
})

test("[func] Can GET a url (cross-domain)", function (assert) {
    xhr({
        uri: "http://www.mocky.io/v2/55a02cb72651260b1a94f024",
        useXDR: true
    }, function (err, resp, body) {
        assert.ifError(err, "no err")
        assert.equal(resp.statusCode, 200)
        assert.equal(typeof resp.rawRequest, "object")
        assert.notEqual(resp.body.length, 0)
        assert.equal(resp.body,'{"a":1}')
        assert.notEqual(body.length, 0)
        assert.end()
    })
})

test("[func] Returns http error responses like npm's request (cross-domain)", function (assert) {
    if (!window.XDomainRequest) {
        xhr({
            uri: "http://www.mocky.io/v2/55a02d63265126221a94f025",
            useXDR: true
        }, function (err, resp, body) {
            assert.ifError(err, "no err")
            assert.equal(resp.statusCode, 404)
            assert.equal(typeof resp.rawRequest, "object")
            assert.end()
        })
    } else {
        assert.end();
    }
})

test("[func] Times out to an error ", function (assert) {
    xhr({
        timeout: 1,
        uri: "/tests-bundle.js?should-take-a-bit-to-parse=1&" + (new Array(300)).join("cachebreaker=" + Math.random().toFixed(5) + "&")
    }, function (err, resp, body) {
        assert.ok(err instanceof Error, "should return error")
        assert.equal(err.message, "XMLHttpRequest timeout")
        assert.equal(err.code, "ETIMEDOUT")
        assert.equal(resp.statusCode, 0)
        assert.end()
    })
})

test("withCredentials option", function (assert) {
    if (!window.XDomainRequest) {
        var req = xhr({}, function () {})
        assert.ok(!req.withCredentials,
            "withCredentials not true"
        )
        req = xhr({
            withCredentials: true
        }, function () {})
        assert.ok(
            req.withCredentials,
            "withCredentials set to true"
        )
    }
    assert.end()
})

test("withCredentials ignored when using synchronous requests", function (assert) {
    if (!window.XDomainRequest) {
        var req = xhr({
            withCredentials: true,
            sync: true
        }, function () {})
        assert.ok(!req.withCredentials,
            "sync overrides withCredentials"
        )
    }
    assert.end()
})

test("XDR usage (run on IE8 or 9)", function (assert) {
    var req = xhr({
        useXDR: true,
        uri: window.location.href,
    }, function () {})

    assert.ok(!window.XDomainRequest || window.XDomainRequest === req.constructor,
        "Uses XDR when told to"
    )


    if (!!window.XDomainRequest) {
        assert.throws(function () {
            xhr({
                useXDR: true,
                uri: window.location.href,
                headers: {
                    "foo": "bar"
                }
            }, function () {})
        }, true, "Throws when trying to send headers with XDR")
    }
    assert.end()
})

test("handles errorFunc call with no arguments provided", function (assert) {
    var req = xhr({}, function (err) {
        assert.ok(err instanceof Error, "callback should get an error")
        assert.equal(err.message, "Unknown XMLHttpRequest Error", "error message incorrect")
    })
    assert.doesNotThrow(function () {
        req.onerror()
    }, "should not throw when error handler called without arguments")
    assert.end()

})

test("constructs and calls callback without throwing", function (assert) {
    assert.throws(function () {
        xhr({})
    }, "callback is not optional")
    assert.end()
})

test("XHR can be overridden", function (assert) {
  var xhrs = 0
  var noop = function () {}
  var fakeXHR = function () {
    xhrs++
    this.open = this.send = noop
  }
  var xdrs = 0
  var fakeXDR = function () {
    xdrs++
    this.open = this.send = noop
  }
  xhr.XMLHttpRequest = fakeXHR
  xhr({}, function () {})
  assert.equal(xhrs, 1, "created the custom XHR")

  xhr.XDomainRequest = fakeXDR
  xhr({ useXDR: true }, function () {});
  assert.equal(xdrs, 1, "created the custom XDR")
  assert.end()
})
