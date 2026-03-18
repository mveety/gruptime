#include "openvms.h"
#include "dat.h"
#include "fns.h"

int
main(int argc, char *argv[])
{
	char hostname[18];
	size_t written;

	memset(&hostname, 0, sizeof(hostname));
	written = getsysname(hostname, sizeof(hostname));
	printf("hostname = \"%s\"\n", hostname);
	printf("length(hostname) = %lu\n", written);
	return 0;
}
