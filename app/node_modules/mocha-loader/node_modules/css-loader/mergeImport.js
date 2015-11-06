module.exports = function(list, importedList, media) {
	for(var i = 0; i < importedList.length; i++) {
		var item = importedList[i];
		if(media && !item[2])
			item[2] = media;
		else if(media) {
			item[2] = "(" + item[2] + ") and (" + media + ")";
		}
		list.push(item);
	}
};