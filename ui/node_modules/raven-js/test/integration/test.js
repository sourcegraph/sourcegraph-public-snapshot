/*global assert*/
function iframeExecute(iframe, done, execute, assertCallback) {
    iframe.contentWindow.done = function () {
        try {
            assertCallback(iframe);
            done();
        } catch (e) {
            done(e);
        }
    };
    // use setTimeout so stack trace doesn't go all the way back to mocha test runner
    iframe.contentWindow.eval('window.originalBuiltIns.setTimeout.call(window, ' + execute.toString() + ');');
}

function createIframe(done) {
    var iframe = document.createElement('iframe');
    iframe.style.display = 'none';
    iframe.src = './frame.html';
    iframe.onload = function () {
        done();
    };
    document.body.appendChild(iframe);
    return iframe;
}

var anchor = document.createElement('a');
function parseUrl(url) {
    var out = {pathname: '', origin: '', protocol: ''};
    if (!url)
    anchor.href = url;
    for (var key in out) {
        out[key] = anchor[key];
    }
    return out;
}

describe('integration', function () {

    beforeEach(function (done) {
        this.iframe = createIframe(done);
    });

    afterEach(function () {
        document.body.removeChild(this.iframe);
    });

    describe('API', function () {
        it('should capture Raven.captureMessage', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    Raven.captureMessage('Hello');
                    done();
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    assert.equal(ravenData.message, 'Hello');
                }
            );
        });

        it('should capture Raven.captureException', function (done) {
            var iframe = this.iframe;
            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    try {
                        foo();
                    } catch (e) {
                        Raven.captureException(e);
                    }
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    assert.isAbove(ravenData.exception.values[0].stacktrace.frames.length, 1);
                }
            );
        });

        it('should capture an Error object passed to Raven.captureException w/ maxMessageLength set (#647)', function (done) {
            var iframe = this.iframe;
            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    Raven._globalOptions.maxMessageLength = 100;
                    Raven.captureException(new Error('lol'), {
                      level: 'warning',
                      extra: {
                        foo: 'bar'
                      }
                    });
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    assert.equal(ravenData.exception.values[0].type, 'Error');
                    assert.equal(ravenData.exception.values[0].value, 'lol');
                    assert.equal(ravenData.exception.values[0].stacktrace.frames.length, 1);
                }
            );
        });
    });

    describe('window.onerror', function () {
        it('should catch syntax errors', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);
                    eval('foo{};');
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    assert.isTrue(/SyntaxError/.test(ravenData.exception.values[0].type)); // full message differs per-browser
                    assert.equal(ravenData.exception.values[0].stacktrace.frames.length, 1); // just one frame
                }
            );
        });

        it('should NOT catch an exception already caught via Raven.wrap', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);
                    Raven.wrap(function () {
                        foo();
                    })();
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData;
                    assert.equal(ravenData.length, 1); // one caught error
                }
            );
        });

        it('should catch an exception already caught [but rethrown] via Raven.captureException', function (done) {
            // unlike Raven.wrap which ALWAYS re-throws, we don't know if the user will
            // re-throw an exception passed to Raven.captureException, and so we cannot
            // automatically suppress the next error caught through window.onerror
            var iframe = this.iframe;
            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);
                    try {
                        foo();
                    } catch (e) {
                        Raven.captureException(e);
                        throw e; // intentionally re-throw
                    }
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData;
                    assert.equal(ravenData.length, 2);
                }
            );
        });
    });

    describe('wrapped built-ins', function () {
        it('should capture exceptions from event listeners', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    var div = document.createElement('div');
                    document.body.appendChild(div);
                    div.addEventListener('click', function () {
                        foo();
                    }, false);

                    var evt;
                    if (document.createEvent) {
                        evt = document.createEvent('MouseEvents');
                        evt.initEvent('click', true, false);
                        div.dispatchEvent(evt);
                    } else if(document.createEventObject) {
                        div.fireEvent('onclick');
                    }
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    assert.isAbove(ravenData.exception.values[0].stacktrace.frames.length, 2);
                }
            );
        });

        it('should transparently remove event listeners from wrapped functions', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
              function () {
                  setTimeout(done);

                  var div = document.createElement('div');
                  document.body.appendChild(div);
                  var fooFn = function () { foo(); };
                  div.addEventListener('click', fooFn, false);
                  div.removeEventListener('click', fooFn);

                  var evt;
                  if (document.createEvent) {
                      evt = document.createEvent('MouseEvents');
                      evt.initEvent('click', true, false);
                      div.dispatchEvent(evt);
                  } else if(document.createEventObject) {
                      div.fireEvent('onclick');
                  }
              },
              function () {
                  var ravenData = iframe.contentWindow.ravenData[0];
                  assert.equal(ravenData, null); // should never trigger error
              }
            );
        });

        it('should capture exceptions inside setTimeout', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(function () {
                        setTimeout(done);
                        foo();
                    }, 10);
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    assert.isAbove(ravenData.exception.values[0].stacktrace.frames.length, 2);
                }
            );
        });

        it('should capture exceptions inside setInterval', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    var exceptionInterval = setInterval(function () {
                        setTimeout(done);
                        clearInterval(exceptionInterval);
                        foo();
                    }, 10);
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    assert.isAbove(ravenData.exception.values[0].stacktrace.frames.length, 2);
                }
            );
        });

        it('should capture exceptions inside requestAnimationFrame', function (done) {
            var iframe = this.iframe;
            // needs to be visible or requestAnimationFrame won't ever fire
            iframe.style.display = 'block';

            iframeExecute(iframe, done,
                function () {
                    requestAnimationFrame(function () {
                        setTimeout(done);
                        foo();
                    });
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    assert.isAbove(ravenData.exception.values[0].stacktrace.frames.length, 2);
                }
            );
        });

        it('should capture exceptions from XMLHttpRequest event handlers (e.g. onreadystatechange)', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
              function () {
                  var xhr = new XMLHttpRequest();

                  // intentionally assign event handlers *after* XMLHttpRequest.prototype.open,
                  // since this is what jQuery does
                  // https://github.com/jquery/jquery/blob/master/src/ajax/xhr.js#L37

                  xhr.open('GET', 'example.json');
                  xhr.onreadystatechange = function () {
                      setTimeout(done);
                      // replace onreadystatechange with no-op so exception doesn't
                      // fire more than once as XHR changes loading state
                      xhr.onreadystatechange = function () {};
                      foo();
                  };
                  xhr.send();
              },
              function () {
                  var ravenData = iframe.contentWindow.ravenData[0];
                  // # of frames alter significantly between chrome/firefox & safari
                  assert.isAbove(ravenData.exception.values[0].stacktrace.frames.length, 2);
              }
            );
        });

        it('should capture exceptions from $.fn.ready (jQuery)', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    $(function () {
                        foo();
                    });
                },
                function () {
                    var ravenData = iframe.contentWindow.ravenData[0];
                    // # of frames alter significantly between chrome/firefox & safari
                    assert.isAbove(ravenData.exception.values[0].stacktrace.frames.length, 2);
                }
            );
        });
    });

    describe('breadcrumbs', function () {

        it('should record an XMLHttpRequest', function (done) {
           var iframe = this.iframe;

            iframeExecute(iframe, done,
              function () {
                  // some browsers trigger onpopstate for load / reset breadcrumb state
                  Raven._breadcrumbs = [];

                  var xhr = new XMLHttpRequest();

                  xhr.open('GET', '/test/integration/example.json');
                  xhr.setRequestHeader('Content-type', 'application/json');
                  xhr.onreadystatechange = function () {
                      // don't fire `done` handler until at least *one* onreadystatechange
                      // has occurred (doesn't actually need to finish)
                      if (xhr.readyState === 4) {
                          setTimeout(done);
                      }
                  };
                  xhr.send();
              },
              function () {
                  var Raven = iframe.contentWindow.Raven,
                      breadcrumbs = Raven._breadcrumbs;

                   assert.equal(breadcrumbs.length, 1);

                   assert.equal(breadcrumbs[0].type, 'http');
                   assert.equal(breadcrumbs[0].data.method, 'GET');
                   // NOTE: not checking status code because we seem to get
                   //       statusCode 0/undefined from Phantom when fetching
                   //       example.json (CORS issue?
              }
            );
        });

        it('should record an XMLHttpRequest without any handlers set', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    // I hate to do a time-based "done" trigger, but unfortunately we can't
                    // set an onload/onreadystatechange handler on XHR to verify that it finished
                    // - that's the whole point of this test! :(
                    setTimeout(done, 1000);

                    // some browsers trigger onpopstate for load / reset breadcrumb state
                    Raven._breadcrumbs = [];

                    var xhr = new XMLHttpRequest();

                    xhr.open('GET', '/test/integration/example.json');
                    xhr.setRequestHeader('Content-type', 'application/json');
                    xhr.send();
                },
                function () {
                    var Raven = iframe.contentWindow.Raven,
                        breadcrumbs = Raven._breadcrumbs;

                    assert.equal(breadcrumbs.length, 1);

                    assert.equal(breadcrumbs[0].type, 'http');
                    assert.equal(breadcrumbs[0].data.method, 'GET');
                    // NOTE: not checking status code because we seem to get
                    //       statusCode 0/undefined from Phantom when fetching
                    //       example.json (CORS issue?
                }
            );
        });

        it('should NOT capture breadcrumbs from XMLHttpRequests to the Sentry store endpoint', function (done) {
            var iframe = this.iframe;
            iframeExecute(iframe, done,
              function () {
                  // some browsers trigger onpopstate for load / reset breadcrumb state
                  Raven._breadcrumbs = [];

                  var xhr = new XMLHttpRequest();
                  xhr.open('GET', 'https://example.com/api/1/store/?sentry_key=public');
                  xhr.setRequestHeader('Content-type', 'application/json');
                  xhr.onreadystatechange = function () {
                      // don't fire `done` handler until at least *one* onreadystatechange
                      // has occurred (doesn't actually need to finish)
                      if (xhr.readyState === 4) {
                          setTimeout(done);
                      }
                  };
                  xhr.send();
              },
              function () {
                  var Raven = iframe.contentWindow.Raven,
                      breadcrumbs = Raven._breadcrumbs;

                   assert.equal(breadcrumbs.length, 0);
              }
            );
        });

        it('should record a mouse click on element WITH click handler present', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    // some browsers trigger onpopstate for load / reset breadcrumb state
                    Raven._breadcrumbs = [];

                    // add an event listener to the input. we want to make sure that
                    // our breadcrumbs still work even if the page has an event listener
                    // on an element that cancels event bubbling
                    var input = document.getElementsByTagName('input')[0];
                    var clickHandler = function (evt) {
                        evt.stopPropagation(); // don't bubble
                    };
                    input.addEventListener('click', clickHandler);

                    // click <input/>
                    var evt = document.createEvent('MouseEvent');
                    evt.initMouseEvent(
                        "click",
                        true /* bubble */,
                        true /* cancelable */,
                        window,
                        null,
                        0, 0, 0, 0, /* coordinates */
                        false, false, false, false, /* modifier keys */
                        0 /*left*/,
                        null
                    );
                    input.dispatchEvent(evt);
                },
                function () {
                    var Raven = iframe.contentWindow.Raven,
                        breadcrumbs = Raven._breadcrumbs;

                    assert.equal(breadcrumbs.length, 1);

                    assert.equal(breadcrumbs[0].category, 'ui.click');
                    assert.equal(breadcrumbs[0].message, 'body > form#foo-form > input[name="foo"]');
                }
            );
        });

        it('should record a mouse click on element WITHOUT click handler present', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    // some browsers trigger onpopstate for load / reset breadcrumb state
                    Raven._breadcrumbs = [];

                    // click <input/>
                    var evt = document.createEvent('MouseEvent');
                    evt.initMouseEvent(
                        "click",
                        true /* bubble */,
                        true /* cancelable */,
                        window,
                        null,
                        0, 0, 0, 0, /* coordinates */
                        false, false, false, false, /* modifier keys */
                        0 /*left*/,
                        null
                    );

                    var input = document.getElementsByTagName('input')[0];
                    input.dispatchEvent(evt);
                },
                function () {
                    var Raven = iframe.contentWindow.Raven,
                        breadcrumbs = Raven._breadcrumbs;

                    assert.equal(breadcrumbs.length, 1);

                    assert.equal(breadcrumbs[0].category, 'ui.click');
                    assert.equal(breadcrumbs[0].message, 'body > form#foo-form > input[name="foo"]');
                }
            );
        });

        it('should only record a SINGLE mouse click for a tree of elements with event listeners', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    // some browsers trigger onpopstate for load / reset breadcrumb state
                    Raven._breadcrumbs = [];

                    var clickHandler = function (evt) {
                        //evt.stopPropagation();
                    };
                    document.querySelector('.a').addEventListener('click', clickHandler);
                    document.querySelector('.b').addEventListener('click', clickHandler);
                    document.querySelector('.c').addEventListener('click', clickHandler);

                    // click <input/>
                    var evt = document.createEvent('MouseEvent');
                    evt.initMouseEvent(
                        "click",
                        true /* bubble */,
                        true /* cancelable */,
                        window,
                        null,
                        0, 0, 0, 0, /* coordinates */
                        false, false, false, false, /* modifier keys */
                        0 /*left*/,
                        null
                    );

                    var input = document.querySelector('.a'); // leaf node
                    input.dispatchEvent(evt);
                },
                function () {
                    var Raven = iframe.contentWindow.Raven,
                        breadcrumbs = Raven._breadcrumbs;

                    assert.equal(breadcrumbs.length, 1);

                    assert.equal(breadcrumbs[0].category, 'ui.click');
                    assert.equal(breadcrumbs[0].message, 'body > div.c > div.b > div.a');
                }
            );
        });

        it('should record consecutive keypress events into a single "input" breadcrumb', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    // some browsers trigger onpopstate for load / reset breadcrumb state
                    Raven._breadcrumbs = [];

                    // keypress <input/> twice
                    var keypress1 = document.createEvent('KeyboardEvent');
                    keypress1.initKeyboardEvent("keypress", true, true, window, "b", 66, 0, "", false);

                    var keypress2 = document.createEvent('KeyboardEvent');
                    keypress2.initKeyboardEvent("keypress", true, true, window, "a", 65, 0, "", false);

                    var input = document.getElementsByTagName('input')[0];
                    input.dispatchEvent(keypress1);
                    input.dispatchEvent(keypress2);
                },
                function () {
                    var Raven = iframe.contentWindow.Raven,
                        breadcrumbs = Raven._breadcrumbs;

                    assert.equal(breadcrumbs.length, 1);

                    assert.equal(breadcrumbs[0].category, 'ui.input');
                    assert.equal(breadcrumbs[0].message, 'body > form#foo-form > input[name="foo"]');
                }
            );
        });

        it('should flush keypress breadcrumbs when an error is thrown', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    // some browsers trigger onpopstate for load / reset breadcrumb state
                    Raven._breadcrumbs = [];

                    // keypress <input/>
                    var keypress = document.createEvent('KeyboardEvent');
                    keypress.initKeyboardEvent("keypress", true, true, window, "b", 66, 0, "", false);

                    var input = document.getElementsByTagName('input')[0];
                    input.dispatchEvent(keypress);

                    foo(); // throw exception
                },
                function () {
                    var Raven = iframe.contentWindow.Raven,
                        breadcrumbs = Raven._breadcrumbs;

                    // 2 breadcrumbs: `ui_event`, then `error`
                    assert.equal(breadcrumbs.length, 2);

                    assert.equal(breadcrumbs[0].category, 'ui.input');
                    assert.equal(breadcrumbs[0].message, 'body > form#foo-form > input[name="foo"]');
                }
            );
        });

        it('should flush keypress breadcrumb when input event occurs immediately after', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    // some browsers trigger onpopstate for load / reset breadcrumb state
                    Raven._breadcrumbs = [];

                    // 1st keypress <input/>
                    var keypress1 = document.createEvent('KeyboardEvent');
                    keypress1.initKeyboardEvent("keypress", true, true, window, "b", 66, 0, "", false);

                    // click <input/>
                    var click= document.createEvent('MouseEvent');
                    click.initMouseEvent(
                        "click",
                        true /* bubble */,
                        true /* cancelable */,
                        window,
                        null,
                        0, 0, 0, 0, /* coordinates */
                        false, false, false, false, /* modifier keys */
                        0 /*left*/,
                        null
                    );

                    // 2nd keypress
                    var keypress2 = document.createEvent('KeyboardEvent');
                    keypress2.initKeyboardEvent("keypress", true, true, window, "a", 65, 0, "", false);

                    var input = document.getElementsByTagName('input')[0];
                    input.dispatchEvent(keypress1);
                    input.dispatchEvent(click);
                    input.dispatchEvent(keypress2);
                },
                function () {
                    var Raven = iframe.contentWindow.Raven,
                        breadcrumbs = Raven._breadcrumbs;

                    // 2x `ui_event`
                    assert.equal(breadcrumbs.length, 3);

                    assert.equal(breadcrumbs[0].category, 'ui.input');
                    assert.equal(breadcrumbs[0].message, 'body > form#foo-form > input[name="foo"]');

                    assert.equal(breadcrumbs[1].category, 'ui.click');
                    assert.equal(breadcrumbs[1].message, 'body > form#foo-form > input[name="foo"]');

                    assert.equal(breadcrumbs[2].category, 'ui.input');
                    assert.equal(breadcrumbs[2].message, 'body > form#foo-form > input[name="foo"]');

                }
            );
        });

        it('should record history.[pushState|back] changes as navigation breadcrumbs', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
              function () {
                  // some browsers trigger onpopstate for load / reset breadcrumb state
                  Raven._breadcrumbs = [];

                  history.pushState({}, '', '/foo');
                  history.pushState({}, '', '/bar?a=1#fragment');
                  history.pushState({}, '', {}); // pushState calls toString on non-string args
                  history.pushState({}, '', null); // does nothing / no-op

                  // can't call history.back() because it will change url of parent document
                  // (e.g. document running mocha) ... instead just "emulate" a back button
                  // press by calling replaceState + onpopstate manually
                  history.replaceState({}, '', '/bar?a=1#fragment');
                  window.onpopstate();
                  done();
              },
              function () {
                  var Raven = iframe.contentWindow.Raven,
                      breadcrumbs = Raven._breadcrumbs,
                      from,
                      to;

                  assert.equal(breadcrumbs.length, 4);
                  assert.equal(breadcrumbs[0].category, 'navigation'); // (start) => foo
                  assert.equal(breadcrumbs[1].category, 'navigation'); // foo => bar?a=1#fragment
                  assert.equal(breadcrumbs[2].category, 'navigation'); // bar?a=1#fragment => [object%20Object]
                  assert.equal(breadcrumbs[3].category, 'navigation'); // [object%20Object] => bar?a=1#fragment (back button)

                  // assert end of string because PhantomJS uses full system path
                  assert.ok(/\/test\/integration\/frame\.html$/.test(Raven._breadcrumbs[0].data.from), '\'from\' url is incorrect');
                  assert.ok(/\/foo$/.test(breadcrumbs[0].data.to), '\'to\' url is incorrect');

                  assert.ok(/\/foo$/.test(breadcrumbs[1].data.from), '\'from\' url is incorrect');
                  assert.ok(/\/bar\?a=1#fragment$/.test(breadcrumbs[1].data.to), '\'to\' url is incorrect');

                  assert.ok(/\/bar\?a=1#fragment$/.test(breadcrumbs[2].data.from), '\'from\' url is incorrect');
                  assert.ok(/\[object Object\]$/.test(breadcrumbs[2].data.to), '\'to\' url is incorrect');

                  assert.ok(/\[object Object\]$/.test(breadcrumbs[3].data.from), '\'from\' url is incorrect');
                  assert.ok(/\/bar\?a=1#fragment/.test(breadcrumbs[3].data.to), '\'to\' url is incorrect');
              }
            );
        });
    });

    describe('uninstall', function () {
        it('should restore original built-ins', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);
                    Raven.uninstall();

                    window.isRestored = {
                        setTimeout: originalBuiltIns.setTimeout === setTimeout,
                        setInterval: originalBuiltIns.setInterval === setInterval,
                        requestAnimationFrame: originalBuiltIns.requestAnimationFrame === requestAnimationFrame,
                        xhrProtoOpen: originalBuiltIns.xhrProtoOpen === XMLHttpRequest.prototype.open,
                        headAddEventListener: originalBuiltIns.headAddEventListener === document.body.addEventListener,
                        headRemoveEventListener: originalBuiltIns.headRemoveEventListener === document.body.removeEventListener
                    };
                },
                function () {
                    var isRestored = iframe.contentWindow.isRestored;
                    assert.isTrue(isRestored.setTimeout);
                    assert.isTrue(isRestored.setInterval);
                    assert.isTrue(isRestored.requestAnimationFrame);
                    assert.isTrue(isRestored.xhrProtoOpen);
                    assert.isTrue(isRestored.headAddEventListener);
                    assert.isTrue(isRestored.headRemoveEventListener);
                }
            );
        });

        it('should not restore XMLHttpRequest instance methods', function (done) {
            var iframe = this.iframe;

            iframeExecute(iframe, done,
                function () {
                    setTimeout(done);

                    var xhr = new XMLHttpRequest();
                    var origOnReadyStateChange = xhr.onreadystatechange = function () {};
                    xhr.open('GET', '/foo/');
                    xhr.abort();

                    Raven.uninstall();

                    window.isOnReadyStateChangeRestored = xhr.onready === origOnReadyStateChange;
                },
                function () {
                    assert.isFalse(iframe.contentWindow.isOnReadyStateChangeRestored);
                }
            );
        });
    });
});
