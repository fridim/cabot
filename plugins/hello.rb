#!/usr/bin/ruby
STDOUT.sync = true

STDIN.each_line do |l|
  if l =~ /PRIVMSG (#\w+) :(.+)/
    channel = $1
    message = $2
    if message =~ /^hello/i
      puts "PRIVMSG #{channel} :Hello o/"
    end
  end
end
