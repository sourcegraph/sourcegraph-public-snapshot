// see http://ejohn.org/blog/jquery-livesearch/ 

jQuery.fn.liveUpdate = function(list){
	list = jQuery(list);

  	if ( list.length ) {
		var rows = list.children('li'),
			cache = rows.map(function(){
				//return this.innerHTML.toLowerCase(); // remove toLC if score uses case
				return text_plus_space(jQuery(this))/*.toLowerCase()*/;
			});
			
		this
			.keyup(filter).keyup()
			.parents('form').submit(function(){
				return false;
			});
	}
		
	return this;
	
	// skip html tags and non-word chars
	function text_plus_space(node) {
		var ret = "";
		
		jQuery.each(node, function() {
			jQuery.each(this.childNodes, function() {
				if (this.nodeType != 8 )
				ret += this.nodeType != 1 ?
					this.nodeValue :
					text_plus_space([this]);
			});
		});
		
		ret = ret.replace(/\W/g, " ");
		return ret + " ";
	}
  
	function filter(){
		var term = jQuery.trim( jQuery(this).val()/*.toLowerCase()*/ ), scores = [];  // remove toLC if score uses case
		
		if ( !term ) {
			rows.show();
		} else {
			rows.hide();

			cache.each(function(i){
				var score = this.score(term);
				//score 0==no match 1==exact match (across whole term!)
				if (score > .50 /*.70*/) { scores.push([score, i]); }
			});

			jQuery.each(scores.sort(function(a, b){return b[0] - a[0];}), function(){
				jQuery(rows[ this[1] ]).show();
			});
		}
	}
};
