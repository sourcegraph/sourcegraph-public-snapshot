(function () {
    function bar() {
        baz();
    }

    function foo() {
        bar();
    }

    foo();
})();
