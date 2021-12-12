#!/usr/bin/ruby
require "base64"

STDOUT.sync = true
STDERR.sync = true
nick="cabot"
name="golang-bot"
channels=['#mychannel']

if File.exists? 'password'
  sasl_password = Base64.strict_encode64(nick + "\0" + nick + "\0" + File.read('password').chomp)
end

sasl_password_lines = []
maxlen=400
while
  if sasl_password.length > maxlen
    sasl_password_lines.push sasl_password[0..maxlen]
    sasl_password = sasl_password[maxlen..]
  else
    sasl_password_lines.push sasl_password
    break
  end
end

# Adapt to your needs. This will work on freenode.
STDIN.each_line do |l|
  if l =~ /^:[^ ]+ NOTICE \* :\*\*\* (Found|Couldn't look up) your hostname/
    if File.exists? 'password'
      puts "CAP LS"
      puts "USER #{name} 0 * :#{name}"
      puts "NICK #{nick}"
      puts "CAP REQ :sasl"
    else
      puts "USER #{name} 0 * :#{name}"
      puts "NICK #{nick}"
    end
  end

  if l =~ /^:[^ ]+ CAP #{nick} ACK :sasl/
    # use SASL
    if File.exists? 'password'
      puts "AUTHENTICATE PLAIN"
    end
  end

  if l =~ /^AUTHENTICATE *\+/
    sasl_password_lines.each do |pl|
      puts "AUTHENTICATE #{pl}"
    end
  end

  if l =~ /^:[^ ]+ 903 #{nick} :SASL authentication successful/
    puts "CAP END"
    channels.each { |c| puts "JOIN #{c}" }
  end
end
