process.nextTick(function() {
	delete require.cache[module.id];
	if(typeof window !== "undefined" && window.mochaPhantomJS)
		mochaPhantomJS.run();
	else
		mocha.run();
});
