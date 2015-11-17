var DEFAULT_LANGS = ['ru', 'en'];

exports.techMixin = {

    getLangs: function() {
        var env = process.env.BEM_I18N_LANGS;
        return env? env.split(' ') : [].concat(DEFAULT_LANGS);
    },

    getSuffixes: function() {

        return this.getLangs()
            .map(function(lang) {
                return lang + '.md';
            });

    }

};
