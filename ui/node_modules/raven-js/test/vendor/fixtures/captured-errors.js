var CapturedExceptions = {};

CapturedExceptions.OPERA_854 = {
    message: "Statement on line 44: Type mismatch (usually a non-object value used where an object is required)\n" +
    "Backtrace:\n" +
    "  Line 44 of linked script http://path/to/file.js\n" +
    "    this.undef();\n" +
    "  Line 31 of linked script http://path/to/file.js\n" +
    "    ex = ex || this.createException();\n" +
    "  Line 18 of linked script http://path/to/file.js\n" +
    "    var p = new printStackTrace.implementation(), result = p.run(ex);\n" +
    "  Line 4 of inline#1 script in http://path/to/file.js\n" +
    "    printTrace(printStackTrace());\n" +
    "  Line 7 of inline#1 script in http://path/to/file.js\n" +
    "    bar(n - 1);\n" +
    "  Line 11 of inline#1 script in http://path/to/file.js\n" +
    "    bar(2);\n" +
    "  Line 15 of inline#1 script in http://path/to/file.js\n" +
    "    foo();\n" +
    "",
    'opera#sourceloc': 44
};

CapturedExceptions.OPERA_902 = {
    message: "Statement on line 44: Type mismatch (usually a non-object value used where an object is required)\n" +
    "Backtrace:\n" +
    "  Line 44 of linked script http://path/to/file.js\n" +
    "    this.undef();\n" +
    "  Line 31 of linked script http://path/to/file.js\n" +
    "    ex = ex || this.createException();\n" +
    "  Line 18 of linked script http://path/to/file.js\n" +
    "    var p = new printStackTrace.implementation(), result = p.run(ex);\n" +
    "  Line 4 of inline#1 script in http://path/to/file.js\n" +
    "    printTrace(printStackTrace());\n" +
    "  Line 7 of inline#1 script in http://path/to/file.js\n" +
    "    bar(n - 1);\n" +
    "  Line 11 of inline#1 script in http://path/to/file.js\n" +
    "    bar(2);\n" +
    "  Line 15 of inline#1 script in http://path/to/file.js\n" +
    "    foo();\n" +
    "",
    'opera#sourceloc': 44
};

CapturedExceptions.OPERA_927 = {
    message: "Statement on line 43: Type mismatch (usually a non-object value used where an object is required)\n" +
    "Backtrace:\n" +
    "  Line 43 of linked script http://path/to/file.js\n" +
    "    bar(n - 1);\n" +
    "  Line 31 of linked script http://path/to/file.js\n" +
    "    bar(2);\n" +
    "  Line 18 of linked script http://path/to/file.js\n" +
    "    foo();\n" +
    "",
    'opera#sourceloc': 43
};

CapturedExceptions.OPERA_964 = {
    message: "Statement on line 42: Type mismatch (usually non-object value supplied where object required)\n" +
    "Backtrace:\n" +
    "  Line 42 of linked script http://path/to/file.js\n" +
    "                this.undef();\n" +
    "  Line 27 of linked script http://path/to/file.js\n" +
    "            ex = ex || this.createException();\n" +
    "  Line 18 of linked script http://path/to/file.js: In function printStackTrace\n" +
    "        var p = new printStackTrace.implementation(), result = p.run(ex);\n" +
    "  Line 4 of inline#1 script in http://path/to/file.js: In function bar\n" +
    "             printTrace(printStackTrace());\n" +
    "  Line 7 of inline#1 script in http://path/to/file.js: In function bar\n" +
    "           bar(n - 1);\n" +
    "  Line 11 of inline#1 script in http://path/to/file.js: In function foo\n" +
    "           bar(2);\n" +
    "  Line 15 of inline#1 script in http://path/to/file.js\n" +
    "         foo();\n" +
    "",
    'opera#sourceloc': 42,
    stacktrace: "  ...  Line 27 of linked script http://path/to/file.js\n" +
    "            ex = ex || this.createException();\n" +
    "  Line 18 of linked script http://path/to/file.js: In function printStackTrace\n" +
    "        var p = new printStackTrace.implementation(), result = p.run(ex);\n" +
    "  Line 4 of inline#1 script in http://path/to/file.js: In function bar\n" +
    "             printTrace(printStackTrace());\n" +
    "  Line 7 of inline#1 script in http://path/to/file.js: In function bar\n" +
    "           bar(n - 1);\n" +
    "  Line 11 of inline#1 script in http://path/to/file.js: In function foo\n" +
    "           bar(2);\n" +
    "  Line 15 of inline#1 script in http://path/to/file.js\n" +
    "         foo();\n" +
    ""
};

