(function() {
    var foo = 'bar';
    window.alert(foo);

    foo = foo === 'baz'
        ? 'hello'
        : 'hi';

    foo = foo === 'baz'
                ? 'hello'
                : 'hi';

    switch(foo) {
        case 'hello':
                window.alert('hello');
            break;
        case 'hi':
            window.alert('hi');
            break;
    }

})();
