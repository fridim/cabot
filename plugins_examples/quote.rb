#!/usr/bin/env ruby

STDOUT.sync = true

$db = "quotes.txt"
if !File.exists? $db
  system("touch #{$db}")
end

def getquote()
  return %x{shuf -n 1 #{$db}}
end

def setquote(quote)
  begin
    open($db, 'a') do |f| 
      f.puts quote
    end
    return true
  rescue
    return false
  end
end

STDIN.each_line do |l| 
  if l =~ /^:[^ ]+ PRIVMSG (#[^ ]+) :(:.*)\r\n$/
    channel = $1
    what = $2

    if what =~ /^:setquote (.*)/
      if setquote($1)
        puts "PRIVMSG #{channel} :quote saved"
      end

    elsif what =~ /^:(get)?quote/
      quote =  getquote()
      puts "PRIVMSG #{channel} :#{quote}" if quote.length > 0
    end
  end
end