CapturedExceptions.OPERA_10 = {
    message: "Statement on line 42: Type mismatch (usually non-object value supplied where object required)",
    'opera#sourceloc': 42,
    stacktrace: "  Line 42 of linked script http://path/to/file.js\n" +
    "                this.undef();\n" +
    "  Line 27 of linked script http://path/to/file.js\n" +
    "            ex = ex || this.createException();\n" +
    "  Line 18 of linked script http://path/to/file.js: In function printStackTrace\n" +
    "        var p = new printStackTrace.implementation(), result = p.run(ex);\n" +
    "  Line 4 of inline#1 script in http://path/to/file.js: In function bar\n" +
    "             printTrace(printStackTrace());\n" +
    "  Line 7 of inline#1 script in http://path/to/file.js: In function bar\n" +
    "           bar(n - 1);\n" +
    "  Line 11 of inline#1 script in http://path/to/file.js: In function foo\n" +
    "           bar(2);\n" +
    "  Line 15 of inline#1 script in http://path/to/file.js\n" +
    "         foo();\n" +
    ""
};

CapturedExceptions.OPERA_11 = {
    message: "'this.undef' is not a function",
    stack: "<anonymous function: run>([arguments not available])@http://path/to/file.js:27\n" +
    "bar([arguments not available])@http://domain.com:1234/path/to/file.js:18\n" +
    "foo([arguments not available])@http://domain.com:1234/path/to/file.js:11\n" +
    "<anonymous function>@http://path/to/file.js:15\n" +
    "Error created at <anonymous function>@http://path/to/file.js:15",
    stacktrace: "Error thrown at line 42, column 12 in <anonymous function: createException>() in http://path/to/file.js:\n" +
    "    this.undef();\n" +
    "called from line 27, column 8 in <anonymous function: run>(ex) in http://path/to/file.js:\n" +
    "    ex = ex || this.createException();\n" +
    "called from line 18, column 4 in printStackTrace(options) in http://path/to/file.js:\n" +
    "    var p = new printStackTrace.implementation(), result = p.run(ex);\n" +
    "called from line 4, column 5 in bar(n) in http://path/to/file.js:\n" +
    "    printTrace(printStackTrace());\n" +
    "called from line 7, column 4 in bar(n) in http://path/to/file.js:\n" +
    "    bar(n - 1);\n" +
    "called from line 11, column 4 in foo() in http://path/to/file.js:\n" +
    "    bar(2);\n" +
    "called from line 15, column 3 in http://path/to/file.js:\n" +
    "    foo();"
};

CapturedExceptions.OPERA_12 = {
    message: "Cannot convert 'x' to object",
    stack: "<anonymous function>([arguments not available])@http://localhost:8000/ExceptionLab.html:48\n" +
    "dumpException3([arguments not available])@http://localhost:8000/ExceptionLab.html:46\n" +
    "<anonymous function>([arguments not available])@http://localhost:8000/ExceptionLab.html:1",
    stacktrace: "Error thrown at line 48, column 12 in <anonymous function>(x) in http://localhost:8000/ExceptionLab.html:\n" +
    "    x.undef();\n" +
    "called from line 46, column 8 in dumpException3() in http://localhost:8000/ExceptionLab.html:\n" +
    "    dumpException((function(x) {\n" +
    "called from line 1, column 0 in <anonymous function>(event) in http://localhost:8000/ExceptionLab.html:\n" +
    "    dumpException3();"
};

