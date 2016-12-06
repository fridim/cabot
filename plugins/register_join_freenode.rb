#!/usr/bin/ruby

STDOUT.sync = true
nick="mybotnickname"
name="golang-bot"
channels=['#channeltojoin']

# Adapt to your needs. This will work on freenode.
STDIN.each_line do |l| 
  if l =~ /^:[^ ]+ NOTICE \* :\*\*\* Found your hostname/
    puts "USER #{name} 0 * :#{name}"
    puts "NICK #{nick}\n"
    if File.exists? 'password'
      puts "PRIVMSG Nickserv :identify #{File.read('password').chomp}"
    end
    channels.each { |c| puts "JOIN #{c}" }
  end
end
