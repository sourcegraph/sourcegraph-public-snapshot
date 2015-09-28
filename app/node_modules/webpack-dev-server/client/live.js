var $ = require("jquery");
var io = require("socket.io-client");
var stripAnsi = require('strip-ansi');
require("./style.css");

$(function() {
	var body = $("body").html(require("./page.jade")());
	var status = $("#status");
	var okness = $("#okness");
	var $errors = $("#errors");
	var iframe = $("#iframe");
	var header = $(".header");
	var hot = false;
	var currentHash = "";

	var contentPage = window.location.pathname.substr("/webpack-dev-server".length) + window.location.search;

	status.text("Connecting to socket.io server...");
	$errors.hide(); iframe.hide();
	header.css({borderColor: "#96b5b4"});
	io = io.connect();

	io.on("hot", function() {
		hot = true;
		iframe.attr("src", contentPage + window.location.hash);
	});

	io.on("invalid", function() {
		okness.text("");
		status.text("App updated. Recompiling...");
		header.css({borderColor: "#96b5b4"});
		$errors.hide(); if(!hot) iframe.hide();
	});

	io.on("hash", function(hash) {
		currentHash = hash;
	});

	io.on("still-ok", function() {
		okness.text("");
		status.text("App ready.");
		header.css({borderColor: ""});
		$errors.hide(); if(!hot) iframe.show();
	});

	io.on("ok", function() {
		okness.text("");
		$errors.hide();
		reloadApp();
	});

	io.on("warnings", function(warnings) {
		okness.text("Warnings while compiling.");
		$errors.hide();
		reloadApp();
	});

	io.on("errors", function(errors) {
		status.text("App updated with errors. No reload!");
		okness.text("Errors while compiling.");
		$errors.text("\n" + stripAnsi(errors.join("\n\n\n")) + "\n\n");
		header.css({borderColor: "#ebcb8b"});
		$errors.show(); iframe.hide();
	});

	io.on("proxy-error", function(errors) {
		status.text("Could not proxy to content base target!");
		okness.text("Proxy error.");
		$errors.text("\n" + stripAnsi(errors.join("\n\n\n")) + "\n\n");
		header.css({borderColor: "#ebcb8b"});
		$errors.show(); iframe.hide();
	});

	io.on("disconnect", function() {
		status.text("");
		okness.text("Disconnected.");
		$errors.text("\n\n\n  Lost connection to webpack-dev-server.\n  Please restart the server to reestablish connection...\n\n\n\n");
		header.css({borderColor: "#ebcb8b"});
		$errors.show(); iframe.hide();
	});

	iframe.load(function() {
		status.text("App ready.");
		header.css({borderColor: ""});
		iframe.show();
	});

	function reloadApp() {
		if(hot) {
			status.text("App hot update.");
			try {
				iframe[0].contentWindow.postMessage("webpackHotUpdate" + currentHash, "*");
			} catch(e) {
				console.warn(e);
			}
			iframe.show();
		} else {
			status.text("App updated. Reloading app...");
			header.css({borderColor: "#96b5b4"});
			try {
				var old = iframe[0].contentWindow.location + "";
				if(old.indexOf("about") == 0) old = null;
				iframe.attr("src", old || (contentPage + window.location.hash));
				old && iframe[0].contentWindow.location.reload();
			} catch(e) {
				iframe.attr("src", contentPage + window.location.hash);
			}
		}
	}

});
