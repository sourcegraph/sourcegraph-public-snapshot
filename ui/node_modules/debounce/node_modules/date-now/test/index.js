var test = require("tape")
var setTimeout = require("timers").setTimeout

var now = require("../index")
var seeded = require("../seed")

test("date", function (assert) {
    var before = new Date().getTime()
    var ts = now()
    var after = new Date().getTime()
    assert.ok(before <= ts)
    assert.ok(after >= ts)
    assert.end()
})

test("seeded", function (assert) {
    var before = now()
    var time = seeded(40)
    var after = now()

    var bts = now()
    var ts = time()
    var ats = now()

    assert.ok(ts >= bts - before + 40)
    assert.ok(ts <= ats - after + 40)

    setTimeout(function () {
        var bts = now()
        var ts = time()
        var ats = now()

        assert.ok(ts >= bts - before + 40)
        assert.ok(ts <= ats - after + 40)

        assert.end()
    }, 50)
})
