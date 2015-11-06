exports.translate = function(tree) {
    return new CSSOTranslator().translate(tree);
};

exports.translator = function() {
    return new CSSOTranslator();
};
