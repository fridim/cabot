#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys
import re
from googletrans import Translator
translator = Translator()

line = sys.stdin.readline()
while line:
    match = re.search('^:([^\s]+) PRIVMSG (#[^\s]+) :(.+)', line)
    if not match:
        line = sys.stdin.readline()
        continue

    who = match.group(1)
    chan = match.group(2)
    what = match.group(3).strip().strip('\r\n')

    def reply(text):
        print("PRIVMSG %s :%s" % (chan, text))
        sys.stdout.flush()

    if what[:10] == ':translate':
        m2 = re.search('^:translate (.*)', what)
        if not m2:
            line = sys.stdin.readline()
            continue
        try:
            reply(translator.translate(m2.group(1), dest='fr').text)
        except:
            reply('Oups!')
    elif what[:4] == ':tr ':
        m2 = re.search('^:tr (\w+\-?\w+) (\w+\-?\w+) (.+)', what)
        if not m2:
            line = sys.stdin.readline()
            continue
        try:
            reply(translator.translate(m2.group(3), src=m2.group(1), dest=m2.group(2)).text)
        except:
            reply('Oups!')
    line = sys.stdin.readline()
