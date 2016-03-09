const requestAnimFrame = typeof window !== "undefined" ? window.requestAnimationFrame || window.webkitRequestAnimationFrame || window.mozRequestAnimationFrame || function(callback) { window.setTimeout(callback, 1000/60); } : function() {};

function easeInOutQuad(t, b, c, d) {
	t /= d/2;
	if (t < 1) return c/2*t*t + b;
	t--;
	return -c/2 * (t*(t-2) - 1) + b;
}

export default function(element, to, duration, callback) {
	let start = element.scrollTop,
		change = to - start,
		animationStart = Number(new Date());
	let animating = true;
	let lastpos = null;

	let animateScroll = function() {
		if (!animating) {
			return;
		}
		requestAnimFrame(animateScroll);
		let now = Number(new Date());
		let val = Math.floor(easeInOutQuad(now - animationStart, start, change, duration));
		if (lastpos) {
			let top = Math.floor(element.scrollTop);
			if (lastpos === top || lastpos - 1 === top) {
				lastpos = val;
				element.scrollTop = val;
			} else {
				animating = false;
			}
		} else {
			lastpos = val;
			element.scrollTop = val;
		}
		if (now > animationStart + duration) {
			element.scrollTop = to;
			animating = false;
			if (callback) { callback(); }
		}
	};
	requestAnimFrame(animateScroll);
}
