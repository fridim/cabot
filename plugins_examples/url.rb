#!/usr/bin/ruby
require 'net/http'

STDOUT.sync = true

$service = "https://is.gd/create.php?format=simple&url="

STDIN.each_line do |l|
  if l =~ /PRIVMSG (#\w+) :(.+)/
    channel = $1
    what = $2
    # extract URL
    urls = URI.extract(what, /https?/)
    results = []
    for url in urls
      next if url.length < 110
      uri = URI.parse("#{$service}#{URI.escape url}")
      http = Net::HTTP.new(uri.host, uri.port)
      http.use_ssl = true
      results << http.get(uri.request_uri).body
    end
    puts "PRIVMSG #{channel} :#{results.join(' ')}" unless results.empty?
  end
end
