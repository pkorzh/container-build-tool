//go:build !remote


#define _GNU_SOURCE
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mount.h>
#include <sys/wait.h>
#include <unistd.h>

/* keep special_exit_code in sync with container_top_linux.go */
int special_exit_code = 255;
char **argv = NULL;

void
create_argv (int len)
{
  /* allocate one extra element because we need a final NULL in c */
  argv = malloc (sizeof (char *) * (len + 1));
  if (argv == NULL)
    {
      fprintf (stderr, "failed to allocate ps argv");
      exit (special_exit_code);
    }
  /* add final NULL */
  argv[len] = NULL;
}

void
set_argv (int pos, char *arg)
{
  argv[pos] = arg;
}

void
exec_ps ()
{
  if (argv == NULL)
    {
      fprintf (stderr, "argv not initialized");
      exit (special_exit_code);
    }
  execve (argv[0], argv, NULL);
  fprintf (stderr, "execve: %m");
  exit (special_exit_code);
}
