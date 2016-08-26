function foo() {
    console.log("lol, i don't do anything")
}

function foo2() {
    foo()
    console.log('i called foo')
}

function broken() {
    try {
        /*fkjdsahfdhskfhdsahfudshafuoidashfudsa*/ fdasfds[0]; // i throw an error h sadhf hadsfdsakf kl;dsjaklf jdklsajfk ljds;klafldsl fkhdas;hf hdsaf hdsalfhjldksahfljkdsahfjkl dhsajkfl hdklsahflkjdsahkfj hdsjakhf dkashfl diusafh kdsjahfkldsahf jkdashfj khdasjkfhdjksahflkjdhsakfhjdksahfjkdhsakf hdajskhf kjdash kjfads fjkadsh jkfdsa jkfdas jkfdjkas hfjkdsajlk fdsajk fjkdsa fjdsa fdkjlsa fjkdaslk hfjlkdsah fhdsahfui
    }catch(e) {
        Raven.captureException(e);
    }
}

function ready() {
    document.getElementById('test').onclick = broken;
}

function foo3() {
    document.getElementById('crap').value = 'barfdasjkfhoadshflkaosfjadiosfhdaskjfasfadsfads';
}

function somethingelse() {
    document.getElementById('somethingelse').value = 'this is some realy really long message just so our minification is largeeeeeeeeee!';
}

function derp() {
    fdas[0];
}

function testOptions() {
    Raven.context({tags: {foo: 'bar'}}, function() {
        throw new Error('foo');
    });
}

function throwString() {
    throw 'oops';
}

function showDialog() {
    broken();
    Raven.showReportDialog();
}

function blobExample() {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', 'stack.js');
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            var blob = new Blob([xhr.responseText], {type: 'application/javascript'});
            var url = URL.createObjectURL(blob);

            var script = document.createElement('script');
            script.src = url;
            document.head.appendChild(script);
        }
    };
    xhr.send();
}

function a() { b(); }
function b() { c(); }
function c() { d(); }
function d() { e(); }
function e() { f(); }
function f() { g(); }
function g() { h(); }
function h() { i(); }
function i() { j(); }
function j() { k(); }
function k() { l(); }
function l() { m(); }
function m() { n(); }
function n() { o(); }
function o() { throw new Error('dang'); }
