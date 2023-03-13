package HTTP::Config;

use strict;
use URI;
use vars qw($VERSION);

$VERSION = "5.835";

sub new {
    my $class = shift;
    return bless [], $class;
}

sub entries {
    my $self = shift;
    @$self;
}

sub empty {
    my $self = shift;
    not @$self;
}

sub add {
    if (@_ == 2) {
        my $self = shift;
        push(@$self, shift);
        return;
    }
    my($self, %spec) = @_;
    push(@$self, \%spec);
    return;
}

sub find2 {
    my($self, %spec) = @_;
    my @found;
    my @rest;
 ITEM:
    for my $item (@$self) {
        for my $k (keys %spec) {
            if (!exists $item->{$k} || $spec{$k} ne $item->{$k}) {
                push(@rest, $item);
                next ITEM;
            }
        }
        push(@found, $item);
    }
    return \@found unless wantarray;
    return \@found, \@rest;
}

sub find {
    my $self = shift;
    my $f = $self->find2(@_);
    return @$f if wantarray;
    return $f->[0];
}

sub remove {
    my($self, %spec) = @_;
    my($removed, $rest) = $self->find2(%spec);
    @$self = @$rest if @$removed;
    return @$removed;
}

my %MATCH = (
    m_scheme => sub {
        my($v, $uri) = @_;
        return $uri->_scheme eq $v;  # URI known to be canonical
    },
    m_secure => sub {
        my($v, $uri) = @_;
        my $secure = $uri->can("secure") ? $uri->secure : $uri->_scheme eq "https";
        return $secure == !!$v;
    },
    m_host_port => sub {
        my($v, $uri) = @_;
        return unless $uri->can("host_port");
        return $uri->host_port eq $v, 7;
    },
    m_host => sub {
        my($v, $uri) = @_;
        return unless $uri->can("host");
        return $uri->host eq $v, 6;
    },
    m_port => sub {
        my($v, $uri) = @_;
        return unless $uri->can("port");
        return $uri->port eq $v;
    },
    m_domain => sub {
        my($v, $uri) = @_;
        return unless $uri->can("host");
        my $h = $uri->host;
        $h = "$h.local" unless $h =~ /\./;
        $v = ".$v" unless $v =~ /^\./;
        return length($v), 5 if substr($h, -length($v)) eq $v;
        return 0;
    },
    m_path => sub {
        my($v, $uri) = @_;
        return unless $uri->can("path");
        return $uri->path eq $v, 4;
    },
    m_path_prefix => sub {
        my($v, $uri) = @_;
        return unless $uri->can("path");
        my $path = $uri->path;
        my $len = length($v);
        return $len, 3 if $path eq $v;
        return 0 if length($path) <= $len;
        $v .= "/" unless $v =~ m,/\z,,;
        return $len, 3 if substr($path, 0, length($v)) eq $v;
        return 0;
    },
    m_path_match => sub {
        my($v, $uri) = @_;
        return unless $uri->can("path");
        return $uri->path =~ $v;
    },
    m_uri__ => sub {
        my($v, $k, $uri) = @_;
        return unless $uri->can($k);
        return 1 unless defined $v;
        return $uri->$k eq $v;
    },
    m_method => sub {
        my($v, $uri, $request) = @_;
        return $request && $request->method eq $v;
    },
    m_proxy => sub {
        my($v, $uri, $request) = @_;
        return $request && ($request->{proxy} || "") eq $v;
    },
    m_code => sub {
        my($v, $uri, $request, $response) = @_;
        $v =~ s/xx\z//;
        return unless $response;
        return length($v), 2 if substr($response->code, 0, length($v)) eq $v;
    },
    m_media_type => sub {  # for request too??
        my($v, $uri, $request, $response) = @_;
        return unless $response;
        return 1, 1 if $v eq "*/*";
        my $ct = $response->content_type;
        return 2, 1 if $v =~ s,/\*\z,, && $ct =~ m,^\Q$v\E/,;
        return 3, 1 if $v eq "html" && $response->content_is_html;
        return 4, 1 if $v eq "xhtml" && $response->content_is_xhtml;
        return 10, 1 if $v eq $ct;
        return 0;
    },
    m_header__ => sub {
        my($v, $k, $uri, $request, $response) = @_;
        return unless $request;
        return 1 if $request->header($k) eq $v;
        return 1 if $response && $response->header($k) eq $v;
        return 0;
    },
    m_response_attr__ => sub {
        my($v, $k, $uri, $request, $response) = @_;
        return unless $response;
        return 1 if !defined($v) && exists $response->{$k};
        return 0 unless exists $response->{$k};
        return 1 if $response->{$k} eq $v;
        return 0;
    },
);

sub matching {
    my $self = shift;
    if (@_ == 1) {
        if ($_[0]->can("request")) {
            unshift(@_, $_[0]->request);
            unshift(@_, undef) unless defined $_[0];
        }
        unshift(@_, $_[0]->uri_canonical) if $_[0] && $_[0]->can("uri_canonical");
    }
    my($uri, $request, $response) = @_;
    $uri = URI->new($uri) unless ref($uri);

    my @m;
 ITEM:
    for my $item (@$self) {
        my $order;
        for my $ikey (keys %$item) {
            my $mkey = $ikey;
            my $k;
            $k = $1 if $mkey =~ s/__(.*)/__/;
            if (my $m = $MATCH{$mkey}) {
                #print "$ikey $mkey\n";
                my($c, $o);
                my @arg = (
                    defined($k) ? $k : (),
                    $uri, $request, $response
                );
                my $v = $item->{$ikey};
                $v = [$v] unless ref($v) eq "ARRAY";
                for (@$v) {
                    ($c, $o) = $m->($_, @arg);
                    #print "  - $_ ==> $c $o\n";
                    last if $c;
                }
                next ITEM unless $c;
                $order->[$o || 0] += $c;
            }
        }
        $order->[7] ||= 0;
        $item->{_order} = join(".", reverse map sprintf("%03d", $_ || 0), @$order);
        push(@m, $item);
    }
    @m = sort { $b->{_order} cmp $a->{_order} } @m;
    delete $_->{_order} for @m;
    return @m if wantarray;
    return $m[0];
}

