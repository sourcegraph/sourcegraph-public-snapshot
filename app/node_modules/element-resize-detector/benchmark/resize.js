(function() {
    "use strict";

    var numElements = 100;

    var count;

    var onAllElementsResized;
    var lastSize;
    var shrink;

    function onResize() {
        count++;

        if(count === numElements) {
            count = 0;
            onAllElementsResized();
        }
    }

    function resize(selector) {
        var newWidth;
        if(shrink) {
            shrink = false;
            newWidth = lastSize / 2;
        } else {
            shrink = true;
            newWidth = lastSize * 2;
        }

        $(selector).width(newWidth + "px");
        lastSize = newWidth;
    }

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

    var suite = new Benchmark.Suite("resize", {
        defer: true,
    });

    var originalRun = suite.run;

    suite.run = function() {
        console.log("Setting up suite...");
        var self = this;
        setup(function ready() {
            console.log("Setup done");
            getComputedStyle(document.body);
            setTimeout(function() {
                originalRun.call(self);
            }, 2000);
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
        loopCreateAndAppend(numElements, createDiv.bind(null, "100%"), $("#fixtures #scroll")[0]);
        loopCreateAndAppend(numElements, createDiv.bind(null, "100%"), $("#fixtures #object")[0]);

        var scrollready = false;
        var objectready = false;

        erdScroll.listenTo({
            onReady: function onReady() {
                console.log("scroll ready");
                scrollready = true;

                if(objectready) {
                    callback();
                }
            }
        },$("#fixtures #scroll .item"), onResize);
        erdObject.listenTo({
            onReady: function onReady() {
                console.log("object ready");

                objectready = true;

                if(scrollready) {
                    callback();
                }
            }
        }, $("#fixtures #object .item"), onResize);
    }

    suite.add("scroll strategy", function(deferred) {
        onAllElementsResized = function() {
            deferred.resolve();
            console.log("Test finished in " + (Date.now() - start) + " ms");
        }

        var start = Date.now();
        resize("#scroll");
    }, {
        defer: true,
        // maxTime: 20,
        onStart: function() {
            lastSize = $("#scroll").width();
            count = 0;
            shrink = true;
        }
    });

    suite.add("object strategy", function(deferred) {
        onAllElementsResized = function() {
            deferred.resolve();
            console.log("Test finished in " + (Date.now() - start) + " ms");
        }

        var start = Date.now();
        resize("#object");
    }, {
        defer: true,
        // maxTime: 20,
        onStart: function() {
            console.time("test");
            lastSize = $("#object").width();
            count = 0;
            shrink = true;
        },
        onComplete: function() {
            console.timeEnd("test");
        }
    });

    registerSuite(suite);
})();

// var objectsready = false;
// suite("resize", function() {
//     benchmark("object strategy initiater dummy", function (deferred) {
//         if(objectsready) {
//             deferred.resolve();
//             return;
//         }

//         setTimeout(function() {
//             if(objectsready) {
//                 deferred.resolve();
//             }
//         }, 0);
//     }, {
//         defer: true,
//         onStart: function() {
//             var erd = elementResizeDetectorMaker({
//                 callOnAdd: true,
//                 strategy: "object"
//             });
            
//             loopCreateAndAppend(numElements, createDiv.bind(null, numElements), document.getElementById("fixtures"));
//             lastSize = $("#fixtures").width();
//             shrink = true;
//             count = 0;

//             var calledcount = 0;

//             erd.listenTo($(".item"), function() {
//                 calledcount++;

//                 if(calledcount === numElements) {
//                     objectsready = true;
//                 }

//                 if(calledcount > numElements) {
//                     onResize();
//                 }
//             });
//         },
//     });

    // benchmark("object strategy", function(deferred) {
    //     onAllElementsResized = function() {
    //         deferred.resolve();
    //         //console.log("Test finished in " + (Date.now() - start) + " ms");
    //     }

    //     var start = Date.now();
    //     resize();
    // }, {
    //     defer: true,
    //     onComplete: function() {
    //         $("#fixtures").html("");
    //     }
    // });

//     benchmark("scroll strategy", function(deferred) {
//         onAllElementsResized = function() {
//             deferred.resolve();
//             //console.log("Test finished in " + (Date.now() - start) + " ms");
//         }

//         var start = Date.now();
//         resize();
//     }, {
//         defer: true,
//         onStart: function() {
//             var erd = elementResizeDetectorMaker({
//                 callOnAdd: false,
//                 strategy: "scroll"
//             });
            
//             loopCreateAndAppend(numElements, createDiv.bind(null, numElements), document.getElementById("fixtures"));
//             lastSize = $("#fixtures").width();
//             shrink = true;
//             count = 0;

//             erd.listenTo($(".item"), onResize);
//         },
//         onComplete: function() {
//             $("#fixtures").html("");
//         }
//     });
// }, {
//     onComplete: function(event) {
//         var bench = event.target;
//         var name = bench.name;
//         var hz = bench.hz;
//         var deviation = bench.stats.deviation;
//         var mean = bench.stats.mean;

//         console.log("bench:     " + name);
//         console.log("hz:        " + hz);
//         console.log("deviation: " + deviation);
//         console.log("mean:      " + mean);
//     }
// });