CapturedExceptions.OPERA_25 = {
    message: "Cannot read property 'undef' of null",
    name: "TypeError",
    stack: "TypeError: Cannot read property 'undef' of null\n" +
    "    at http://path/to/file.js:47:22\n" +
    "    at foo (http://path/to/file.js:52:15)\n" +
    "    at bar (http://path/to/file.js:108:168)"
};

CapturedExceptions.CHROME_15 = {
    'arguments': ["undef"],
    message: "Object #<Object> has no method 'undef'",
    stack: "TypeError: Object #<Object> has no method 'undef'\n" +
    "    at bar (http://path/to/file.js:13:17)\n" +
    "    at bar (http://path/to/file.js:16:5)\n" +
    "    at foo (http://path/to/file.js:20:5)\n" +
    "    at http://path/to/file.js:24:4"
};

CapturedExceptions.CHROME_36 = {
    message: "Default error",
    name: "Error",
    stack: "Error: Default error\n" +
    "    at dumpExceptionError (http://localhost:8080/file.js:41:27)\n" +
    "    at HTMLButtonElement.onclick (http://localhost:8080/file.js:107:146)\n" +
    "    at I.e.fn.(anonymous function) [as index] (http://localhost:8080/file.js:10:3651)"
};

CapturedExceptions.FIREFOX_3 = {
    fileName: "http://127.0.0.1:8000/js/stacktrace.js",
    lineNumber: 44,
    message: "this.undef is not a function",
    name: "TypeError",
    stack: "()@http://127.0.0.1:8000/js/stacktrace.js:44\n" +
    "(null)@http://127.0.0.1:8000/js/stacktrace.js:31\n" +
    "printStackTrace()@http://127.0.0.1:8000/js/stacktrace.js:18\n" +
    "bar(1)@http://127.0.0.1:8000/js/file.js:13\n" +
    "bar(2)@http://127.0.0.1:8000/js/file.js:16\n" +
    "foo()@http://127.0.0.1:8000/js/file.js:20\n" +
    "@http://127.0.0.1:8000/js/file.js:24\n" +
    ""
};

CapturedExceptions.FIREFOX_7 = {
    fileName: "file:///G:/js/stacktrace.js",
    lineNumber: 44,
    stack: "()@file:///G:/js/stacktrace.js:44\n" +
    "(null)@file:///G:/js/stacktrace.js:31\n" +
    "printStackTrace()@file:///G:/js/stacktrace.js:18\n" +
    "bar(1)@file:///G:/js/file.js:13\n" +
    "bar(2)@file:///G:/js/file.js:16\n" +
    "foo()@file:///G:/js/file.js:20\n" +
    "@file:///G:/js/file.js:24\n" +
    ""
};

CapturedExceptions.FIREFOX_14 = {
    message: "x is null",
    stack: "@http://path/to/file.js:48\n" +
    "dumpException3@http://path/to/file.js:52\n" +
    "onclick@http://path/to/file.js:1\n" +
    "",
    fileName: "http://path/to/file.js",
    lineNumber: 48
};

CapturedExceptions.FIREFOX_31 = {
    message: "Default error",
    name: "Error",
    stack: "foo@http://path/to/file.js:41:13\n" +
    "bar@http://path/to/file.js:1:1\n" +
    ".plugin/e.fn[c]/<@http://path/to/file.js:1:1\n" +
    "",
    fileName: "http://path/to/file.js",
    lineNumber: 41,
    columnNumber: 12
};

// Internal errors sometimes thrown by Firefox
// More here: https://developer.mozilla.org/en-US/docs/Mozilla/Errors
//
// Note that such errors are instanceof "Exception", not "Error"
CapturedExceptions.FIREFOX_44_NS_EXCEPTION = {
    message: "",
    name: "NS_ERROR_FAILURE",
    stack: "[2]</Bar.prototype._baz/</<@http://path/to/file.js:703:28\n" +
    "App.prototype.foo@file:///path/to/file.js:15:2\n" +
    "bar@file:///path/to/file.js:20:3\n" +
    "@file:///path/to/index.html:23:1\n" + // inside <script> tag
    "",
    fileName: "http://path/to/file.js",
    columnNumber: 0,
    lineNumber: 703,
    result: 2147500037
};

