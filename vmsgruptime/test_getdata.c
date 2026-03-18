#include "openvms.h"
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
	printf("\tlifetime = %llu\n", uptime->lifetime);
	printf("};\n");
}

int
main(int argc, char *argv[])
{
	Uptime *uptime;

	uptime = getdata();
	print_uptime("uptime", uptime);
	return 0;
}
