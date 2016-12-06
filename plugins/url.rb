#!/usr/bin/ruby
require 'net/http'
require 'uri'


STDOUT.sync = true

$service = "https://is.gd/create.php?format=simple&url="


def open(url)
  Net::HTTP.get(URI.parse(url))
end


STDIN.each_line do |l| 
  if l =~ /PRIVMSG (#\w+) :(.+)/
    channel = $1
    what = $2
    # extract URL
    urls = URI.extract(what, /https?/)
    results = []
    for url in urls
      next if url.length < 110
      results << open("#{$service}#{URI.escape url}")
    end
    puts "PRIVMSG #{channel} :#{results.join(' ')}" unless results.empty? 
  end
end
