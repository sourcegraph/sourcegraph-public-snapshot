const mocha = require("mocha");
const { Console } = require("console");
const fs = require("fs");
const {
    EVENT_RUN_BEGIN,
    EVENT_TEST_FAIL,
    EVENT_TEST_PASS,
    EVENT_TEST_PENDING,
    EVENT_RUN_END,
} = mocha.Runner.constants;
const Spec = mocha.reporters.Spec;
const Base = mocha.reporters.Base;

class SpecFileReporter extends mocha.reporters.Spec {
    constructor(runner, options) {
        super(runner, options);
        this.console = new Console({
            stdout: fs.createWriteStream("s-test.stdout.txt"),
            stderr: fs.createWriteStream("s-test.stderr.txt"),
        });
        Base.call(this, runner, options);
        //var self = this

        this._runner = runner;
        this._spec = new Spec(runner, options);

        var passes = [];
        var failures = [];
        var pending = [];

        runner.on(EVENT_TEST_PASS, function(test) {
            passes.push(test);
        });

        runner.on(EVENT_TEST_FAIL, function(test) {
            failures.push(test);
        });

        runner.on(EVENT_TEST_PENDING, function(test) {
            pending.push(test);
        });

        runner.once(EVENT_RUN_END, function() {
            let tmp = Base.consoleLog;
            Base.consoleLog = c.log;
            Base.list(self.failures);
            Base.consoleLog = tmp;
        });


        return this;
    }
}
SpecFileReporter.prototype.__proto__ = Base.prototype;

module.exports = SpecFileReporter;
