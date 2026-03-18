#include "openvms.h"
#include "dat.h"
#include "fns.h"

size_t
getsysname(char *buf, size_t sz)
{
	char sysname[16];
	size_t written;
	int status;
	ItemList64 sysname_desc = {1, SYI$_NODENAME, -1, 15, (uint64_t)&sysname, (uint64_t)&written};

	if(sz < 15)
		panic("sysname buffer too small");
	memset(&sysname, 0, 16);
	status = sys$getsyiw(0, 0, 0, &sysname_desc, 0, 0, 0);
	if(!($VMS_STATUS_SUCCESS(status)))
		lib$stop(status);
	strcpy(buf, (char*)&sysname);
	return written;
}
