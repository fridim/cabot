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

def wait_for(regex, max=20)
  count = 0
  STDIN.each_line do |l|
    count = count + 1
    return false if count >= max

    if l =~ regex
      return true
    end
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
      wait_for(/^:[^ ]+ CAP #{nick} ACK :sasl/) or break
      puts "AUTHENTICATE PLAIN"
      wait_for(/^AUTHENTICATE \+/) or break
      sasl_password_lines.each do |pl|
        puts "AUTHENTICATE #{pl}"
      end

      wait_for(/^:[^ ]+ 903 #{nick} :SASL authentication successful/) or break
      puts "CAP END"

      # Join channels
      channels.each { |c| puts "JOIN #{c}" }
    else
      puts "USER #{name} 0 * :#{name}"
      puts "NICK #{nick}"
      channels.each { |c| puts "JOIN #{c}" }
    end
  end
end
