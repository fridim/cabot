#!/usr/bin/ruby
# This plugin WHOIS every joining user, checks if user is
# connected to freenode with a secure connection and ask user
# to do so if he doesn't.

STDOUT.sync = true

def secure?(nickname)
  puts "WHOIS #{nickname}"
  STDIN.each_line do |l|
    if l =~ /:[^ ]+ (\d+)/
      cr = $1
      case cr
      when "318" # RPL_ENDOFWHOIS
        return false
      when "671" # Freenode secure connected user
        return true
      end
    end
  end
end

STDIN.each_line do |l|
  if l =~ /^:([^! ]+)![^ ]+ JOIN (#[^ ]+)/
    nickname = $1
    channel = $2.chomp()
    if ! secure?(nickname)
      puts "PRIVMSG #{channel} :#{nickname}: hey, could you connect using SSL please?"
    end
  end
end
