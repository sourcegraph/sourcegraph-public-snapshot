(function() {
    "use strict";

    var numElements = 100;

    function createDiv(width) {
        var d = document.createElement("div");
        d.className = "item";
        d.style.width = width;
        // erd.listenTo({
        //     callOnAdd: false
        // }, d, onElementResize);
        return d;
    }

    function loopCreateAndAppend(numNested, create, target) {
        for(var i = 0; i < numNested; i++) {
            var d = create();
            target.appendChild(d);
        }

        return target;
    }

    var suite = new Benchmark.Suite("install", {
        defer: true,
    });

    var originalRun = suite.run;

    suite.run = function() {
        console.log("Setting up suite...");
        var self = this;
        setup(function ready() {
            console.log("Setup done");
            getComputedStyle(document.body);
            originalRun.call(self);
        });
    }

    var erdScroll = elementResizeDetectorMaker({
        callOnAdd: false,
        strategy: "scroll"
    });
    
    var erdObject = elementResizeDetectorMaker({
        callOnAdd: false,
        strategy: "object"
    });

    function setup(callback) {
        $("#fixtures").html("<div id=scroll></div><div id=object></div>");
        callback();
    }

    suite.add("scroll strategy", function(deferred) {
        $("#fixtures #scroll").html("");
        loopCreateAndAppend(numElements, createDiv.bind(null, "100%"), $("#fixtures #scroll")[0]);

        var start = Date.now();
        erdScroll.listenTo({
            onReady: function() {
                deferred.resolve();
                var diff = Date.now() - start;
                if(diff === 0) {
                    throw new Error("lol");
                }
                console.log("Test finished in " + (diff) + " ms");
            }
        }, $("#scroll .item"), function noop() {
            //noop.
        });
    }, {
        defer: true,
        // maxTime: 20,
    });

    suite.add("object strategy", function(deferred) {
        $("#fixtures #object").html("");
        loopCreateAndAppend(numElements, createDiv.bind(null, "100%"), $("#fixtures #object")[0]);

        var start = Date.now();

        erdObject.listenTo({
            onReady: function() {
                deferred.resolve();
                console.log("Test finished in " + (Date.now() - start) + " ms");
            }
        }, $("#object .item"), function noop() {
            //noop.
        });
    }, {
        defer: true,
        // maxTime: 20
    });

    registerSuite(suite);
})();