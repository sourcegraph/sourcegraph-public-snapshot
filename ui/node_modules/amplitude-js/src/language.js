var getLanguage = function() {
    return (navigator && ((navigator.languages && navigator.languages[0]) ||
        navigator.language || navigator.userLanguage)) || undefined;
};

module.exports = {
    language: getLanguage()
};
