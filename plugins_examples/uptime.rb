#!/usr/bin/ruby
STDOUT.sync = true

STDIN.each_line do |l|
  if l =~ /PRIVMSG (#\w+) ::uptime/
    channel = $1
    mess=%x[ps -hp  "#{Process.ppid}" -o etime]
    puts "PRIVMSG #{channel} :#{mess.strip}"
  end
end