CapturedExceptions.SAFARI_6 = {
    message: "'null' is not an object (evaluating 'x.undef')",
    stack: "@http://path/to/file.js:48\n" +
    "dumpException3@http://path/to/file.js:52\n" +
    "onclick@http://path/to/file.js:82\n" +
    "[native code]",
    line: 48,
    sourceURL: "http://path/to/file.js"
};

CapturedExceptions.SAFARI_7 = {
    message: "'null' is not an object (evaluating 'x.undef')",
    name: "TypeError",
    stack: "http://path/to/file.js:48:22\n" +
    "foo@http://path/to/file.js:52:15\n" +
    "bar@http://path/to/file.js:108:107",
    line: 47,
    sourceURL: "http://path/to/file.js"
};

CapturedExceptions.SAFARI_8 = {
    message: "null is not an object (evaluating 'x.undef')",
    name: "TypeError",
    stack: "http://path/to/file.js:47:22\n" +
    "foo@http://path/to/file.js:52:15\n" +
    "bar@http://path/to/file.js:108:23",
    line: 47,
    column: 22,
    sourceURL: "http://path/to/file.js"
};

CapturedExceptions.SAFARI_8_EVAL = {
    message: "Can't find variable: getExceptionProps",
    name: "ReferenceError",
    stack: "eval code\n" +
    "eval@[native code]\n" +
    "foo@http://path/to/file.js:58:21\n" +
    "bar@http://path/to/file.js:109:91",
    line: 1,
    column: 18
};

CapturedExceptions.IE_9 = {
    message: "Unable to get property 'undef' of undefined or null reference",
    description: "Unable to get property 'undef' of undefined or null reference"
};

CapturedExceptions.IE_10 = {
    message: "Unable to get property 'undef' of undefined or null reference",
    stack: "TypeError: Unable to get property 'undef' of undefined or null reference\n" +
    "   at Anonymous function (http://path/to/file.js:48:13)\n" +
    "   at foo (http://path/to/file.js:46:9)\n" +
    "   at bar (http://path/to/file.js:82:1)",
    description: "Unable to get property 'undef' of undefined or null reference",
    number: -2146823281
};

CapturedExceptions.IE_11 = {
    message: "Unable to get property 'undef' of undefined or null reference",
    name: "TypeError",
    stack: "TypeError: Unable to get property 'undef' of undefined or null reference\n" +
    "   at Anonymous function (http://path/to/file.js:47:21)\n" +
    "   at foo (http://path/to/file.js:45:13)\n" +
    "   at bar (http://path/to/file.js:108:1)",
    description: "Unable to get property 'undef' of undefined or null reference",
    number: -2146823281
};

CapturedExceptions.IE_11_EVAL = {
    message: "'getExceptionProps' is undefined",
    name: "ReferenceError",
    stack: "ReferenceError: 'getExceptionProps' is undefined\n" +
    "   at eval code (eval code:1:1)\n" +
    "   at foo (http://path/to/file.js:58:17)\n" +
    "   at bar (http://path/to/file.js:109:1)",
    description: "'getExceptionProps' is undefined",
    number: -2146823279
};

CapturedExceptions.CHROME_48_BLOB = {
    message: "Error: test",
    name: "Error",
    stack: "Error: test\n" +
    "    at Error (native)\n" +
    "    at s (blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379:31:29146)\n" +
    "    at Object.d [as add] (blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379:31:30039)\n" +
    "    at blob:http%3A//localhost%3A8080/d4eefe0f-361a-4682-b217-76587d9f712a:15:10978\n" +
    "    at blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379:1:6911\n" +
    "    at n.fire (blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379:7:3019)\n" +
    "    at n.handle (blob:http%3A//localhost%3A8080/abfc40e9-4742-44ed-9dcd-af8f99a29379:7:2863)"
};

CapturedExceptions.PHANTOMJS_1_19 = {
    stack: "Error: foo\n" +
    "    at file:///path/to/file.js:878\n" +
    "    at foo (http://path/to/file.js:4283)\n" +
    "    at http://path/to/file.js:4287"
};

module.exports = CapturedExceptions;
