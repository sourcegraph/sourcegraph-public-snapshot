#!/usr/bin/perl

use Switch;

use XML::Parser;
use Data::Dumper;

# TODO: save to some strcuture, output that in one go
# output is only correct if the XML is in the correct form.

$PREF = "% ";

sub ext { return 'perl-unhandled-extern-ref' }

sub start {
    shift;
    $e = shift;
    switch ($e) {
        case "title" {
            for ($i = 0; $i < scalar @_ - 1; $i++) {
                print $PREF . $_[$i] . " = " . '"' . $_[$i+1] . '"' . "\n";
            }
            print "${PREF}title = \"";
        }
        case "rfc" {
            for ($i = 0; $i < scalar @_ - 1; $i++) {
                print $PREF . $_[$i] . " = " . '"' . $_[$i+1] . '"' . "\n";
                $i++;
            }
        }
        case "author" {
            print "$PREF\n${PREF}[[author]]\n";
            for ($i = 0; $i < scalar @_ - 1; $i++) {
                print $PREF . $_[$i] . " = " . '"' . $_[$i+1] . '"' . "\n";
                $i++;
            }
        }
        case "address"      { print $PREF . "[author.address]\n" }
        case "postal"       { print $PREF . "[author.address.postal]\n" }
        # todo street code country
        # todo date
        case "keyword"      { print $PREF . "keyword = [" }
        case "area"         { print $PREF . "area = \"" }
        case "workgroup"    { print $PREF . "workgroup = \"" }
        case "organization" { print $PREF . "organization = \"" }
        case "email"        { print $PREF . "email = \"" }
        case "uri"          { print $PREF . "uri = \"" }
        case "phone"        { print $PREF . "phone = \"" }
    }
}

sub end {
    shift;
    $e = shift;
    switch ($e) {
        # rfc ipr="trust200902" category="exp" docName="draft-gieben-nsec4-02">
        case "rfc"          { }
        case "address"      { }
        case "keyword"      { print "]\n"; }
#        case "author"       { print "$PREF\n"; }
        case "email"        { print "\"\n"; }
        case "uri"          { print "\"\n"; }
        case "phone"        { print "\"\n"; }
        case "title"        { print "\"\n"; }
        case "area"         { print "\"\n"; }
        case "organization" { print "\"\n"; }
        case "workgroup"    { print "\"\n"; }
        case "front"        { exit; }
    }
}

sub char {
    $p = shift;
    switch ($p->current_element) {
        case "keyword"        { print join ' ', map qq("$_"), @_; }
        case "title"          { print $_[0] ; }
        case "area"           { print $_[0] ; }
        case "workgroup"      { print $_[0] ; }
        case "organization"   { print $_[0] ; }
        case "email"          { print $_[0] ; }
        case "uri"            { print $_[0] ; }
        case "phone"          { print $_[0] ; }
    }
}

my $xmlfile = shift @ARGV;
my $parser = XML::Parser->new( Style => 'Tree', ErrorContext => 2, NoExpand => 1 );
$parser->setHandlers(ExternEnt => \&ext, Start => \&start, End => \&end, Char => \&char);
eval { $dom = $parser->parsefile( $xmlfile ); };
