#include <string.h>
#include <stdio.h>
#include <unistd.h>
#include <sys/wait.h>
#include <stdarg.h>
#include <fcntl.h>

/* This is just a wrapper with low memory footprint to parse and match when
   to run twitter command line program to tweet.

   52K in ram with musl-gcc -static

   The twitter client used here is twty.
   Requires binary ~/bin/twty present and properly configured
   see https://github.com/mattn/twty
 */

#define TWTY "/home/fridim/bin/twty"

void say(const char *channel, const char *fmt, ...) {
  char content[256];

  va_list argptr;
  va_start(argptr, fmt);
  int n = vsnprintf(content, 255, fmt, argptr);
  va_end(argptr);
  if (n <= 0 || n >= 255) {
    return ;
  }

  printf("PRIVMSG %s :%s\n", channel, content);
  fflush(stdout);
}

int main() {
  size_t len = 0;
  ssize_t i;
  char * line = NULL;

  while ((i = getline(&line, &len, stdin)) != -1) {
    if (strstr(line, "PRIVMSG") == NULL) {
      continue;
    }
    if (strstr(line, ":tweet ") == NULL) {
      continue;
    }

    if (strlen(line) < 1) { continue; }

    char who[strlen(line)];
    char channel[strlen(line)];
    char content[strlen(line)];

    // We cannot use system() because of Shell injection in the string, so let's fork
    // create pipe and use execl.

    if ((sscanf(line, "%s PRIVMSG %s ::tweet %[^\n\r]", who, channel, content)) == 3) {
      int out[2], err[2];
      if (pipe(out) == -1 || pipe(err) == -1) {
        say(channel, "FAIL pipe()");
        continue;
      }

      if (fork() == 0) { // CHILD
        // read ends not used
        close(out[0]);
        close(err[0]);

        if (dup2(out[1], STDOUT_FILENO) == -1) {
          fprintf(stderr, "FAIL dup2()\n");
          return 2;
        }
        if (dup2(err[1], STDERR_FILENO) == -1) {
          fprintf(stderr, "FAIL dup2()\n");
          return 2;
        }

        execl(TWTY, "twty", content, NULL);
        return 2;
      } else {
        int status;
        wait(&status);
        if (!WIFEXITED(status)) {
          say(channel, "twty failed, ret = %d", WEXITSTATUS(status));
          goto CLOSE;
        }
        // set err[0] fd non-blocking
        // since child has already exited we do not need to wait.
        int flags = fcntl(err[0], F_GETFL, 0);
        fcntl(err[0], F_SETFL, flags | O_NONBLOCK);

        char buf[256];
        int k = 0;
        while ((k = read(err[0], &buf, 255)) > 0) {
          buf[k] = '\0';
          fprintf(stderr, "%s", buf);
        }
        fflush(stderr);


        if ((k = read(out[0], &buf, 255)) == -1) {
          say(channel, "FAIL read pipe.");
          goto CLOSE;
        }
        buf[k] = '\0';

        char tweetid[128] = "";
        if ((sscanf(buf, "tweeted: %127s", tweetid) == 1)) {
          say(channel, "tweeted: https://twitter.com/statuses/%s", tweetid);
        } else {
          say(channel, "could not parse twty output.");
        }

      CLOSE:
        close(out[1]);
        close(err[1]);
        close(out[0]);
        close(err[0]);
      }
    }
  }

  return 0;
}
