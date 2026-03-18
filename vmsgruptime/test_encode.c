#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include "dat.h"
#include "fns.h"

void
print_uptime(char *name, Uptime *uptime)
{
	printf("%s = Uptime(%p){\n", name, uptime);
	printf("\thostname = \"%s\"\n", uptime->hostname);
	printf("\tostype = %s\n", byte2os(uptime->ostype));
	printf("\tuptime = %llu\n", uptime->uptime);
	printf("\tload1 = %f\n", uptime->load1);
	printf("\tload5 = %f\n", uptime->load5);
	printf("\tload15 = %f\n", uptime->load15);
	printf("\tnusers = %llu\n", uptime->nusers);
	printf("};\n");
}

int
main(int argc, char *argv[])
{
	char *testhost = "thisisatest";
	Uptime in_uptime;
	char *uptime_bytes;
	Uptime *out_uptime;

	in_uptime.hostname = testhost;
	in_uptime.ostype = OPENVMS;
	in_uptime.uptime = 1234567890;
	in_uptime.load1 = 1.23;
	in_uptime.load5 = 4.56;
	in_uptime.load15 = 7.89;
	in_uptime.nusers = 15;

	uptime_bytes = encodemessage(&in_uptime);
	out_uptime = decodemessage(uptime_bytes);

	print_uptime("in_uptime", &in_uptime);
	print_uptime("out_uptime", out_uptime);

	return 0;
}
