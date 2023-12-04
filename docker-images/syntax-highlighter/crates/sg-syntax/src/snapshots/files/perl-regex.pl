package URI;

use strict;
use vars qw($VERSION);

sub new
{
    my($class, $uri, $scheme) = @_;

    $uri = defined ($uri) ? "$uri" : "";
    $uri =~ s/^<(?:URL:)?(.*)>$/$1/;  #
    $uri =~ s/^"(.*)"$/$1/;
    $uri =~ s/^\s+//;
    $uri =~ s/\s+$//;

    my $impclass;
    if ($uri =~ m/^($scheme_re):/so) {
        $scheme = $1;
    }
    else {
        if (($impclass = ref($scheme))) {
            $scheme = $scheme->scheme;
        }
        elsif ($scheme && $scheme =~ m/^($scheme_re)(?::|$)/o) {
            $scheme = $1;
        }
    }
    $impclass ||= implementor($scheme) ||
    do {
        require URI::_foreign;
        $impclass = 'URI::_foreign';
    };

    return $impclass->_init($uri, $scheme);
}
