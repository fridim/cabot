#!/usr/bin/perl -w

use strict;
use LWP::Simple;
# autoflush
$|++;
while (<>) {
    if (/:[^ ]+ PRIVMSG (#[^ ]+) ::insult ?([^ ]+)?/) {
        local $/ = "\r\n";
        my $channel = $1;
        chomp(my $who = $2);
        $who = "${who}: " if $who;
        my $content = get("http://www.randominsults.net/");
        if ($content =~ /<strong><i>(.*?)<\/i><\/strong>/) {
            my $strip = $1;
            $strip =~ s/\n/ /g;
            if ($strip) {
                print "PRIVMSG ${channel} :${who}${strip}\n";
            }
        } else {
            print "PRIVMSG ${channel} :yo momma!\n";
        }
    }
}