sub add_item {
    my $self = shift;
    my $item = shift;
    return $self->add(item => $item, @_);
}

sub remove_items {
    my $self = shift;
    return map $_->{item}, $self->remove(@_);
}

sub matching_items {
    my $self = shift;
    return map $_->{item}, $self->matching(@_);
}

1;

__END__

=head1 NAME

HTTP::Config - Configuration for request and response objects

=head1 SYNOPSIS

 use HTTP::Config;
 my $c = HTTP::Config->new;
 $c->add(m_domain => ".example.com", m_scheme => "http", verbose => 1);

 use HTTP::Request;
 my $request = HTTP::Request->new(GET => "http://www.example.com");

 if (my @m = $c->matching($request)) {
    print "Yadayada\n" if $m[0]->{verbose};
 }

=head1 DESCRIPTION

An C<HTTP::Config> object is a list of entries that
can be matched against request or request/response pairs.  Its
purpose is to hold configuration data that can be looked up given a
request or response object.

Each configuration entry is a hash.  Some keys specify matching to
occur against attributes of request/response objects.  Other keys can
be used to hold user data.

The following methods are provided:

=over 4

=item $conf = HTTP::Config->new

Constructs a new empty C<HTTP::Config> object and returns it.

=item $conf->entries

Returns the list of entries in the configuration object.
In scalar context returns the number of entries.

=item $conf->empty

Return true if there are no entries in the configuration object.
This is just a shorthand for C<< not $conf->entries >>.

=item $conf->add( %matchspec, %other )

=item $conf->add( \%entry )

Adds a new entry to the configuration.
You can either pass separate key/value pairs or a hash reference.

=item $conf->remove( %spec )

Removes (and returns) the entries that have matches for all the key/value pairs in %spec.
If %spec is empty this will match all entries; so it will empty the configuation object.

=item $conf->matching( $uri, $request, $response )

=item $conf->matching( $uri )

=item $conf->matching( $request )

=item $conf->matching( $response )

Returns the entries that match the given $uri, $request and $response triplet.

If called with a single $request object then the $uri is obtained by calling its 'uri_canonical' method.
If called with a single $response object, then the request object is obtained by calling its 'request' method;
and then the $uri is obtained as if a single $request was provided.

The entries are returned with the most specific matches first.
In scalar context returns the most specific match or C<undef> in none match.

=item $conf->add_item( $item, %matchspec )

=item $conf->remove_items( %spec )

=item $conf->matching_items( $uri, $request, $response )

Wrappers that hides the entries themselves.

=back

=head2 Matching

The following keys on a configuration entry specify matching.  For all
of these you can provide an array of values instead of a single value.
The entry matches if at least one of the values in the array matches.

Entries that require match against a response object attribute will never match
unless a response object was provided.

=over

=item m_scheme => $scheme

Matches if the URI uses the specified scheme; e.g. "http".

=item m_secure => $bool

If $bool is TRUE; matches if the URI uses a secure scheme.  If $bool
is FALSE; matches if the URI does not use a secure scheme.  An example
of a secure scheme is "https".

=item m_host_port => "$hostname:$port"

Matches if the URI's host_port method return the specified value.

=item m_host => $hostname

Matches if the URI's host method returns the specified value.

=item m_port => $port

Matches if the URI's port method returns the specified value.

=item m_domain => ".$domain"

Matches if the URI's host method return a value that within the given
domain.  The hostname "www.example.com" will for instance match the
domain ".com".

=item m_path => $path

Matches if the URI's path method returns the specified value.

=item m_path_prefix => $path

Matches if the URI's path is the specified path or has the specified
path as prefix.

=item m_path_match => $Regexp

Matches if the regular expression matches the URI's path.  Eg. qr/\.html$/.

=item m_method => $method

Matches if the request method matches the specified value. Eg. "GET" or "POST".

=item m_code => $digit

=item m_code => $status_code

Matches if the response status code matches.  If a single digit is
specified; matches for all response status codes beginning with that digit.

=item m_proxy => $url

Matches if the request is to be sent to the given Proxy server.

=item m_media_type => "*/*"

=item m_media_type => "text/*"

=item m_media_type => "html"

=item m_media_type => "xhtml"

=item m_media_type => "text/html"

Matches if the response media type matches.

With a value of "html" matches if $response->content_is_html returns TRUE.
With a value of "xhtml" matches if $response->content_is_xhtml returns TRUE.

=item m_uri__I<$method> => undef

Matches if the URI object provides the method.

=item m_uri__I<$method> => $string

Matches if the URI's $method method returns the given value.

=item m_header__I<$field> => $string

Matches if either the request or the response have a header $field with the given value.

=item m_response_attr__I<$key> => undef

=item m_response_attr__I<$key> => $string

Matches if the response object has that key, or the entry has the given value.

=back

=head1 SEE ALSO

L<URI>, L<HTTP::Request>, L<HTTP::Response>

=head1 COPYRIGHT

Copyright 2008, Gisle Aas

This library is free software; you can redistribute it and/or
modify it under the same terms as Perl itself.

=cut
